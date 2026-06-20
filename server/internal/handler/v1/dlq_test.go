package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/gin-gonic/gin"
)

type mockDLQService struct {
	getReq     *dtov1.GetDLQMessagesReq
	retryID    int64
	resolvedID int64
	err        error
}

func (m *mockDLQService) GetDLQMessages(_ context.Context, req *dtov1.GetDLQMessagesReq) (*dtov1.GetDLQMessagesResp, error) {
	m.getReq = req
	if m.err != nil {
		return nil, m.err
	}
	return &dtov1.GetDLQMessagesResp{Total: 0}, nil
}

func (m *mockDLQService) RetryDLQMessage(_ context.Context, id int64) error {
	m.retryID = id
	return m.err
}

func (m *mockDLQService) MarkAsResolved(_ context.Context, id int64) error {
	m.resolvedID = id
	return m.err
}

func performDLQRequest(handler gin.HandlerFunc, method, pattern, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Handle(method, pattern, handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestDLQHandler_GetDLQMessagesDefaultsPagination(t *testing.T) {
	t.Parallel()

	svc := &mockDLQService{}
	h := NewDLQHandler(svc)

	w := performDLQRequest(h.GetDLQMessages, http.MethodGet, "/dlq", "/dlq")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.getReq == nil {
		t.Fatal("expected service request")
	}
	if svc.getReq.PageNum != 1 || svc.getReq.PageSize != 10 {
		t.Fatalf("unexpected default pagination: %#v", svc.getReq)
	}
}

func TestDLQHandler_RetryInvalidID(t *testing.T) {
	t.Parallel()

	h := NewDLQHandler(&mockDLQService{})
	w := performDLQRequest(h.RetryDLQMessage, http.MethodPost, "/dlq/retry/:id", "/dlq/retry/bad")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestDLQHandler_RetryPassesID(t *testing.T) {
	t.Parallel()

	svc := &mockDLQService{}
	h := NewDLQHandler(svc)
	w := performDLQRequest(h.RetryDLQMessage, http.MethodPost, "/dlq/retry/:id", "/dlq/retry/99")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.retryID != 99 {
		t.Fatalf("expected retry id 99, got %d", svc.retryID)
	}
}

func TestDLQHandler_MarkAsResolvedPassesID(t *testing.T) {
	t.Parallel()

	svc := &mockDLQService{}
	h := NewDLQHandler(svc)
	w := performDLQRequest(h.MarkAsResolved, http.MethodPost, "/dlq/resolve/:id", "/dlq/resolve/7")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.resolvedID != 7 {
		t.Fatalf("expected resolved id 7, got %d", svc.resolvedID)
	}
}

func TestDLQHandler_ServiceErrorReturns500(t *testing.T) {
	t.Parallel()

	h := NewDLQHandler(&mockDLQService{err: errors.New("service failed")})
	w := performDLQRequest(h.MarkAsResolved, http.MethodPost, "/dlq/resolve/:id", "/dlq/resolve/7")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}
