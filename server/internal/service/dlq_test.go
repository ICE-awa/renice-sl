package service

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
)

type mockDLQServiceRepo struct {
	getFn      func(context.Context, *dtov1.GetDLQMessagesReq) (*dtov1.GetDLQMessagesResp, error)
	retryFn    func(context.Context, int64) (*dtov1.RetryDLQMessageData, error)
	resolvedFn func(context.Context, int64) (string, error)

	retryCalledWith    int64
	resolvedCalledWith int64
}

func (m *mockDLQServiceRepo) RecordDLQMessage(context.Context, *dtov1.DLQMessage) error {
	return nil
}

func (m *mockDLQServiceRepo) GetDLQMessages(ctx context.Context, req *dtov1.GetDLQMessagesReq) (*dtov1.GetDLQMessagesResp, error) {
	if m.getFn != nil {
		return m.getFn(ctx, req)
	}
	return nil, errors.New("unexpected GetDLQMessages call")
}

func (m *mockDLQServiceRepo) SetDLQMessageRetrying(ctx context.Context, id int64) (*dtov1.RetryDLQMessageData, error) {
	m.retryCalledWith = id
	if m.retryFn != nil {
		return m.retryFn(ctx, id)
	}
	return &dtov1.RetryDLQMessageData{Subject: "link.clicked", Payload: json.RawMessage(`{"ok":true}`)}, nil
}

func (m *mockDLQServiceRepo) MarkAsResolved(ctx context.Context, id int64) (string, error) {
	m.resolvedCalledWith = id
	if m.resolvedFn != nil {
		return m.resolvedFn(ctx, id)
	}
	return "link.clicked", nil
}

func (m *mockDLQServiceRepo) SetSafetyStatusUnknown(context.Context, string) error {
	return nil
}

func TestDLQService_GetDLQMessagesDelegatesToRepo(t *testing.T) {
	t.Parallel()

	wantResp := &dtov1.GetDLQMessagesResp{Total: 1, PageNum: 2, PageSize: 10}
	repo := &mockDLQServiceRepo{
		getFn: func(_ context.Context, req *dtov1.GetDLQMessagesReq) (*dtov1.GetDLQMessagesResp, error) {
			if req.PageNum != 2 || req.PageSize != 10 {
				t.Fatalf("unexpected request: %#v", req)
			}
			return wantResp, nil
		},
	}
	svc := &dlqService{repo: repo}

	got, err := svc.GetDLQMessages(context.Background(), &dtov1.GetDLQMessagesReq{PageNum: 2, PageSize: 10})
	if err != nil {
		t.Fatalf("GetDLQMessages returned error: %v", err)
	}
	if got != wantResp {
		t.Fatal("expected repo response to be returned")
	}
}

func TestDLQService_RetryReturnsRepoErrorBeforePublishing(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("repo failed")
	repo := &mockDLQServiceRepo{
		retryFn: func(context.Context, int64) (*dtov1.RetryDLQMessageData, error) {
			return nil, wantErr
		},
	}
	svc := &dlqService{repo: repo, nc: nil}

	err := svc.RetryDLQMessage(context.Background(), 99)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
	if repo.retryCalledWith != 99 {
		t.Fatalf("expected retry id 99, got %d", repo.retryCalledWith)
	}
}

func TestDLQService_MarkAsResolvedDelegatesToRepo(t *testing.T) {
	t.Parallel()

	repo := &mockDLQServiceRepo{}
	svc := &dlqService{repo: repo}

	if err := svc.MarkAsResolved(context.Background(), 7); err != nil {
		t.Fatalf("MarkAsResolved returned error: %v", err)
	}
	if repo.resolvedCalledWith != 7 {
		t.Fatalf("expected resolved id 7, got %d", repo.resolvedCalledWith)
	}
}

func TestDLQService_MarkAsResolvedPropagatesRepoError(t *testing.T) {
	t.Parallel()

	wantErr := errors.New("not found")
	repo := &mockDLQServiceRepo{
		resolvedFn: func(context.Context, int64) (string, error) {
			return "", wantErr
		},
	}
	svc := &dlqService{repo: repo}

	err := svc.MarkAsResolved(context.Background(), 7)
	if !errors.Is(err, wantErr) {
		t.Fatalf("expected repo error, got %v", err)
	}
}
