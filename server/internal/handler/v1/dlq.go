package v1

import (
	dtov1 "github.com/ICE-awa/renice-sl/internal/dto/v1"
	"github.com/ICE-awa/renice-sl/internal/service"
	"github.com/ICE-awa/renice-sl/shared/httputil"
	"github.com/gin-gonic/gin"
	"log/slog"
	"strconv"
)

type DLQHandler struct {
	svc service.DLQService
}

func NewDLQHandler(svc service.DLQService) *DLQHandler {
	return &DLQHandler{svc}
}

func (h *DLQHandler) GetDLQMessages(c *gin.Context) {
	var req dtov1.GetDLQMessagesReq
	if err := c.ShouldBindQuery(&req); err != nil {
		httputil.BadRequest(c, "Invalid params")
		return
	}

	if req.PageNum <= 0 {
		req.PageNum = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	messages, err := h.svc.GetDLQMessages(c.Request.Context(), &req)
	if err != nil {
		slog.Error("get dlq messages request failed",
			slog.String("request_id", c.GetString("X-Request-ID")),
			slog.String("handler", "DLQHandler.GetDLQMessages"),
			slog.String("error", err.Error()),
		)
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, messages)
}

func (h *DLQHandler) RetryDLQMessage(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.BadRequest(c, "Invalid id")
		return
	}

	err = h.svc.RetryDLQMessage(c.Request.Context(), int64(id))
	if err != nil {
		slog.Error("retry dlq message request failed",
			slog.String("request_id", c.GetString("X-Request-ID")),
			slog.String("handler", "DLQHandler.RetryDLQMessage"),
			slog.String("error", err.Error()),
		)
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, nil)
}

func (h *DLQHandler) MarkAsResolved(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		httputil.BadRequest(c, "Invalid id")
		return
	}

	err = h.svc.MarkAsResolved(c.Request.Context(), int64(id))
	if err != nil {
		slog.Error("mark as resolved request failed",
			slog.String("request_id", c.GetString("X-Request-ID")),
			slog.String("handler", "DLQHandler.MarkAsResolved"),
			slog.String("error", err.Error()),
		)
		httputil.InternalServerError(c, "Server Temporarily Unavailable")
		return
	}

	httputil.OK(c, nil)
}
