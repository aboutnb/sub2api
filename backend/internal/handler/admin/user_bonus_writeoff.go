package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

func (h *UserHandler) GetBonusWriteoff(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	page, _ := strconv.Atoi(c.Query("page"))
	data, err := h.userService.GetBonusWriteoff(c.Request.Context(), id, page)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, data)
}
func (h *UserHandler) SetBonusWriteoff(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid user ID")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "Authentication required")
		return
	}
	var p service.BonusWriteoffPolicy
	if err := c.ShouldBindJSON(&p); err != nil {
		response.BadRequest(c, "Invalid policy")
		return
	}
	if err := h.userService.SetBonusWriteoff(c.Request.Context(), id, subject.UserID, p); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"updated": true})
}
