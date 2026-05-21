package admin

import (
	"strings"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type projectMihomoRequest struct {
	SubscriptionURL   string   `json:"subscription_url"`
	SubscriptionURLs  []string `json:"subscription_urls"`
	SubscriptionNames []string `json:"subscription_names"`
	SubscriptionUA    string   `json:"subscription_user_agent"`
	UpdateInterval    int      `json:"update_interval"`
	Protocol          string   `json:"protocol" binding:"required,oneof=http https socks5 socks5h"`
	TargetHost        string   `json:"target_host" binding:"required"`
	StartPort         int      `json:"start_port" binding:"required,min=1,max=65535"`
	ListenerCount     int      `json:"listener_count" binding:"required,min=1,max=32"`
	ControllerURL     string   `json:"controller_url" binding:"required"`
	ControllerSecret  string   `json:"controller_secret"`
	ProxyNamePrefix   string   `json:"proxy_name_prefix"`
	ListenerRegions   []string `json:"listener_regions"`
}

type projectMihomoNodeTestRequest struct {
	projectMihomoRequest
	Node service.ProjectMihomoNode `json:"node" binding:"required"`
}

func (h *ProxyHandler) GetProjectMihomo(c *gin.Context) {
	status, err := h.projectMihomoService.GetStatus(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, status)
}

func (h *ProxyHandler) UpdateProjectMihomo(c *gin.Context) {
	var req projectMihomoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	settings, err := h.projectMihomoService.SetSettings(c.Request.Context(), projectMihomoSettingsFromRequest(req))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, settings)
}

func (h *ProxyHandler) SyncProjectMihomo(c *gin.Context) {
	var req projectMihomoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := h.projectMihomoService.Sync(c.Request.Context(), projectMihomoSettingsFromRequest(req))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProxyHandler) TestProjectMihomoNodes(c *gin.Context) {
	var req projectMihomoRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := h.projectMihomoService.TestNodes(c.Request.Context(), projectMihomoSettingsFromRequest(req))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *ProxyHandler) TestProjectMihomoNode(c *gin.Context) {
	var req projectMihomoNodeTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}

	result, err := h.projectMihomoService.TestNode(c.Request.Context(), projectMihomoSettingsFromRequest(req.projectMihomoRequest), req.Node)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func projectMihomoSettingsFromRequest(req projectMihomoRequest) *service.ProjectMihomoSettings {
	return &service.ProjectMihomoSettings{
		SubscriptionURL:   strings.TrimSpace(req.SubscriptionURL),
		SubscriptionURLs:  req.SubscriptionURLs,
		SubscriptionNames: req.SubscriptionNames,
		SubscriptionUA:    strings.TrimSpace(req.SubscriptionUA),
		UpdateInterval:    req.UpdateInterval,
		Protocol:          strings.TrimSpace(req.Protocol),
		TargetHost:        strings.TrimSpace(req.TargetHost),
		StartPort:         req.StartPort,
		ListenerCount:     req.ListenerCount,
		ControllerURL:     strings.TrimSpace(req.ControllerURL),
		ControllerSecret:  strings.TrimSpace(req.ControllerSecret),
		ProxyNamePrefix:   strings.TrimSpace(req.ProxyNamePrefix),
		ListenerRegions:   req.ListenerRegions,
	}
}
