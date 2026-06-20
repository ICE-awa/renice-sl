package v1

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/gin-gonic/gin"
)

type mockStatService struct {
	linkReq  *dtov1.GetLinkStatReq
	clickReq *dtov1.GetClickStatReq
	userReq  *dtov1.GetUserStatReq
	err      error
}

func (m *mockStatService) GetLinkStats(context.Context, *dtov1.GetLinkStatReq) ([]*dtov1.LinkStatItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*dtov1.LinkStatItem{{Time: time.Now(), Count: 1}}, nil
}

func (m *mockStatService) GetClickStats(context.Context, *dtov1.GetClickStatReq) ([]*dtov1.ClickStatItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*dtov1.ClickStatItem{{Time: time.Now(), Count: 2}}, nil
}

func (m *mockStatService) GetUserStats(context.Context, *dtov1.GetUserStatReq) ([]*dtov1.UserStatItem, error) {
	if m.err != nil {
		return nil, m.err
	}
	return []*dtov1.UserStatItem{{Time: time.Now(), Count: 3}}, nil
}

func performStatRequest(handler gin.HandlerFunc, target string) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/stats", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, target, nil)
	r.ServeHTTP(w, req)
	return w
}

func TestStatHandler_GetLinkStatsSuccess(t *testing.T) {
	t.Parallel()

	h := NewStatHandler(&mockStatService{})
	w := performStatRequest(h.GetLinkStats, "/stats?range=7&bucket=day")

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if int(resp["code"].(float64)) != consts.CodeSuccess {
		t.Fatalf("expected success code, got %#v", resp["code"])
	}
}

func TestStatHandler_InvalidBucket(t *testing.T) {
	t.Parallel()

	h := NewStatHandler(&mockStatService{err: consts.ErrInvalidBucket})
	w := performStatRequest(h.GetClickStats, "/stats?range=7&bucket=week")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if int(resp["code"].(float64)) != consts.CodeInvalidBucket {
		t.Fatalf("expected invalid bucket code, got %#v", resp["code"])
	}
}

func TestStatHandler_InternalError(t *testing.T) {
	t.Parallel()

	h := NewStatHandler(&mockStatService{err: errors.New("repo failed")})
	w := performStatRequest(h.GetUserStats, "/stats?range=7&bucket=day")

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestStatHandler_InvalidQuery(t *testing.T) {
	t.Parallel()

	h := NewStatHandler(&mockStatService{})
	w := performStatRequest(h.GetLinkStats, "/stats?range=0&bucket=day")

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
