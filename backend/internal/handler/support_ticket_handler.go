package handler

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

type CreateSupportTicketRequest struct {
	Subject  string `json:"subject"`
	Body     string `json:"body"`
	Category string `json:"category"`
}

type ReplySupportTicketRequest struct {
	Body string `json:"body"`
}

func ticketUser(c *gin.Context) (int64, bool) {
	subject, ok := middleware.GetAuthSubjectFromContext(c)
	if !ok {
		response.Unauthorized(c, "User not found in context")
		return 0, false
	}
	return subject.UserID, true
}

func ticketID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid ticket ID")
		return 0, false
	}
	return id, true
}

func (h *SupportTicketHandler) List(c *gin.Context) {
	userID, ok := ticketUser(c)
	if !ok {
		return
	}
	page, size, ok := ticketrequest.Pagination(c)
	if !ok {
		return
	}
	items, result, err := h.service.ListForUser(c.Request.Context(), userID, pagination.PaginationParams{Page: page, PageSize: size}, c.Query("category"), c.Query("status"), c.Query("search"))
	if response.ErrorFrom(c, err) {
		return
	}
	response.Paginated(c, dto.SupportTicketsFromService(items), result.Total, result.Page, result.PageSize)
}

func (h *SupportTicketHandler) Create(c *gin.Context) {
	userID, ok := ticketUser(c)
	if !ok {
		return
	}
	var req CreateSupportTicketRequest
	if !ticketrequest.Bind(c, &req) {
		return
	}
	ticket, err := h.service.Create(c.Request.Context(), userID, service.CreateTicketInput{Subject: req.Subject, Body: req.Body, Category: req.Category})
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.SupportTicketFromService(ticket))
}

func (h *SupportTicketHandler) Get(c *gin.Context) {
	userID, ok := ticketUser(c)
	if !ok {
		return
	}
	id, ok := ticketID(c)
	if !ok {
		return
	}
	ticket, err := h.service.GetForUser(c.Request.Context(), userID, id)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.SupportTicketFromService(ticket))
}

func (h *SupportTicketHandler) Reply(c *gin.Context) {
	userID, ok := ticketUser(c)
	if !ok {
		return
	}
	id, ok := ticketID(c)
	if !ok {
		return
	}
	var req ReplySupportTicketRequest
	if !ticketrequest.Bind(c, &req) {
		return
	}
	ticket, err := h.service.ReplyAsUser(c.Request.Context(), userID, id, req.Body)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.SupportTicketFromService(ticket))
}

func (h *SupportTicketHandler) Close(c *gin.Context) {
	userID, ok := ticketUser(c)
	if !ok {
		return
	}
	id, ok := ticketID(c)
	if !ok {
		return
	}
	ticket, err := h.service.CloseForUser(c.Request.Context(), userID, id)
	if response.ErrorFrom(c, err) {
		return
	}
	response.Success(c, dto.SupportTicketFromService(ticket))
}
