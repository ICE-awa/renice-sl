package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ICE-awa/renice-sl/internal/consts"
	"github.com/gin-gonic/gin"
)

func TestHealthHandler_Healthz(t *testing.T) {
	t.Parallel()
	gin.SetMode(gin.TestMode)

	h := NewHealthHandler(nil, nil, nil)
	r := gin.New()
	r.GET("/healthz", h.Healthz)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if int(body["code"].(float64)) != consts.CodeSuccess {
		t.Fatalf("expected success code, got %#v", body["code"])
	}
}
