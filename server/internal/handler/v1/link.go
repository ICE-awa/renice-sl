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
	"strconv"
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

	links, err := h.svc.GetLinks(c.Request.Context(), &req)
	if err != nil {
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, links)
}

func (h *LinkHandler) UpdateLink(c *gin.Context) {
	var req dtov1.UpdateLinkReq
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid params")
		return
	}

	req.UserID = c.GetInt64("user_id")

	if err := h.svc.UpdateLink(c.Request.Context(), &req); err != nil {
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

	link, err := h.svc.GetLinkByID(c.Request.Context(), int64(id))
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
	if err := c.ShouldBindJSON(&req); err != nil {
		httputil.BadRequest(c, "Invalid params")
		return
	}

	req.UserID = c.GetInt64("user_id")

	if err := h.svc.DeleteLink(c.Request.Context(), &req); err != nil {
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, nil)
}

func (h *LinkHandler) Redirect(c *gin.Context) {
	code := c.Param("code")

	originalURL, err := h.svc.Redirect(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			httputil.Fail(c, http.StatusNotFound, consts.CodeLinkNotFound, "Link Not Found")
			return
		}
		httputil.Fail(c, http.StatusInternalServerError, consts.CodeFailedToRedirect, "Failed to redirect")
		return
	}

	httputil.Redirect(c, http.StatusFound, originalURL)
}
