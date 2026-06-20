package v1

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
)

type mockAuthService struct {
	registerFn func(context.Context, *dtov1.UserRegisterReq) (*dtov1.UserRegisterConflictResp, error)
	loginFn    func(context.Context, *dtov1.UserLoginReq) (*dtov1.TokenPair, error)
	refreshFn  func(context.Context, string) (*dtov1.TokenPair, error)
	logoutFn   func(context.Context, int64) error
	meFn       func(context.Context, int64) (*dtov1.MeResp, error)

	logoutUserID int64
}

func (m *mockAuthService) Register(ctx context.Context, req *dtov1.UserRegisterReq) (*dtov1.UserRegisterConflictResp, error) {
	if m.registerFn != nil {
		return m.registerFn(ctx, req)
	}
	return nil, nil
}

func (m *mockAuthService) Login(ctx context.Context, req *dtov1.UserLoginReq) (*dtov1.TokenPair, error) {
	if m.loginFn != nil {
		return m.loginFn(ctx, req)
	}
	return &dtov1.TokenPair{AccessToken: "at", RefreshToken: "rt"}, nil
}

func (m *mockAuthService) Refresh(ctx context.Context, token string) (*dtov1.TokenPair, error) {
	if m.refreshFn != nil {
		return m.refreshFn(ctx, token)
	}
	return &dtov1.TokenPair{AccessToken: "new-at", RefreshToken: "new-rt"}, nil
}

func (m *mockAuthService) Logout(ctx context.Context, userID int64) error {
	m.logoutUserID = userID
	if m.logoutFn != nil {
		return m.logoutFn(ctx, userID)
	}
	return nil
}

func (m *mockAuthService) Me(ctx context.Context, userID int64) (*dtov1.MeResp, error) {
	if m.meFn != nil {
		return m.meFn(ctx, userID)
	}
	return &dtov1.MeResp{ID: userID, Username: "ice", Email: "ice@example.com", Role: consts.RoleUser}, nil
}

func authHandlerTestConfig() *config.JwtConfig {
	return &config.JwtConfig{
		AccessExpires:  time.Hour,
		RefreshExpires: 24 * time.Hour,
	}
}

func performAuthRequest(handler gin.HandlerFunc, method, target string, body string, setup ...func(*gin.Context)) *httptest.ResponseRecorder {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Handle(method, target, func(c *gin.Context) {
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

func decodeResponse(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()

	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response %q: %v", w.Body.String(), err)
	}
	return resp
}

func TestAuthHandler_RegisterSuccess(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		registerFn: func(_ context.Context, req *dtov1.UserRegisterReq) (*dtov1.UserRegisterConflictResp, error) {
			if req.Username != "ice" || req.Email != "ice@example.com" {
				t.Fatalf("unexpected request: %#v", req)
			}
			return nil, nil
		},
	}
	h := NewAuthHandler(svc, authHandlerTestConfig())

	body := `{"username":"ice","password":"password123","email":"ice@example.com","code":"000000"}`
	w := performAuthRequest(h.Register, http.MethodPost, "/register", body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if int(resp["code"].(float64)) != consts.CodeSuccess {
		t.Fatalf("expected success code, got %#v", resp["code"])
	}
}

func TestAuthHandler_RegisterConflict(t *testing.T) {
	t.Parallel()

	conflict := &dtov1.UserRegisterConflictResp{IsUsernameConflict: true}
	svc := &mockAuthService{
		registerFn: func(context.Context, *dtov1.UserRegisterReq) (*dtov1.UserRegisterConflictResp, error) {
			return conflict, consts.ErrRegisterParamConflict
		},
	}
	h := NewAuthHandler(svc, authHandlerTestConfig())

	body := `{"username":"ice","password":"password123","email":"ice@example.com","code":"000000"}`
	w := performAuthRequest(h.Register, http.MethodPost, "/register", body)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if int(resp["code"].(float64)) != consts.CodeParamConflict {
		t.Fatalf("expected param conflict code, got %#v", resp["code"])
	}
}

func TestAuthHandler_LoginSetsCookies(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(&mockAuthService{}, authHandlerTestConfig())
	body := `{"identifier":"ice","password":"password123"}`
	w := performAuthRequest(h.Login, http.MethodPost, "/login", body)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	cookies := w.Result().Cookies()
	got := map[string]string{}
	for _, cookie := range cookies {
		got[cookie.Name] = cookie.Value
	}
	if got["access_token"] != "at" {
		t.Fatalf("expected access_token cookie, got %#v", got)
	}
	if got["refresh_token"] != "rt" {
		t.Fatalf("expected refresh_token cookie, got %#v", got)
	}
}

func TestAuthHandler_LoginMapsCredentialErrors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		err        error
		wantStatus int
		wantCode   int
	}{
		{name: "unknown identifier", err: pgx.ErrNoRows, wantStatus: http.StatusBadRequest, wantCode: consts.CodeInvalidIdentifier},
		{name: "wrong password", err: bcrypt.ErrMismatchedHashAndPassword, wantStatus: http.StatusBadRequest, wantCode: consts.CodeInvalidPassword},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := &mockAuthService{
				loginFn: func(context.Context, *dtov1.UserLoginReq) (*dtov1.TokenPair, error) {
					return nil, tt.err
				},
			}
			h := NewAuthHandler(svc, authHandlerTestConfig())
			body := `{"identifier":"ice","password":"password123"}`
			w := performAuthRequest(h.Login, http.MethodPost, "/login", body)

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

func TestAuthHandler_RefreshRequiresCookie(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(&mockAuthService{}, authHandlerTestConfig())
	w := performAuthRequest(h.Refresh, http.MethodPost, "/refresh", "")

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected status 401, got %d", w.Code)
	}
}

func TestAuthHandler_LogoutClearsCookiesAndUsesContextUserID(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{}
	h := NewAuthHandler(svc, authHandlerTestConfig())

	w := performAuthRequest(h.Logout, http.MethodPost, "/logout", "", func(c *gin.Context) {
		c.Set("user_id", int64(42))
	})

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
	if svc.logoutUserID != 42 {
		t.Fatalf("expected logout user id 42, got %d", svc.logoutUserID)
	}

	for _, cookie := range w.Result().Cookies() {
		if (cookie.Name == "access_token" || cookie.Name == "refresh_token") && cookie.MaxAge != -1 {
			t.Fatalf("expected cookie %s to be cleared, got MaxAge=%d", cookie.Name, cookie.MaxAge)
		}
	}
}

func TestAuthHandler_MeMapsNotFound(t *testing.T) {
	t.Parallel()

	svc := &mockAuthService{
		meFn: func(context.Context, int64) (*dtov1.MeResp, error) {
			return nil, pgx.ErrNoRows
		},
	}
	h := NewAuthHandler(svc, authHandlerTestConfig())

	w := performAuthRequest(h.Me, http.MethodGet, "/me", "", func(c *gin.Context) {
		c.Set("user_id", int64(42))
	})

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", w.Code)
	}
	resp := decodeResponse(t, w)
	if int(resp["code"].(float64)) != consts.CodeUserNotFound {
		t.Fatalf("expected user not found code, got %#v", resp["code"])
	}
}

func TestAuthHandler_InternalErrorsReturn500(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(&mockAuthService{
		logoutFn: func(context.Context, int64) error {
			return errors.New("redis down")
		},
	}, authHandlerTestConfig())

	w := performAuthRequest(h.Logout, http.MethodPost, "/logout", "", func(c *gin.Context) {
		c.Set("user_id", int64(42))
	})

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d", w.Code)
	}
}

func TestAuthHandler_InvalidJSON(t *testing.T) {
	t.Parallel()

	h := NewAuthHandler(&mockAuthService{}, authHandlerTestConfig())

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/register", h.Register)
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBufferString(`{"username":`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
