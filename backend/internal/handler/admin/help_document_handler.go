package admin

import (
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	middleware2 "github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"strconv"
)

type HelpDocumentHandler struct{ service *service.HelpDocumentService }

func NewHelpDocumentHandler(s *service.HelpDocumentService) *HelpDocumentHandler {
	return &HelpDocumentHandler{service: s}
}

type helpDocumentRequest struct {
	Slug            string         `json:"slug"`
	Title           string         `json:"title"`
	Category        string         `json:"category"`
	Summary         string         `json:"summary"`
	ContentMarkdown string         `json:"content_markdown"`
	SelectorSchema  map[string]any `json:"selector_schema"`
	Version         int            `json:"version"`
}

func helpID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id < 1 {
		response.BadRequest(c, "invalid help document id")
		return 0, false
	}
	return id, true
}
func (h *HelpDocumentHandler) List(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	page, size := response.ParsePagination(c)
	items, result, err := h.service.List(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: size}, c.Query("category"), c.Query("status"))
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, result.Total, page, size)
}
func (h *HelpDocumentHandler) Get(c *gin.Context) {
	c.Header("Cache-Control", "private, no-store")
	id, ok := helpID(c)
	if !ok {
		return
	}
	d, err := h.service.GetByID(c.Request.Context(), id)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, d)
}
func (h *HelpDocumentHandler) Create(c *gin.Context)  { h.save(c, false, false) }
func (h *HelpDocumentHandler) Update(c *gin.Context)  { h.save(c, true, false) }
func (h *HelpDocumentHandler) Preview(c *gin.Context) { h.save(c, true, true) }
func (h *HelpDocumentHandler) save(c *gin.Context, existing, preview bool) {
	c.Header("Cache-Control", "private, no-store")
	actor, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	var id int64
	if existing {
		id, ok = helpID(c)
		if !ok {
			return
		}
		if _, err := h.service.GetByID(c.Request.Context(), id); err != nil {
			response.ErrorFrom(c, err)
			return
		}
	}
	var req helpDocumentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "invalid help document")
		return
	}
	d := &service.HelpDocument{ID: id, Slug: req.Slug, Title: req.Title, Category: req.Category, Summary: req.Summary, ContentMarkdown: req.ContentMarkdown, SelectorSchema: req.SelectorSchema, Version: req.Version, UpdatedBy: &actor.UserID}
	var err error
	if preview {
		err = h.service.Preview(d)
	} else {
		err = h.service.Save(c.Request.Context(), d)
	}
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, d)
}
func (h *HelpDocumentHandler) Publish(c *gin.Context)   { h.transition(c, "publish") }
func (h *HelpDocumentHandler) Rollback(c *gin.Context)  { h.transition(c, "rollback") }
func (h *HelpDocumentHandler) Unpublish(c *gin.Context) { h.transition(c, "unpublish") }
func (h *HelpDocumentHandler) transition(c *gin.Context, action string) {
	c.Header("Cache-Control", "private, no-store")
	id, ok := helpID(c)
	if !ok {
		return
	}
	actor, ok := middleware2.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "authentication required")
		return
	}
	var req struct {
		Version int `json:"version" binding:"required,min=1"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "version required")
		return
	}
	d, err := h.service.Transition(c.Request.Context(), id, action, req.Version, actor.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, d)
}
func (h *HelpDocumentHandler) Delete(c *gin.Context) {
	id, ok := helpID(c)
	if !ok {
		return
	}
	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, gin.H{"message": "help document deleted"})
}
