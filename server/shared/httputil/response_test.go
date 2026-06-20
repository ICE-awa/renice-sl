package httputil

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ICE-awa/renice-sl/internal/consts"
	"github.com/gin-gonic/gin"
)

func runResponseHandler(handler gin.HandlerFunc) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/test", handler)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)
	return w
}

func decodeBody(t *testing.T, w *httptest.ResponseRecorder) Response {
	t.Helper()

	var resp Response
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response body %q: %v", w.Body.String(), err)
	}
	return resp
}

func TestOK(t *testing.T) {
	t.Parallel()

	w := runResponseHandler(func(c *gin.Context) {
		OK(c, gin.H{"id": 42})
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	resp := decodeBody(t, w)
	if resp.Code != consts.CodeSuccess || resp.Message != "ok" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestFail(t *testing.T) {
	t.Parallel()

	w := runResponseHandler(func(c *gin.Context) {
		Fail(c, http.StatusBadRequest, consts.CodeInvalidParam, "Invalid params")
	})

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
	resp := decodeBody(t, w)
	if resp.Code != consts.CodeInvalidParam || resp.Message != "Invalid params" {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestFailWithData(t *testing.T) {
	t.Parallel()

	w := runResponseHandler(func(c *gin.Context) {
		FailWithData(c, http.StatusConflict, consts.CodeParamConflict, gin.H{"field": "username"}, "Conflict")
	})

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}
	resp := decodeBody(t, w)
	if resp.Code != consts.CodeParamConflict || resp.Message != "Conflict" || resp.Data == nil {
		t.Fatalf("unexpected response: %#v", resp)
	}
}

func TestRedirect(t *testing.T) {
	t.Parallel()

	w := runResponseHandler(func(c *gin.Context) {
		Redirect(c, http.StatusFound, "https://example.com")
	})

	if w.Code != http.StatusFound {
		t.Fatalf("expected status 302, got %d", w.Code)
	}
	if got := w.Header().Get("Location"); got != "https://example.com" {
		t.Fatalf("expected Location header, got %q", got)
	}
}
