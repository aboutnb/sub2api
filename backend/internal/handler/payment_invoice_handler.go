package handler

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
)

type validateInvoiceOrdersRequest struct {
	OrderIDs   []int64 `json:"order_ids" binding:"required,min=1,max=20"`
	NeedPayTax bool    `json:"need_pay_tax"`
}

type checkInvoiceTaxPaymentRequest struct {
	TaxOrderNo string `json:"tax_order_no" binding:"required"`
}

func (h *PaymentHandler) GetInvoiceConfig(c *gin.Context) {
	if h.invoiceService == nil {
		response.Success(c, service.InvoiceConfigResponse{})
		return
	}
	config, err := h.invoiceService.Config(c.Request.Context())
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, config)
}

func (h *PaymentHandler) ValidateInvoiceOrders(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	var req validateInvoiceOrdersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.invoiceService.ValidateOrders(c.Request.Context(), subject.UserID, req.OrderIDs, req.NeedPayTax)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *PaymentHandler) GetCurrentInvoiceDraft(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	result, err := h.invoiceService.CurrentDraft(c.Request.Context(), subject.UserID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *PaymentHandler) AbandonInvoiceDraft(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	draftID, ok := parseInvoiceID(c)
	if !ok {
		return
	}
	if err := h.invoiceService.AbandonDraft(c.Request.Context(), subject.UserID, draftID); err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, nil)
}

func (h *PaymentHandler) CheckInvoiceTaxPayment(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	draftID, ok := parseInvoiceID(c)
	if !ok {
		return
	}
	var req checkInvoiceTaxPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.invoiceService.CheckTaxPayment(c.Request.Context(), subject.UserID, draftID, req.TaxOrderNo)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *PaymentHandler) ApplyInvoice(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	draftID, ok := parseInvoiceID(c)
	if !ok {
		return
	}
	var req service.InvoiceApplyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "Invalid request: "+err.Error())
		return
	}
	result, err := h.invoiceService.Apply(c.Request.Context(), subject.UserID, draftID, req)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Created(c, result)
}

func (h *PaymentHandler) ListInvoices(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	page, pageSize := response.ParsePagination(c)
	items, total, err := h.invoiceService.ListApplications(c.Request.Context(), subject.UserID, page, pageSize)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Paginated(c, items, int64(total), page, pageSize)
}

func (h *PaymentHandler) CancelInvoice(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	applicationID, ok := parseInvoiceID(c)
	if !ok {
		return
	}
	result, err := h.invoiceService.Cancel(c.Request.Context(), subject.UserID, applicationID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	response.Success(c, result)
}

func (h *PaymentHandler) DownloadInvoicePDF(c *gin.Context) {
	subject, ok := requireAuth(c)
	if !ok {
		return
	}
	applicationID, ok := parseInvoiceID(c)
	if !ok {
		return
	}
	pdf, err := h.invoiceService.DownloadPDF(c.Request.Context(), subject.UserID, applicationID)
	if err != nil {
		response.ErrorFrom(c, err)
		return
	}
	defer func() { _ = pdf.Body.Close() }()
	contentLength := pdf.ContentLength
	if contentLength < 0 {
		contentLength = -1
	}
	headers := map[string]string{
		"Content-Disposition": fmt.Sprintf("attachment; filename=invoice-%d.pdf", applicationID),
		"Cache-Control":       "private, no-store",
	}
	c.DataFromReader(http.StatusOK, contentLength, pdf.ContentType, pdf.Body, headers)
}

func parseInvoiceID(c *gin.Context) (int64, bool) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil || id <= 0 {
		response.BadRequest(c, "Invalid invoice application ID")
		return 0, false
	}
	return id, true
}
