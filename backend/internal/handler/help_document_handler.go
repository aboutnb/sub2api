package handler

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type HelpDocumentHandler struct{ service *service.HelpDocumentService }

func NewHelpDocumentHandler(s *service.HelpDocumentService) *HelpDocumentHandler {
	return &HelpDocumentHandler{service: s}
}
func (h *HelpDocumentHandler) List(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	items, err := h.service.ListPublished(c.Request.Context(), c.Query("category"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, items)
}
func (h *HelpDocumentHandler) Get(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	item, err := h.service.GetPublished(c.Request.Context(), c.Param("slug"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, item)
}
