package v1

import (
	"errors"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/service"
	"github.com/ICE-awa/renice-sl/shared/config"
	"github.com/ICE-awa/renice-sl/shared/httputil"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
	"net/http"
)

type AuthHandler struct {
	svc service.AuthService
	cfg *config.JwtConfig
}

func NewAuthHandler(svc service.AuthService, cfg *config.JwtConfig) *AuthHandler {
	return &AuthHandler{svc: svc, cfg: cfg}
}

// POST /api/v1/auth/register
func (h *AuthHandler) Register(c *gin.Context) {
	var req dtov1.UserRegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Fail(c, http.StatusBadRequest, consts.CodeInvalidParam, "Invalid params")
		return
	}

	conflict, err := h.svc.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, consts.ErrRegisterParamConflict) {
			httputil.FailWithData(c, http.StatusConflict, consts.CodeParamConflict, conflict, "Conflict username or email")
		} else if errors.Is(err, consts.ErrInvalidEmailCode) {
			httputil.Fail(c, http.StatusBadRequest, consts.CodeInvalidEmailCode, "Invalid email code")
		} else {
			slog.Error("[handler] Internal Server Error in Register",
				slog.String("error", err.Error()))
			httputil.InternalServerError(c, "Server Temporarily Unavailable")
		}
		return
	}

	httputil.OK(c, nil)
}

// POST /api/v1/auth/login
func (h *AuthHandler) Login(c *gin.Context) {
	var req dtov1.UserLoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.Fail(c, http.StatusBadRequest, consts.CodeInvalidParam, "Invalid params")
		return
	}

	tokenPair, err := h.svc.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Fail(c, http.StatusBadRequest, consts.CodeInvalidIdentifier, "Invalid Identifier")
		} else if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			httputil.Fail(c, http.StatusBadRequest, consts.CodeInvalidPassword, "Invalid Password")
		} else {
			slog.Error("[handler] Internal Server Error in Login",
				slog.String("error", err.Error()))
			httputil.InternalServerError(c, "Server Temporarily Unavailable")
		}
		return
	}

	// 设置 Cookie
	c.SetCookie(
		"access_token",
		tokenPair.AccessToken,
		int(h.cfg.AccessExpires.Seconds()),
		"/",
		"",
		false,
		true,
	)

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(h.cfg.RefreshExpires.Seconds()),
		"/api/v1/auth/refresh",
		"",
		false,
		true,
	)

	resp := &dtov1.UserLoginResp{
		ExpiresIn: int64(h.cfg.AccessExpires.Seconds()),
	}
	httputil.OK(c, resp)
}

// POST /api/v1/auth/refresh
func (h *AuthHandler) Refresh(c *gin.Context) {
	token, err := c.Cookie("refresh_token")
	if err != nil {
		httputil.Fail(c, http.StatusBadRequest, consts.CodeInvalidRefreshToken, "Invalid refresh token")
		return
	}

	tokenPair, err := h.svc.Refresh(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, consts.ErrInvalidRefreshToken) {
			httputil.Fail(c, http.StatusBadRequest, consts.CodeInvalidRefreshToken, "Invalid refresh token")
		} else {
			httputil.InternalServerError(c, "Server Temporarily Unavailable")
		}
		return
	}

	// 设置 Cookie
	c.SetCookie(
		"access_token",
		tokenPair.AccessToken,
		int(h.cfg.AccessExpires.Seconds()),
		"/",
		"",
		false,
		true,
	)

	c.SetCookie(
		"refresh_token",
		tokenPair.RefreshToken,
		int(h.cfg.RefreshExpires.Seconds()),
		"/api/v1/auth/refresh",
		"",
		false,
		true,
	)

	httputil.OK(c, nil)
}

// POST /api/v1/auth/logout
func (h *AuthHandler) Logout(c *gin.Context) {
	userID := c.GetInt64("user_id")

	err := h.svc.Logout(c.Request.Context(), userID)
	if err != nil {
		httputil.Fail(c, http.StatusInternalServerError, consts.CodeInternalServerError, "Server Temporarily Unavailable")
		return
	}

	c.SetCookie(
		"access_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	c.SetCookie(
		"refresh_token",
		"",
		-1,
		"/",
		"",
		false,
		true,
	)

	httputil.OK(c, nil)
}
