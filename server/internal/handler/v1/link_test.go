package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
)

type mockLinkService struct {
	createFn      func(context.Context, *dtov1.CreateLinkReq) error
	getLinksFn    func(context.Context, *dtov1.GetLinksReq) (*dtov1.GetLinksResp, error)
	updateFn      func(context.Context, *dtov1.UpdateLinkReq) error
	getByIDFn     func(context.Context, int64, int64) (*dtov1.LinkItem, error)
	deleteFn      func(context.Context, *dtov1.DeleteLinkReq) error
	redirectFn    func(context.Context, *dtov1.ClickLinkReq) (string, error)
	getStatsFn    func(context.Context, int64) (*dtov1.GetStatsResponse, error)
	initBloomFn   func() error
	createReq     *dtov1.CreateLinkReq
	getLinksReq   *dtov1.GetLinksReq
	updateReq     *dtov1.UpdateLinkReq
	deleteReq     *dtov1.DeleteLinkReq
	redirectReq   *dtov1.ClickLinkReq
	statsUserID   int64
	getByIDID     int64
	getByIDUserID int64
}

func (m *mockLinkService) CreateLink(ctx context.Context, req *dtov1.CreateLinkReq) error {
	m.createReq = req
	if m.createFn != nil {
		return m.createFn(ctx, req)
	}
	return nil
}

func (m *mockLinkService) GetLinks(ctx context.Context, req *dtov1.GetLinksReq) (*dtov1.GetLinksResp, error) {
	m.getLinksReq = req
	if m.getLinksFn != nil {
		return m.getLinksFn(ctx, req)
	}
	return &dtov1.GetLinksResp{PageNum: req.PageNum, PageSize: req.PageSize}, nil
}

func (m *mockLinkService) UpdateLink(ctx context.Context, req *dtov1.UpdateLinkReq) error {
	m.updateReq = req
	if m.updateFn != nil {
		return m.updateFn(ctx, req)
	}
	return nil
}

func (m *mockLinkService) GetLinkByID(ctx context.Context, id int64, userID int64) (*dtov1.LinkItem, error) {
	m.getByIDID = id
	m.getByIDUserID = userID
	if m.getByIDFn != nil {
		return m.getByIDFn(ctx, id, userID)
	}
	return &dtov1.LinkItem{ID: id, Code: "abc123"}, nil
}

func (m *mockLinkService) DeleteLink(ctx context.Context, req *dtov1.DeleteLinkReq) error {
	m.deleteReq = req
	if m.deleteFn != nil {
		return m.deleteFn(ctx, req)
	}
	return nil
}

func (m *mockLinkService) Redirect(ctx context.Context, req *dtov1.ClickLinkReq) (string, error) {
	m.redirectReq = req
	if m.redirectFn != nil {
		return m.redirectFn(ctx, req)
	}
	return "https://example.com", nil
}

func (m *mockLinkService) GetStats(ctx context.Context, userID int64) (*dtov1.GetStatsResponse, error) {
	m.statsUserID = userID
	if m.getStatsFn != nil {
		return m.getStatsFn(ctx, userID)
	}
	return &dtov1.GetStatsResponse{LinkCount: 2, ViewCount: 3}, nil
}

func (m *mockLinkService) InitBloomFilter() error {
	if m.initBloomFn != nil {
		return m.initBloomFn()
	}
	return nil
}

func performLinkRequest(handler gin.HandlerFunc, method, pattern, target string, body string, setup ...func(*gin.Context)) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Handle(method, pattern, func(c *gin.Context) {
		for _, fn := range setup {
			fn(c)
		}
		handler(c)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(method, target, strings.NewReader(body))
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	r.ServeHTTP(w, req)
	return w
}

func TestLinkHandler_CreateLinkSetsUserID(t *testing.T) {
	t.Parallel()

	svc := &mockLinkService{}
	h := NewLinkHandler(svc)

	body := `{"original_url":"https://example.com"}`
	w := performLinkRequest(h.CreateLink, http.MethodPost, "/link", "/link", body, func(c *gin.Context) {
		c.Set("user_id", int64(42))
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.createReq == nil || svc.createReq.UserID != 42 {
		t.Fatalf("expected user id 42 in create request, got %#v", svc.createReq)
	}
}

func TestLinkHandler_GetLinksDefaultsPagination(t *testing.T) {
	t.Parallel()

	svc := &mockLinkService{}
	h := NewLinkHandler(svc)

	w := performLinkRequest(h.GetLinks, http.MethodGet, "/links", "/links", "", func(c *gin.Context) {
		c.Set("user_id", int64(42))
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.getLinksReq == nil {
		t.Fatal("expected service request")
	}
	if svc.getLinksReq.UserID != 42 || svc.getLinksReq.PageNum != 1 || svc.getLinksReq.PageSize != 10 {
		t.Fatalf("unexpected request defaults: %#v", svc.getLinksReq)
	}
}

func TestLinkHandler_UpdateLinkMapsPathAndUserID(t *testing.T) {
	t.Parallel()

	svc := &mockLinkService{}
	h := NewLinkHandler(svc)

	body := `{"status":"inactive"}`
	w := performLinkRequest(h.UpdateLink, http.MethodPut, "/link/:id", "/link/12", body, func(c *gin.Context) {
		c.Set("user_id", int64(42))
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.updateReq == nil || svc.updateReq.ID != 12 || svc.updateReq.UserID != 42 {
		t.Fatalf("unexpected update request: %#v", svc.updateReq)
	}
}

func TestLinkHandler_UpdateLinkInvalidID(t *testing.T) {
	t.Parallel()

	h := NewLinkHandler(&mockLinkService{})

	w := performLinkRequest(h.UpdateLink, http.MethodPut, "/link/:id", "/link/bad", `{}`)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestLinkHandler_GetLinkByIDNotFound(t *testing.T) {
	t.Parallel()

	svc := &mockLinkService{
		getByIDFn: func(context.Context, int64, int64) (*dtov1.LinkItem, error) {
			return nil, pgx.ErrNoRows
		},
	}
	h := NewLinkHandler(svc)

	w := performLinkRequest(h.GetLinkByID, http.MethodGet, "/link/:id", "/link/12", "", func(c *gin.Context) {
		c.Set("user_id", int64(42))
	})

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if int(resp["code"].(float64)) != consts.CodeLinkNotFound {
		t.Fatalf("expected link not found code, got %#v", resp["code"])
	}
}

func TestLinkHandler_DeleteLinkMapsPathAndUserID(t *testing.T) {
	t.Parallel()

	svc := &mockLinkService{}
	h := NewLinkHandler(svc)

	w := performLinkRequest(h.DeleteLink, http.MethodDelete, "/link/:id", "/link/12", "", func(c *gin.Context) {
		c.Set("user_id", int64(42))
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.deleteReq == nil || svc.deleteReq.ID != 12 || svc.deleteReq.UserID != 42 {
		t.Fatalf("unexpected delete request: %#v", svc.deleteReq)
	}
}

func TestLinkHandler_RedirectSetsSkipStatsForPrefetch(t *testing.T) {
	t.Parallel()

	svc := &mockLinkService{}
	h := NewLinkHandler(svc)

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/s/:code", h.Redirect)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/s/abc123", nil)
	req.Header.Set("Sec-Purpose", "prefetch")
	req.RemoteAddr = "203.0.113.10:12345"
	r.ServeHTTP(w, req)

	if w.Code != http.StatusFound {
		t.Fatalf("expected status 302, got %d", w.Code)
	}
	if location := w.Header().Get("Location"); location != "https://example.com" {
		t.Fatalf("expected redirect location, got %q", location)
	}
	if svc.redirectReq == nil {
		t.Fatal("expected redirect request")
	}
	if svc.redirectReq.Code != "abc123" {
		t.Fatalf("expected code abc123, got %q", svc.redirectReq.Code)
	}
	if !svc.redirectReq.SkipStats {
		t.Fatal("prefetch request should skip stats")
	}
}

func TestLinkHandler_RedirectMapsDomainErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
	}{
		{name: "not found", err: consts.ErrLinkNotFound, wantStatus: http.StatusNotFound, wantCode: consts.CodeLinkNotFound},
		{name: "inactive", err: consts.ErrLinkInactive, wantStatus: http.StatusForbidden, wantCode: consts.CodeLinkInactive},
		{name: "expired", err: consts.ErrLinkExpired, wantStatus: http.StatusForbidden, wantCode: consts.CodeLinkExpired},
		{name: "pending", err: consts.ErrLinkPending, wantStatus: http.StatusForbidden, wantCode: consts.CodeLinkPending},
		{name: "unsafe", err: consts.ErrLinkUnsafe, wantStatus: http.StatusForbidden, wantCode: consts.CodeLinkUnsafe},
		{name: "unknown", err: consts.ErrLinkUnknown, wantStatus: http.StatusForbidden, wantCode: consts.CodeLinkUnknown},
		{name: "other", err: errors.New("db down"), wantStatus: http.StatusInternalServerError, wantCode: consts.CodeFailedToRedirect},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &mockLinkService{
				redirectFn: func(context.Context, *dtov1.ClickLinkReq) (string, error) {
					return "", tt.err
				},
			}
			h := NewLinkHandler(svc)
			w := performLinkRequest(h.Redirect, http.MethodGet, "/s/:code", "/s/abc123", "")

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
			resp := decodeResponse(t, w)
			if int(resp["code"].(float64)) != tt.wantCode {
				t.Fatalf("expected code %d, got %#v", tt.wantCode, resp["code"])
			}
		})
	}
}

func TestLinkHandler_GetStatsUsesContextUserID(t *testing.T) {
	t.Parallel()

	svc := &mockLinkService{}
	h := NewLinkHandler(svc)

	w := performLinkRequest(h.GetStats, http.MethodGet, "/stats", "/stats", "", func(c *gin.Context) {
		c.Set("user_id", int64(42))
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.statsUserID != 42 {
		t.Fatalf("expected stats user id 42, got %d", svc.statsUserID)
	}
}
