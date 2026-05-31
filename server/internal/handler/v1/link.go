package v1

import (
	"errors"
	"github.com/ICE-awa/renice-sl/internal/consts"
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/service"
	"github.com/ICE-awa/renice-sl/shared/httputil"
	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"net/http"
	"net/netip"
	"strconv"
	"strings"
	"time"
)

type LinkHandler struct {
	svc service.LinkService
}

func NewLinkHandler(svc service.LinkService) *LinkHandler {
	return &LinkHandler{svc: svc}
}

// POST /api/v1/link
func (h *LinkHandler) CreateLink(c *gin.Context) {
	var req *dtov1.CreateLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid params")
		return
	}

	req.UserID = c.GetInt64("user_id")

	if err := h.svc.CreateLink(c.Request.Context(), req); err != nil {
		if errors.Is(err, consts.ErrFailedToGenerateCode) {
			httputil.Fail(c, http.StatusInternalServerError, consts.CodeFailedToGenerateCode, consts.ErrFailedToGenerateCode.Error())
			return
		} else if errors.Is(err, consts.ErrURLNotAllowed) {
			httputil.Fail(c, http.StatusBadRequest, consts.CodeURLNotAllowed, consts.ErrURLNotAllowed.Error())
		} else {
			httputil.InternalServerError(c, "Server Temporarily Unavailable")
			return
		}
	}

	httputil.OK(c, nil)

}

func (h *LinkHandler) GetLinks(c *gin.Context) {
	var req dtov1.GetLinksReq
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid params")
		return
	}

	req.UserID = c.GetInt64("user_id")
	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	data, err := h.svc.GetLinks(c.Request.Context(), &req)
	if err != nil {
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, data)
}

func (h *LinkHandler) UpdateLink(c *gin.Context) {
	var req dtov1.UpdateLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid params")
		return
	}

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.BadRequest(c, "Invalid id")
		return
	}

	req.ID = int64(id)
	req.UserID = c.GetInt64("user_id")

	if err := h.svc.UpdateLink(c.Request.Context(), &req); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Fail(c, http.StatusNotFound, consts.CodeLinkNotFound, "Link Not Found")
			return
		} else if errors.Is(err, consts.ErrURLNotAllowed) {
			httputil.Fail(c, http.StatusBadRequest, consts.CodeURLNotAllowed, consts.ErrURLNotAllowed.Error())
			return
		}
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, nil)
}

func (h *LinkHandler) GetLinkByID(c *gin.Context) {
	idStr := c.Param("id")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.BadRequest(c, "Invalid id")
		return
	}

	userID := c.GetInt64("user_id")

	link, err := h.svc.GetLinkByID(c.Request.Context(), int64(id), userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Fail(c, http.StatusNotFound, consts.CodeLinkNotFound, "Link Not Found")
			return
		} else {
			httputil.InternalServerError(c, "Server Temporarily Unavailable")
			return
		}
	}

	httputil.OK(c, link)
}

func (h *LinkHandler) DeleteLink(c *gin.Context) {
	var req dtov1.DeleteLinkReq

	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.BadRequest(c, "Invalid id")
		return
	}

	req.ID = int64(id)
	req.UserID = c.GetInt64("user_id")

	if err := h.svc.DeleteLink(c.Request.Context(), &req); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Fail(c, http.StatusNotFound, consts.CodeLinkNotFound, "Link Not Found")
			return
		}
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, nil)
}

func (h *LinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")
	ipStr := c.ClientIP()
	ip, err := netip.ParseAddr(ipStr)
	if err != nil {
		httputil.Fail(c, http.StatusInternalServerError, consts.CodeInternalServerError, "Failed to parse IP")
		return
	}

	ua := c.Request.UserAgent()
	referer := c.Request.Referer()
	clickedAt := time.Now()

	req := &dtov1.ClickLinkReq{
		Code:      code,
		IP:        ip,
		UserAgent: ua,
		Referer:   referer,
		ClickedAt: clickedAt,
		SkipStats: isBrowserPrefetch(c),
	}

	originalURL, err := h.svc.Redirect(c.Request.Context(), req)
	if err != nil {
		if errors.Is(err, consts.ErrLinkNotFound) {
			httputil.Fail(c, http.StatusNotFound, consts.CodeLinkNotFound, "Link Not Found")
			return
		} else if errors.Is(err, consts.ErrLinkInactive) {
			httputil.Fail(c, http.StatusForbidden, consts.CodeLinkInactive, "Link Inactive")
			return
		} else if errors.Is(err, consts.ErrLinkExpired) {
			httputil.Fail(c, http.StatusForbidden, consts.CodeLinkExpired, "Link Expired")
			return
		} else if errors.Is(err, consts.ErrLinkPending) {
			httputil.Fail(c, http.StatusForbidden, consts.CodeLinkPending, "Link Pending")
			return
		} else if errors.Is(err, consts.ErrLinkUnsafe) {
			httputil.Fail(c, http.StatusForbidden, consts.CodeLinkUnsafe, "Link Unsafe")
			return
		} else if errors.Is(err, consts.ErrLinkUnknown) {
			httputil.Fail(c, http.StatusForbidden, consts.CodeLinkUnknown, "Link Unknown")
			return
		}
		httputil.Fail(c, http.StatusInternalServerError, consts.CodeFailedToRedirect, "Failed to redirect")
		return
	}

	httputil.Redirect(c, http.StatusFound, originalURL)
}

func (h *LinkHandler) GetStats(c *gin.Context) {
	userID := c.GetInt64("user_id")

	resp, err := h.svc.GetStats(c.Request.Context(), userID)
	if err != nil {
		httputil.InternalServerError(c, "Failed to get stats")
		return
	}

	httputil.OK(c, resp)
}

func isBrowserPrefetch(c *gin.Context) bool {
	secPurpose := strings.ToLower(c.GetHeader("Sec-Purpose"))
	purpose := strings.ToLower(c.GetHeader("Purpose"))

	return strings.Contains(secPurpose, "prefetch") ||
		strings.Contains(secPurpose, "prerender") ||
		strings.Contains(purpose, "prefetch") ||
		strings.Contains(purpose, "prerender")
}
