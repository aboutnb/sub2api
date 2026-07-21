package admin

import (
	"strconv"
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"

	"github.com/gin-gonic/gin"
)

type AuthIPBanHandler struct {
	service *service.AuthIPBanService
}

func NewAuthIPBanHandler(banService *service.AuthIPBanService) *AuthIPBanHandler {
	return &AuthIPBanHandler{service: banService}
}

// List returns authentication IP ban records for administrator review.
func (h *AuthIPBanHandler) List(c *gin.Context) {
	page, pageSize := response.ParsePagination(c)
	if pageSize > 200 {
		pageSize = 200
	}
	status := strings.ToLower(strings.TrimSpace(c.Query("status")))
	switch status {
	case "", "all", "active", "expired", "released":
	default:
		response.BadRequest(c, "无效的封禁状态")
		return
	}
	result, err := h.service.List(c.Request.Context(), &service.AuthIPBanFilter{
		Page:     page,
		PageSize: pageSize,
		Status:   status,
		Query:    strings.TrimSpace(c.Query("q")),
	})
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, result.Items, int64(result.Total), result.Page, result.PageSize)
}

func (h *AuthIPBanHandler) Policy(c *gin.Context) {
	response.Success(c, h.service.Policies())
}

type authIPBanReleaseRequest struct {
	Note string `json:"note"`
}

// Release removes an active automatic restriction while preserving its audit record.
func (h *AuthIPBanHandler) Release(c *gin.Context) {
	id, err := strconv.ParseInt(strings.TrimSpace(c.Param("id")), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "无效的封禁记录 ID")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok || subject.UserID <= 0 {
		response.Unauthorized(c, "未登录或登录已过期")
		return
	}
	var req authIPBanReleaseRequest
	if c.Request.ContentLength != 0 {
		if err := c.ShouldBindJSON(&req); err != nil {
			response.BadRequest(c, "解除说明格式无效")
			return
		}
	}
	record, err := h.service.Release(c.Request.Context(), id, subject.UserID, req.Note)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, record)
}
