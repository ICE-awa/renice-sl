package v1

import (
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/service"
	"github.com/ICE-awa/renice-sl/shared/httputil"
	"github.com/gin-gonic/gin"
	"log/slog"
)

type StatHandler struct {
	statSvc service.StatService
}

func NewStatHandler(statSvc service.StatService) *StatHandler {
	return &StatHandler{statSvc: statSvc}
}

func (h *StatHandler) GetLinkStats(c *gin.Context) {
	var req *dtov1.GetLinkStatReq
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid params")
		return
	}

	stats, err := h.statSvc.GetLinkStats(c.Request.Context(), req)
	if err != nil {
		slog.Warn("failed to get link stats",
			slog.String("error", err.Error()))
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, stats)
}

func (h *StatHandler) GetClickStats(c *gin.Context) {
	var req *dtov1.GetClickStatReq
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid params")
		return
	}

	stats, err := h.statSvc.GetClickStats(c.Request.Context(), req)
	if err != nil {
		slog.Warn("failed to get click stats",
			slog.String("error", err.Error()))
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, stats)
}

func (h *StatHandler) GetUserStats(c *gin.Context) {
	var req *dtov1.GetUserStatReq
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid params")
		return
	}

	stats, err := h.statSvc.GetUserStats(c.Request.Context(), req)
	if err != nil {
		slog.Warn("failed to get user stats",
			slog.String("error", err.Error()))
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, stats)
}
