package worker

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/event"
	"github.com/ICE-awa/renice-sl/shared/mq"
	natsserver "github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats.go"
)

// --- Mock DLQ Repository ---

type mockDLQRepo struct {
	mu                      sync.Mutex
	recordedMessages        []*dtov1.DLQMessage
	safetyStatusUnknownCode []string
}

func (m *mockDLQRepo) RecordDLQMessage(_ context.Context, msg *dtov1.DLQMessage) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.recordedMessages = append(m.recordedMessages, msg)
	return nil
}

func (m *mockDLQRepo) SetSafetyStatusUnknown(_ context.Context, code string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.safetyStatusUnknownCode = append(m.safetyStatusUnknownCode, code)
	return nil
}

func (m *mockDLQRepo) GetDLQMessages(context.Context, *dtov1.GetDLQMessagesReq) (*dtov1.GetDLQMessagesResp, error) {
	return nil, nil
}
func (m *mockDLQRepo) SetDLQMessageRetrying(context.Context, int64) (*dtov1.RetryDLQMessageData, error) {
	return nil, nil
}
func (m *mockDLQRepo) MarkAsResolved(_ context.Context, _ int64) (string, error) {
	return "", nil
}

// --- Test NATS Server Helper ---

func startTestNATSServer(t *testing.T) (*natsserver.Server, *mq.NatsClient) {
	t.Helper()

	opts := &natsserver.Options{
		Host:      "127.0.0.1",
		Port:      -1, // random available port
		JetStream: true,
		StoreDir:  t.TempDir(),
		NoLog:     true,
		NoSigs:    true,
	}

	s, err := natsserver.NewServer(opts)
	if err != nil {
		t.Fatalf("failed to create NATS server: %v", err)
	}
	s.Start()

	if !s.ReadyForConnections(5 * time.Second) {
		t.Fatal("NATS server not ready")
	}

	nc, err := nats.Connect(s.ClientURL())
	if err != nil {
		t.Fatalf("failed to connect to NATS: %v", err)
	}

	js, err := nc.JetStream()
	if err != nil {
		t.Fatalf("failed to get JetStream context: %v", err)
	}

	client := &mq.NatsClient{
		Conn:      nc,
		JetStream: js,
	}

	t.Cleanup(func() {
		nc.Close()
		s.Shutdown()
	})

	return s, client
}

// --- Tests ---

// 完整流程：advisory → 取原消息 → 发布 DLQ → 记录 DB
func TestDLQWorker_ProcessAdvisory_GenericSubject(t *testing.T) {
	_, client := startTestNATSServer(t)

	// 创建 Stream 用于存放原始消息
	_, err := client.JetStream.AddStream(&nats.StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"link.>"},
	})
	if err != nil {
		t.Fatalf("failed to add stream: %v", err)
	}

	// 创建 DLQ Stream 以接收死信
	_, err = client.JetStream.AddStream(&nats.StreamConfig{
		Name:     "DLQ",
		Subjects: []string{"dlq.>"},
	})
	if err != nil {
		t.Fatalf("failed to add DLQ stream: %v", err)
	}

	// 发布一条原始消息到 EVENTS stream
	clickPayload := json.RawMessage(`{"event_id":"evt-1","code":"abc123"}`)
	ack, err := client.JetStream.Publish(event.SubjectLinkClicked, clickPayload)
	if err != nil {
		t.Fatalf("failed to publish original msg: %v", err)
	}
	originalSeq := ack.Sequence

	// 启动 DLQ Worker
	repo := &mockDLQRepo{}
	worker := NewDLQWorker(client, repo)
	if err := worker.StartDLQWorker(); err != nil {
		t.Fatalf("failed to start DLQ worker: %v", err)
	}

	// 模拟 MaxDeliver Advisory 消息
	advisory := MaxDeliverAdvisory{
		Stream:    "EVENTS",
		Consumer:  "test-consumer",
		StreamSeq: originalSeq,
	}
	advData, _ := json.Marshal(advisory)
	err = client.Conn.Publish(
		"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.EVENTS.test-consumer",
		advData,
	)
	if err != nil {
		t.Fatalf("failed to publish advisory: %v", err)
	}
	client.Conn.Flush()

	// 等待处理完成
	time.Sleep(200 * time.Millisecond)

	// 验证 RecordDLQMessage 被调用
	repo.mu.Lock()
	defer repo.mu.Unlock()

	if len(repo.recordedMessages) != 1 {
		t.Fatalf("expected 1 recorded DLQ message, got %d", len(repo.recordedMessages))
	}

	msg := repo.recordedMessages[0]
	if msg.SourceStream != "EVENTS" {
		t.Errorf("expected stream 'EVENTS', got %q", msg.SourceStream)
	}
	if msg.SourceConsumer != "test-consumer" {
		t.Errorf("expected consumer 'test-consumer', got %q", msg.SourceConsumer)
	}
	if msg.StreamSeq != originalSeq {
		t.Errorf("expected seq %d, got %d", originalSeq, msg.StreamSeq)
	}
	if msg.Subject != event.SubjectLinkClicked {
		t.Errorf("expected subject %q, got %q", event.SubjectLinkClicked, msg.Subject)
	}
	if msg.Reason != "max_deliveries" {
		t.Errorf("expected reason 'max_deliveries', got %q", msg.Reason)
	}

	// link.clicked 不触发 SetSafetyStatusUnknown
	if len(repo.safetyStatusUnknownCode) != 0 {
		t.Errorf("SetSafetyStatusUnknown should NOT be called for link.clicked, got %v",
			repo.safetyStatusUnknownCode)
	}
}

// link.checked 触发 SetSafetyStatusUnknown
func TestDLQWorker_ProcessAdvisory_LinkChecked_SetsSafetyUnknown(t *testing.T) {
	_, client := startTestNATSServer(t)

	_, err := client.JetStream.AddStream(&nats.StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"link.>"},
	})
	if err != nil {
		t.Fatalf("failed to add stream: %v", err)
	}

	_, err = client.JetStream.AddStream(&nats.StreamConfig{
		Name:     "DLQ",
		Subjects: []string{"dlq.>"},
	})
	if err != nil {
		t.Fatalf("failed to add DLQ stream: %v", err)
	}

	// 发布一条 link.checked 消息
	checkReq := dtov1.CheckLinkReq{
		EventID:     "evt-check-1",
		Code:        "dangerous7",
		OriginalURL: "https://evil.example.com",
	}
	checkData, _ := json.Marshal(checkReq)
	ack, err := client.JetStream.Publish(event.SubjectLinkChecked, checkData)
	if err != nil {
		t.Fatalf("failed to publish link.checked msg: %v", err)
	}

	// 启动 Worker
	repo := &mockDLQRepo{}
	worker := NewDLQWorker(client, repo)
	if err := worker.StartDLQWorker(); err != nil {
		t.Fatalf("failed to start DLQ worker: %v", err)
	}

	// 发送 advisory
	advisory := MaxDeliverAdvisory{
		Stream:    "EVENTS",
		Consumer:  "check-consumer",
		StreamSeq: ack.Sequence,
	}
	advData, _ := json.Marshal(advisory)
	err = client.Conn.Publish(
		"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.EVENTS.check-consumer",
		advData,
	)
	if err != nil {
		t.Fatalf("failed to publish advisory: %v", err)
	}
	client.Conn.Flush()

	time.Sleep(200 * time.Millisecond)

	repo.mu.Lock()
	defer repo.mu.Unlock()

	// 验证 RecordDLQMessage
	if len(repo.recordedMessages) != 1 {
		t.Fatalf("expected 1 recorded message, got %d", len(repo.recordedMessages))
	}
	if repo.recordedMessages[0].Subject != event.SubjectLinkChecked {
		t.Errorf("expected subject %q, got %q",
			event.SubjectLinkChecked, repo.recordedMessages[0].Subject)
	}

	// 核心断言：link.checked 触发 SetSafetyStatusUnknown
	if len(repo.safetyStatusUnknownCode) != 1 {
		t.Fatalf("expected SetSafetyStatusUnknown called once, got %d calls",
			len(repo.safetyStatusUnknownCode))
	}
	if repo.safetyStatusUnknownCode[0] != "dangerous7" {
		t.Errorf("expected code 'dangerous7', got %q", repo.safetyStatusUnknownCode[0])
	}
}

// 无效 advisory JSON 不 panic，静默跳过
func TestDLQWorker_InvalidAdvisoryJSON(t *testing.T) {
	_, client := startTestNATSServer(t)

	repo := &mockDLQRepo{}
	worker := NewDLQWorker(client, repo)
	if err := worker.StartDLQWorker(); err != nil {
		t.Fatalf("failed to start DLQ worker: %v", err)
	}

	// 发送垃圾数据
	err := client.Conn.Publish(
		"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.FAKE.consumer1",
		[]byte("not valid json {{{"),
	)
	if err != nil {
		t.Fatalf("failed to publish invalid advisory: %v", err)
	}
	client.Conn.Flush()

	time.Sleep(100 * time.Millisecond)

	repo.mu.Lock()
	defer repo.mu.Unlock()

	// 不应有任何 DB 调用
	if len(repo.recordedMessages) != 0 {
		t.Errorf("expected 0 recorded messages for invalid JSON, got %d",
			len(repo.recordedMessages))
	}
}

// stream_seq 对应的消息不存在 → GetMsg 报错 → 静默跳过
func TestDLQWorker_GetMsgFails_NoRecord(t *testing.T) {
	_, client := startTestNATSServer(t)

	_, err := client.JetStream.AddStream(&nats.StreamConfig{
		Name:     "EVENTS",
		Subjects: []string{"link.>"},
	})
	if err != nil {
		t.Fatalf("failed to add stream: %v", err)
	}

	repo := &mockDLQRepo{}
	worker := NewDLQWorker(client, repo)
	if err := worker.StartDLQWorker(); err != nil {
		t.Fatalf("failed to start DLQ worker: %v", err)
	}

	// advisory 引用不存在的 stream_seq
	advisory := MaxDeliverAdvisory{
		Stream:    "EVENTS",
		Consumer:  "ghost-consumer",
		StreamSeq: 99999,
	}
	advData, _ := json.Marshal(advisory)
	err = client.Conn.Publish(
		"$JS.EVENT.ADVISORY.CONSUMER.MAX_DELIVERIES.EVENTS.ghost-consumer",
		advData,
	)
	if err != nil {
		t.Fatalf("failed to publish advisory: %v", err)
	}
	client.Conn.Flush()

	time.Sleep(100 * time.Millisecond)

	repo.mu.Lock()
	defer repo.mu.Unlock()

	if len(repo.recordedMessages) != 0 {
		t.Errorf("expected 0 recorded messages when GetMsg fails, got %d",
			len(repo.recordedMessages))
	}
}
