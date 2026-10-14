package admin

import (
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/handler/dto"
	"github.com/Wei-Shaw/sub2api/internal/handler/ticketrequest"
	"github.com/Wei-Shaw/sub2api/internal/pkg/pagination"
	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/server/middleware"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type SupportTicketHandler struct{ service *service.SupportTicketService }

func NewSupportTicketHandler(s *service.SupportTicketService) *SupportTicketHandler {
	return &SupportTicketHandler{service: s}
}

type ReplySupportTicketRequest struct {
	Body string `json:"body"`
}
type UpdateSupportTicketStatusRequest struct {
	Status string `json:"status"`
}

func adminTicketID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid ticket ID")
		return 0, false
	}
	return id, true
}

func (h *SupportTicketHandler) List(c *gin.Context) {
	page, size, ok := ticketrequest.Pagination(c)
	if !ok {
		return
	}
	items, result, err := h.service.ListForAdmin(c.Request.Context(), pagination.PaginationParams{Page: page, PageSize: size}, c.Query("category"), c.Query("queue"), c.Query("search"))
	if response.ErrorFrom(c, err) {
		return
	}
	response.Paginated(c, dto.SupportTicketsFromService(items), result.Total, result.Page, result.PageSize)
}

func (h *SupportTicketHandler) Get(c *gin.Context) {
	id, ok := adminTicketID(c)
	if !ok {
		return
	}
	ticket, err := h.service.GetForAdmin(c.Request.Context(), id)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.SupportTicketFromService(ticket))
}

func (h *SupportTicketHandler) Reply(c *gin.Context) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return
	}
	id, ok := adminTicketID(c)
	if !ok {
		return
	}
	var req ReplySupportTicketRequest
	if !ticketrequest.Bind(c, &req) {
		return
	}
	ticket, err := h.service.ReplyAsAdmin(c.Request.Context(), subject.UserID, id, req.Body)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.SupportTicketFromService(ticket))
}

func (h *SupportTicketHandler) UpdateStatus(c *gin.Context) {
	id, ok := adminTicketID(c)
	if !ok {
		return
	}
	var req UpdateSupportTicketStatusRequest
	if !ticketrequest.Bind(c, &req) {
		return
	}
	ticket, err := h.service.UpdateStatus(c.Request.Context(), id, req.Status)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.SupportTicketFromService(ticket))
}
