package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type RegistrationProtectionHandler struct {
	service *service.RegistrationProtectionService
}

func NewRegistrationProtectionHandler(s *service.RegistrationProtectionService) *RegistrationProtectionHandler {
	return &RegistrationProtectionHandler{service: s}
}

func (h *RegistrationProtectionHandler) Settings(c *gin.Context) {
	settings, err := h.service.Settings.GetRegistrationProtectionSettings(c.Request.Context())
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, settings)
}

func (h *RegistrationProtectionHandler) ReleaseBlock(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "无效的来源限制 ID")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "登录已过期")
		return
	}
	var req struct {
		Note string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "解除请求格式无效")
		return
	}
	if response.ErrorFrom(c, h.service.ReleaseSourceBlock(c.Request.Context(), id, subject.UserID, req.Note)) {
		return
	}
	response.Success(c, gin.H{"message": "来源限制已解除"})
}

func (h *RegistrationProtectionHandler) UpdateSettings(c *gin.Context) {
	var settings service.RegistrationProtectionSettings
	if err := c.ShouldBindJSON(&settings); err != nil {
		response.BadRequest(c, "注册防护配置格式无效")
		return
	}
	if err := settings.Validate(); err != nil {
		response.BadRequest(c, err.Error())
		return
	}
	if response.ErrorFrom(c, h.service.Settings.SetRegistrationProtectionSettings(c.Request.Context(), settings)) {
		return
	}
	response.Success(c, settings)
}

func (h *RegistrationProtectionHandler) List(c *gin.Context) {
	page, size := response.ParsePagination(c)
	result, err := h.service.List(c.Request.Context(), service.RegistrationRiskFilter{
		Kind: c.Param("kind"), Status: c.Query("status"), Query: c.Query("q"), Page: page, PageSize: size,
	})
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, result)
}

func (h *RegistrationProtectionHandler) Review(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "无效的风险记录 ID")
		return
	}
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "登录已过期")
		return
	}
	var req struct {
		Action string `json:"action"`
		Note   string `json:"note"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "审核请求格式无效")
		return
	}
	if response.ErrorFrom(c, h.service.ReviewAccount(c.Request.Context(), id, subject.UserID, req.Action, req.Note)) {
		return
	}
	response.Success(c, gin.H{"message": "审核操作已完成"})
}
