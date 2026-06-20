package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/model"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/util"
	"github.com/gin-gonic/gin"
)

type authMiddlewareUserRepo struct {
	user *model.User
	err  error
}

func (r *authMiddlewareUserRepo) CreateUser(context.Context, *model.User) (int64, error) {
	return 0, nil
}

func (r *authMiddlewareUserRepo) UpdateUser(context.Context, *dtov1.UserUpdateReq) error {
	return nil
}

func (r *authMiddlewareUserRepo) DeleteUser(context.Context, int64) error {
	return nil
}

func (r *authMiddlewareUserRepo) FindUserByID(context.Context, int64) (*model.User, error) {
	return r.user, r.err
}

func (r *authMiddlewareUserRepo) FindUserByIdentifier(context.Context, string) (*model.User, error) {
	return nil, nil
}

func (r *authMiddlewareUserRepo) CheckConflict(context.Context, string, string) (*dtov1.UserRegisterConflictResp, error) {
	return nil, nil
}

func testJwtConfig() *config.JwtConfig {
	return &config.JwtConfig{
		AccessSecret:   "test-access-secret",
		RefreshSecret:  "test-refresh-secret",
		AccessExpires:  time.Hour,
		RefreshExpires: time.Hour,
	}
}

func TestAuthRequiredRejectsMissingToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/protected", AuthRequired(testJwtConfig()), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}

	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	if int(body["code"].(float64)) != consts.CodeUnauthorized {
		t.Fatalf("expected unauthorized code, got %#v", body["code"])
	}
}

func TestAuthRequiredAcceptsValidAccessTokenCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testJwtConfig()
	token, err := util.GenerateAccessToken(cfg, 42, "ice")
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.GET("/protected", AuthRequired(cfg), func(c *gin.Context) {
		if got := c.GetInt64("user_id"); got != 42 {
			t.Fatalf("expected user_id 42, got %d", got)
		}
		if got := c.GetString("username"); got != "ice" {
			t.Fatalf("expected username ice, got %q", got)
		}
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", w.Code)
	}
}

func TestAuthRequiredRejectsRefreshTokenAsAccessToken(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := testJwtConfig()
	token, err := util.GenerateRefreshToken(cfg, 42, "ice")
	if err != nil {
		t.Fatal(err)
	}

	r := gin.New()
	r.GET("/protected", AuthRequired(cfg), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.AddCookie(&http.Cookie{Name: "access_token", Value: token})
	r.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAdminRequiredChecksUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		role       string
		wantStatus int
	}{
		{name: "admin allowed", role: consts.RoleAdmin, wantStatus: http.StatusNoContent},
		{name: "user forbidden", role: consts.RoleUser, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo := &authMiddlewareUserRepo{user: &model.User{ID: 42, Role: tt.role}}

			r := gin.New()
			r.GET("/admin",
				func(c *gin.Context) {
					c.Set("user_id", int64(42))
					c.Next()
				},
				AdminRequired(repo),
				func(c *gin.Context) {
					c.Status(http.StatusNoContent)
				},
			)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/admin", nil)
			r.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Fatalf("expected status %d, got %d", tt.wantStatus, w.Code)
			}
		})
	}
}
