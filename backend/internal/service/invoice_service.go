package service

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/invoiceapplication"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	invoiceScope            = "invoice.apply"
	invoiceMaxOrders        = 20
	invoiceJSONBodyLimit    = 1 << 20
	invoicePDFBodyLimit     = 32 << 20
	invoiceRemotePageSize   = 100
	invoiceRemoteMaxPages   = 20
	invoiceTaxPollAttempts  = 3
	invoiceStatusDraft      = "draft"
	invoiceStatusAbandoned  = "abandoned"
	invoiceStatusSubmitting = "submitting"
	invoiceStatusUnknown    = "submission_unknown"
)

var invoiceClaimingStatuses = []string{
	invoiceStatusDraft,
	"failed",
	invoiceStatusSubmitting,
	invoiceStatusUnknown,
	"pending",
	"approved",
	"completed",
}

type InvoiceService struct {
	entClient  *dbent.Client
	config     config.InvoiceIntegrationConfig
	httpClient *http.Client

	tokenMu     sync.Mutex
	accessToken string
	tokenExpiry time.Time
	applyMu     sync.Mutex
	draftMu     sync.Mutex
}

type InvoiceConfigResponse struct {
	Enabled            bool `json:"enabled"`
	SupportsTaxPayment bool `json:"supports_tax_payment"`
	MaxOrders          int  `json:"max_orders"`
}

type InvoiceDraftResponse struct {
	DraftID     int64          `json:"draft_id"`
	OrderIDs    []int64        `json:"order_ids"`
	NeedPayTax  bool           `json:"need_pay_tax"`
	Validation  map[string]any `json:"validation"`
	TaxOrderNos []string       `json:"tax_order_nos"`
}

type InvoiceTaxStatusResponse struct {
	Paid        bool           `json:"paid"`
	Ready       bool           `json:"ready"`
	Validation  map[string]any `json:"validation,omitempty"`
	TaxOrderNos []string       `json:"tax_order_nos"`
}

type InvoiceApplyRequest struct {
	BuyerType        string `json:"buyer_type"`
	Title            string `json:"title"`
	TaxpayerID       string `json:"taxpayer_id"`
	BuyerAddress     string `json:"buyer_address"`
	BuyerPhone       string `json:"buyer_phone"`
	BuyerBank        string `json:"buyer_bank"`
	BuyerBankAccount string `json:"buyer_bank_account"`
	RecipientEmail   string `json:"recipient_email"`
}

type InvoiceApplicationResponse struct {
	ID             int64     `json:"id"`
	ExternalID     string    `json:"external_id,omitempty"`
	OrderIDs       []int64   `json:"order_ids"`
	OrderNos       []string  `json:"order_nos"`
	NeedPayTax     bool      `json:"need_pay_tax"`
	TaxOrderNos    []string  `json:"tax_order_nos"`
	Status         string    `json:"status"`
	Title          string    `json:"title,omitempty"`
	RecipientEmail string    `json:"recipient_email,omitempty"`
	TotalAmount    string    `json:"total_amount"`
	Currency       string    `json:"currency"`
	RequestID      string    `json:"request_id,omitempty"`
	ErrorCode      string    `json:"error_code,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

type InvoicePDF struct {
	Body               io.ReadCloser
	ContentLength      int64
	ContentType        string
	ContentDisposition string
}

type invoiceTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int64  `json:"expires_in"`
	Scope       string `json:"scope"`
}

type invoiceEnvelope struct {
	Code      any             `json:"code"`
	Message   string          `json:"message"`
	Data      json.RawMessage `json:"data"`
	RequestID string          `json:"requestId"`
}

type invoiceUpstreamError struct {
	StatusCode int
	Code       string
	Message    string
	RequestID  string
}

func (e *invoiceUpstreamError) Error() string {
	return fmt.Sprintf("invoice upstream error: status=%d code=%s request_id=%s", e.StatusCode, e.Code, e.RequestID)
}

func NewInvoiceService(entClient *dbent.Client, cfg *config.Config) *InvoiceService {
	timeout := time.Duration(cfg.Invoice.TimeoutSeconds) * time.Second
	if timeout <= 0 {
		timeout = 15 * time.Second
	}
	client := &http.Client{
		Timeout: timeout,
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	return &InvoiceService{entClient: entClient, config: cfg.Invoice, httpClient: client}
}

func (s *InvoiceService) Config() InvoiceConfigResponse {
	return InvoiceConfigResponse{
		Enabled:            s.config.Enabled && s.config.ClientID != "" && s.config.ClientSecret != "" && s.config.BaseURL != "",
		SupportsTaxPayment: true,
		MaxOrders:          invoiceMaxOrders,
	}
}

func (s *InvoiceService) ValidateOrders(ctx context.Context, userID int64, orderIDs []int64, needPayTax bool) (*InvoiceDraftResponse, error) {
	if err := s.requireConfigured(); err != nil {
		return nil, err
	}
	s.draftMu.Lock()
	defer s.draftMu.Unlock()

	orders, normalizedIDs, orderNos, err := s.resolveOwnedCompletedOrders(ctx, userID, orderIDs)
	if err != nil {
		return nil, err
	}
	_ = orders
	activeDraft, err := s.currentDraft(ctx, userID)
	if err != nil {
		return nil, err
	}
	if activeDraft != nil {
		if sameInvoiceOrderIDs(activeDraft.OrderIds, normalizedIDs) && activeDraft.NeedPayTax == needPayTax {
			return invoiceDraftResponse(activeDraft), nil
		}
		return nil, infraerrors.Conflict("INVOICE_DRAFT_ACTIVE", "complete or resume the existing invoice draft before starting another")
	}
	if err := s.ensureOrdersNotClaimed(ctx, normalizedIDs, 0); err != nil {
		return nil, err
	}

	validation, err := s.validateRemoteOrders(ctx, orderNos, needPayTax, nil)
	if err != nil {
		return nil, invoicePublicError(err)
	}
	draft, err := s.entClient.InvoiceApplication.Create().
		SetUserID(userID).
		SetOrderIds(normalizedIDs).
		SetOrderNos(orderNos).
		SetNeedPayTax(needPayTax).
		SetValidationSnapshot(validation).
		SetTotalAmount(invoiceAmount(validation)).
		SetCurrency(defaultString(mapString(validation, "currency"), "CNY")).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("create invoice draft: %w", err)
	}
	return invoiceDraftResponse(draft), nil
}

// CurrentDraft returns the sole in-progress draft so a browser reload can resume it.
func (s *InvoiceService) CurrentDraft(ctx context.Context, userID int64) (*InvoiceDraftResponse, error) {
	if err := s.requireConfigured(); err != nil {
		return nil, err
	}
	draft, err := s.currentDraft(ctx, userID)
	if err != nil || draft == nil {
		return nil, err
	}
	return invoiceDraftResponse(draft), nil
}

// AbandonDraft only releases a no-tax draft. A tax checkout can have been paid
// outside the browser, so it must remain recoverable until submitted.
func (s *InvoiceService) AbandonDraft(ctx context.Context, userID, draftID int64) error {
	if err := s.requireConfigured(); err != nil {
		return err
	}
	draft, err := s.getOwnedApplication(ctx, userID, draftID)
	if err != nil {
		return err
	}
	if draft.Status != invoiceStatusDraft && draft.Status != "failed" {
		return infraerrors.Conflict("INVOICE_DRAFT_NOT_ACTIVE", "invoice draft is no longer active")
	}
	if draft.NeedPayTax {
		return infraerrors.Conflict("INVOICE_TAX_DRAFT_MUST_RESUME", "a tax invoice draft must be resumed to preserve any tax payment")
	}
	if _, err := draft.Update().SetStatus(invoiceStatusAbandoned).Save(ctx); err != nil {
		return fmt.Errorf("abandon invoice draft: %w", err)
	}
	return nil
}

func (s *InvoiceService) CheckTaxPayment(ctx context.Context, userID, draftID int64, taxOrderNo string) (*InvoiceTaxStatusResponse, error) {
	if err := s.requireConfigured(); err != nil {
		return nil, err
	}
	draft, err := s.getOwnedApplication(ctx, userID, draftID)
	if err != nil {
		return nil, err
	}
	if !draft.NeedPayTax || draft.Status != invoiceStatusDraft {
		return nil, infraerrors.BadRequest("INVOICE_TAX_PAYMENT_NOT_REQUIRED", "invoice tax payment is not required")
	}
	taxOrderNo = strings.TrimSpace(taxOrderNo)
	if !taxOrderBelongsToValidation(draft.ValidationSnapshot, taxOrderNo) && !containsString(draft.TaxOrderNos, taxOrderNo) {
		return nil, infraerrors.BadRequest("INVOICE_INVALID_TAX_ORDERS", "tax order does not belong to this invoice draft")
	}

	_, paid, err := s.pollRemoteTaxPaymentStatus(ctx, taxOrderNo)
	if err != nil {
		return nil, invoicePublicError(err)
	}
	if !paid {
		return &InvoiceTaxStatusResponse{Paid: false, Ready: false, TaxOrderNos: append([]string(nil), draft.TaxOrderNos...)}, nil
	}

	taxOrderNos := appendUniqueString(draft.TaxOrderNos, taxOrderNo)
	validation, err := s.validateRemoteOrders(ctx, draft.OrderNos, true, taxOrderNos)
	if err != nil {
		return nil, invoicePublicError(err)
	}
	ready := mapString(validation, "taxDueAmount") == "0.00"
	_, err = draft.Update().
		SetTaxOrderNos(taxOrderNos).
		SetValidationSnapshot(validation).
		SetTotalAmount(invoiceAmount(validation)).
		SetCurrency(defaultString(mapString(validation, "currency"), draft.Currency)).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("update invoice tax reconciliation: %w", err)
	}
	return &InvoiceTaxStatusResponse{Paid: true, Ready: ready, Validation: validation, TaxOrderNos: taxOrderNos}, nil
}

func (s *InvoiceService) Apply(ctx context.Context, userID, draftID int64, input InvoiceApplyRequest) (*InvoiceApplicationResponse, error) {
	if err := s.requireConfigured(); err != nil {
		return nil, err
	}
	input = normalizeInvoiceApplyRequest(input)
	if err := validateInvoiceApplyRequest(input); err != nil {
		return nil, err
	}

	s.applyMu.Lock()
	defer s.applyMu.Unlock()

	draft, err := s.getOwnedApplication(ctx, userID, draftID)
	if err != nil {
		return nil, err
	}
	if draft.ExternalID != nil && *draft.ExternalID != "" {
		return invoiceApplicationResponse(draft), nil
	}
	if draft.Status != invoiceStatusDraft && draft.Status != "failed" {
		return nil, infraerrors.Conflict("INVOICE_SUBMISSION_IN_PROGRESS", "invoice application is already being submitted")
	}
	if err := s.ensureOrdersNotClaimed(ctx, draft.OrderIds, draft.ID); err != nil {
		return nil, err
	}

	validation, err := s.validateRemoteOrders(ctx, draft.OrderNos, draft.NeedPayTax, draft.TaxOrderNos)
	if err != nil {
		return nil, invoicePublicError(err)
	}
	if draft.NeedPayTax && mapString(validation, "taxDueAmount") != "0.00" {
		return nil, infraerrors.Conflict("INVOICE_TAX_ORDER_REQUIRED", "invoice tax payment has not been fully reconciled")
	}

	draft, err = draft.Update().
		SetStatus(invoiceStatusSubmitting).
		SetTitle(input.Title).
		SetRecipientEmail(input.RecipientEmail).
		SetValidationSnapshot(validation).
		SetTotalAmount(invoiceAmount(validation)).
		SetCurrency(defaultString(mapString(validation, "currency"), draft.Currency)).
		ClearErrorCode().ClearErrorMessage().ClearRequestID().
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("mark invoice application submitting: %w", err)
	}

	payload := map[string]any{
		"orderNos":       draft.OrderNos,
		"buyerType":      input.BuyerType,
		"title":          input.Title,
		"taxpayerId":     input.TaxpayerID,
		"recipientEmail": input.RecipientEmail,
	}
	setNonEmpty(payload, "buyerAddress", input.BuyerAddress)
	setNonEmpty(payload, "buyerPhone", input.BuyerPhone)
	setNonEmpty(payload, "buyerBank", input.BuyerBank)
	setNonEmpty(payload, "buyerBankAccount", input.BuyerBankAccount)
	if draft.NeedPayTax {
		payload["needPayTax"] = true
		payload["taxOrderNos"] = draft.TaxOrderNos
	}

	remote, err := s.createRemoteApplication(ctx, payload)
	if err != nil {
		status := "failed"
		var upstream *invoiceUpstreamError
		if !errors.As(err, &upstream) {
			status = invoiceStatusUnknown
		}
		update := draft.Update().SetStatus(status)
		if upstream != nil {
			update.SetNillableErrorCode(nonEmptyStringPtr(upstream.Code)).
				SetNillableErrorMessage(nonEmptyStringPtr(upstream.Message)).
				SetNillableRequestID(nonEmptyStringPtr(upstream.RequestID))
		} else {
			update.SetErrorCode("INVOICE_SUBMISSION_UNKNOWN").SetErrorMessage("invoice submission result is unknown")
		}
		_, _ = update.Save(ctx)
		return nil, invoicePublicError(err)
	}

	externalID := mapIdentifier(remote, "id")
	if externalID == "" {
		_, _ = draft.Update().SetStatus(invoiceStatusUnknown).SetErrorCode("INVOICE_INVALID_RESPONSE").SetErrorMessage("invoice application ID is missing").Save(ctx)
		return nil, infraerrors.ServiceUnavailable("INVOICE_INVALID_RESPONSE", "invoice service returned an invalid response")
	}
	status := defaultString(strings.ToLower(mapString(remote, "status")), "pending")
	draft, err = draft.Update().
		SetExternalID(externalID).
		SetStatus(status).
		SetExternalSnapshot(remote).
		ClearErrorCode().ClearErrorMessage().ClearRequestID().
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("save invoice application result: %w", err)
	}
	return invoiceApplicationResponse(draft), nil
}

func (s *InvoiceService) ListApplications(ctx context.Context, userID int64, page, pageSize int) ([]InvoiceApplicationResponse, int, error) {
	if err := s.requireConfigured(); err != nil {
		return nil, 0, err
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}
	query := s.entClient.InvoiceApplication.Query().Where(
		invoiceapplication.UserIDEQ(userID),
		invoiceapplication.Or(
			invoiceapplication.ExternalIDNotNil(),
			invoiceapplication.StatusIn(invoiceStatusUnknown, invoiceStatusSubmitting, "failed"),
		),
	)
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("count invoice applications: %w", err)
	}
	items, err := query.Order(dbent.Desc(invoiceapplication.FieldCreatedAt)).Limit(pageSize).Offset((page - 1) * pageSize).All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list invoice applications: %w", err)
	}
	// A remote status refresh is best-effort. Locally persisted ownership and the
	// last known state remain useful when XZNOAuth is temporarily unavailable.
	_ = s.syncRemoteApplications(ctx, items)
	result := make([]InvoiceApplicationResponse, 0, len(items))
	for _, item := range items {
		result = append(result, *invoiceApplicationResponse(item))
	}
	return result, total, nil
}

func (s *InvoiceService) Cancel(ctx context.Context, userID, applicationID int64) (*InvoiceApplicationResponse, error) {
	if err := s.requireConfigured(); err != nil {
		return nil, err
	}
	application, err := s.getOwnedApplication(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}
	if application.ExternalID == nil || *application.ExternalID == "" {
		return nil, infraerrors.Conflict("INVOICE_CANNOT_CANCEL", "invoice application has no confirmed remote record")
	}
	if application.Status != "pending" && application.Status != "approved" {
		return nil, infraerrors.Conflict("INVOICE_CANNOT_CANCEL", "invoice application cannot be canceled in its current status")
	}
	remote, err := s.cancelRemoteApplication(ctx, *application.ExternalID)
	if err != nil {
		return nil, invoicePublicError(err)
	}
	status := defaultString(strings.ToLower(mapString(remote, "status")), "canceled")
	application, err = application.Update().SetStatus(status).SetExternalSnapshot(remote).Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("save canceled invoice application: %w", err)
	}
	return invoiceApplicationResponse(application), nil
}

func (s *InvoiceService) DownloadPDF(ctx context.Context, userID, applicationID int64) (*InvoicePDF, error) {
	if err := s.requireConfigured(); err != nil {
		return nil, err
	}
	application, err := s.getOwnedApplication(ctx, userID, applicationID)
	if err != nil {
		return nil, err
	}
	if application.ExternalID == nil || *application.ExternalID == "" || application.Status != "completed" {
		return nil, infraerrors.Conflict("INVOICE_PDF_NOT_READY", "invoice PDF is not ready")
	}
	return s.downloadRemotePDF(ctx, *application.ExternalID)
}

func (s *InvoiceService) requireConfigured() error {
	if !s.Config().Enabled {
		return infraerrors.ServiceUnavailable("INVOICE_NOT_CONFIGURED", "invoice service is not configured")
	}
	return nil
}

func (s *InvoiceService) resolveOwnedCompletedOrders(ctx context.Context, userID int64, orderIDs []int64) ([]*dbent.PaymentOrder, []int64, []string, error) {
	normalized, err := normalizeInvoiceOrderIDs(orderIDs)
	if err != nil {
		return nil, nil, nil, err
	}
	orders, err := s.entClient.PaymentOrder.Query().Where(
		paymentorder.UserIDEQ(userID),
		paymentorder.IDIn(normalized...),
		paymentorder.StatusEQ(payment.OrderStatusCompleted),
	).All(ctx)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("query invoice orders: %w", err)
	}
	if len(orders) != len(normalized) {
		return nil, nil, nil, infraerrors.BadRequest("INVOICE_ORDER_NOT_ELIGIBLE", "one or more orders are not eligible for invoicing")
	}
	byID := make(map[int64]*dbent.PaymentOrder, len(orders))
	for _, order := range orders {
		byID[order.ID] = order
	}
	ordered := make([]*dbent.PaymentOrder, 0, len(normalized))
	orderNos := make([]string, 0, len(normalized))
	for _, id := range normalized {
		order := byID[id]
		orderNo := strings.TrimSpace(order.OutTradeNo)
		if orderNo == "" {
			orderNo = strings.TrimSpace(order.PaymentTradeNo)
		}
		if orderNo == "" {
			return nil, nil, nil, infraerrors.BadRequest("INVOICE_ORDER_NOT_ELIGIBLE", "an order is missing a verifiable payment number")
		}
		ordered = append(ordered, order)
		orderNos = append(orderNos, orderNo)
	}
	return ordered, normalized, orderNos, nil
}

func (s *InvoiceService) ensureOrdersNotClaimed(ctx context.Context, orderIDs []int64, excludeID int64) error {
	query := s.entClient.InvoiceApplication.Query().Where(invoiceapplication.StatusIn(invoiceClaimingStatuses...))
	if excludeID > 0 {
		query = query.Where(invoiceapplication.IDNEQ(excludeID))
	}
	applications, err := query.All(ctx)
	if err != nil {
		return fmt.Errorf("query claimed invoice orders: %w", err)
	}
	wanted := make(map[int64]struct{}, len(orderIDs))
	for _, id := range orderIDs {
		wanted[id] = struct{}{}
	}
	for _, application := range applications {
		for _, id := range application.OrderIds {
			if _, ok := wanted[id]; ok {
				return infraerrors.Conflict("INVOICE_ORDER_ALREADY_CLAIMED", "one or more orders already belong to an invoice application")
			}
		}
	}
	return nil
}

func (s *InvoiceService) getOwnedApplication(ctx context.Context, userID, id int64) (*dbent.InvoiceApplication, error) {
	application, err := s.entClient.InvoiceApplication.Query().Where(
		invoiceapplication.IDEQ(id),
		invoiceapplication.UserIDEQ(userID),
	).Only(ctx)
	if err != nil {
		if dbent.IsNotFound(err) {
			return nil, infraerrors.NotFound("NOT_FOUND", "invoice application not found")
		}
		return nil, fmt.Errorf("query invoice application: %w", err)
	}
	return application, nil
}

func (s *InvoiceService) currentDraft(ctx context.Context, userID int64) (*dbent.InvoiceApplication, error) {
	draft, err := s.entClient.InvoiceApplication.Query().Where(
		invoiceapplication.UserIDEQ(userID),
		invoiceapplication.StatusIn(invoiceStatusDraft, "failed"),
	).Order(dbent.Desc(invoiceapplication.FieldUpdatedAt)).First(ctx)
	if err == nil {
		return draft, nil
	}
	if dbent.IsNotFound(err) {
		return nil, nil
	}
	return nil, fmt.Errorf("query active invoice draft: %w", err)
}

func (s *InvoiceService) validateRemoteOrders(ctx context.Context, orderNos []string, needPayTax bool, taxOrderNos []string) (map[string]any, error) {
	payload := map[string]any{"orderNos": orderNos}
	if needPayTax {
		payload["needPayTax"] = true
		if len(taxOrderNos) > 0 {
			payload["taxOrderNos"] = taxOrderNos
		}
	}
	var result map[string]any
	if err := s.doJSON(ctx, http.MethodPost, "/api/v1/invoice-orders/validate", payload, &result); err != nil {
		return nil, err
	}
	if err := setInvoiceAmount(result, needPayTax); err != nil {
		return nil, err
	}
	return result, nil
}

func setInvoiceAmount(validation map[string]any, needPayTax bool) error {
	if validation == nil {
		return &invoiceUpstreamError{StatusCode: http.StatusBadGateway, Code: "INVOICE_INVALID_RESPONSE", Message: "invoice validation response is empty"}
	}
	orderAmount, err := parseInvoiceMoney(validation, "totalAmount")
	if err != nil {
		return err
	}
	if !needPayTax {
		// Merchant-side fees must never change the amount shown or submitted as the user's invoice total.
		validation["invoiceAmount"] = orderAmount.StringFixed(2)
		return nil
	}

	taxAmount, err := parseInvoiceMoney(validation, "taxAmount")
	if err != nil {
		return err
	}
	taxPaidAmount, err := parseInvoiceMoney(validation, "taxPaidAmount")
	if err != nil {
		return err
	}
	taxDueAmount, err := parseInvoiceMoney(validation, "taxDueAmount")
	if err != nil {
		return err
	}
	if taxPaidAmount.GreaterThan(taxAmount) || !taxDueAmount.Equal(taxAmount.Sub(taxPaidAmount)) {
		return &invoiceUpstreamError{StatusCode: http.StatusBadGateway, Code: "INVOICE_INVALID_RESPONSE", Message: "invoice tax amounts are inconsistent"}
	}

	expectedInvoiceAmount := orderAmount.Add(taxAmount)
	if provided := mapString(validation, "invoiceAmount"); provided != "" {
		providedAmount, err := decimal.NewFromString(provided)
		if err != nil || !providedAmount.Equal(expectedInvoiceAmount) {
			return &invoiceUpstreamError{StatusCode: http.StatusBadGateway, Code: "INVOICE_INVALID_RESPONSE", Message: "invoice amount is inconsistent"}
		}
	}
	validation["taxAmount"] = taxAmount.StringFixed(2)
	validation["taxPaidAmount"] = taxPaidAmount.StringFixed(2)
	validation["taxDueAmount"] = taxDueAmount.StringFixed(2)
	validation["invoiceAmount"] = expectedInvoiceAmount.StringFixed(2)
	return nil
}

func parseInvoiceMoney(validation map[string]any, key string) (decimal.Decimal, error) {
	amount, err := decimal.NewFromString(mapString(validation, key))
	if err != nil || amount.IsNegative() {
		return decimal.Zero, &invoiceUpstreamError{StatusCode: http.StatusBadGateway, Code: "INVOICE_INVALID_RESPONSE", Message: "invoice " + key + " is invalid"}
	}
	return amount, nil
}

func invoiceAmount(validation map[string]any) string {
	return defaultString(mapString(validation, "invoiceAmount"), mapString(validation, "totalAmount"))
}

func (s *InvoiceService) checkRemoteTaxStatus(ctx context.Context, taxOrderNo string) (map[string]any, error) {
	var result map[string]any
	if err := s.doJSON(ctx, http.MethodPost, "/api/v1/invoice-tax-payments/status", map[string]string{"taxOrderNo": taxOrderNo}, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *InvoiceService) pollRemoteTaxPaymentStatus(ctx context.Context, taxOrderNo string) (map[string]any, bool, error) {
	var status map[string]any
	for attempt := 0; attempt < invoiceTaxPollAttempts; attempt++ {
		var err error
		status, err = s.checkRemoteTaxStatus(ctx, taxOrderNo)
		if err != nil {
			return nil, false, err
		}
		if mapBool(status, "paid") {
			return status, true, nil
		}
		if attempt == invoiceTaxPollAttempts-1 {
			break
		}
		backoff := time.Duration(1<<uint(attempt)) * time.Second
		timer := time.NewTimer(backoff)
		select {
		case <-ctx.Done():
			if !timer.Stop() {
				<-timer.C
			}
			return nil, false, ctx.Err()
		case <-timer.C:
		}
	}
	return status, false, nil
}

func (s *InvoiceService) createRemoteApplication(ctx context.Context, payload map[string]any) (map[string]any, error) {
	var result map[string]any
	if err := s.doJSON(ctx, http.MethodPost, "/api/v1/invoices", payload, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *InvoiceService) cancelRemoteApplication(ctx context.Context, externalID string) (map[string]any, error) {
	var result map[string]any
	path := "/api/v1/invoices/" + url.PathEscape(externalID) + "/cancel"
	if err := s.doJSON(ctx, http.MethodPost, path, nil, &result); err != nil {
		return nil, err
	}
	return result, nil
}

func (s *InvoiceService) syncRemoteApplications(ctx context.Context, local []*dbent.InvoiceApplication) error {
	wanted := make(map[string]*dbent.InvoiceApplication)
	unconfirmed := make(map[string]*dbent.InvoiceApplication)
	for _, application := range local {
		if application.ExternalID != nil && *application.ExternalID != "" {
			wanted[*application.ExternalID] = application
		} else if application.Status == invoiceStatusUnknown || application.Status == invoiceStatusSubmitting {
			key := invoiceOrderSetKey(application.OrderNos)
			if key != "" {
				unconfirmed[key] = application
			}
		}
	}
	if len(wanted) == 0 && len(unconfirmed) == 0 {
		return nil
	}
	for page := 1; page <= invoiceRemoteMaxPages && (len(wanted) > 0 || len(unconfirmed) > 0); page++ {
		var data struct {
			Items    []map[string]any `json:"items"`
			Total    int              `json:"total"`
			Page     int              `json:"page"`
			PageSize int              `json:"pageSize"`
		}
		path := fmt.Sprintf("/api/v1/invoices?page=%d&pageSize=%d", page, invoiceRemotePageSize)
		if err := s.doJSON(ctx, http.MethodGet, path, nil, &data); err != nil {
			return err
		}
		for _, remote := range data.Items {
			externalID := mapIdentifier(remote, "id")
			application, ok := wanted[externalID]
			if !ok {
				application, ok = unconfirmed[invoiceOrderSetKey(mapStringSlice(remote, "orderNos"))]
				if !ok || externalID == "" {
					continue
				}
			}
			status := strings.ToLower(mapString(remote, "status"))
			update := application.Update().SetExternalSnapshot(remote)
			if application.ExternalID == nil || *application.ExternalID == "" {
				update.SetExternalID(externalID)
			}
			if status != "" {
				update.SetStatus(status)
			}
			if updated, err := update.Save(ctx); err == nil {
				application.ExternalID = updated.ExternalID
				application.Status = updated.Status
				application.ExternalSnapshot = updated.ExternalSnapshot
				application.UpdatedAt = updated.UpdatedAt
			}
			delete(wanted, externalID)
			delete(unconfirmed, invoiceOrderSetKey(application.OrderNos))
		}
		if len(data.Items) == 0 || (data.Total > 0 && page*invoiceRemotePageSize >= data.Total) {
			break
		}
	}
	return nil
}

func (s *InvoiceService) downloadRemotePDF(ctx context.Context, externalID string) (*InvoicePDF, error) {
	path := "/api/v1/invoices/" + url.PathEscape(externalID) + "/pdf"
	for attempt := 0; attempt < 2; attempt++ {
		token, err := s.getAccessToken(ctx, attempt > 0)
		if err != nil {
			return nil, err
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.endpoint(path), nil)
		if err != nil {
			return nil, fmt.Errorf("create invoice PDF request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/pdf")
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return nil, fmt.Errorf("download invoice PDF: %w", err)
		}
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			if attempt == 0 {
				_ = resp.Body.Close()
				s.invalidateToken(token)
				continue
			}
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			err := decodeInvoiceUpstreamError(resp)
			_ = resp.Body.Close()
			return nil, err
		}
		if resp.ContentLength > invoicePDFBodyLimit {
			_ = resp.Body.Close()
			return nil, infraerrors.ServiceUnavailable("INVOICE_PDF_TOO_LARGE", "invoice PDF exceeds the download limit")
		}
		contentType := strings.TrimSpace(strings.Split(resp.Header.Get("Content-Type"), ";")[0])
		if !strings.EqualFold(contentType, "application/pdf") {
			_ = resp.Body.Close()
			return nil, infraerrors.ServiceUnavailable("INVOICE_INVALID_RESPONSE", "invoice service returned a non-PDF response")
		}
		return &InvoicePDF{
			Body:               &limitedReadCloser{Reader: io.LimitReader(resp.Body, invoicePDFBodyLimit+1), Closer: resp.Body},
			ContentLength:      resp.ContentLength,
			ContentType:        resp.Header.Get("Content-Type"),
			ContentDisposition: resp.Header.Get("Content-Disposition"),
		}, nil
	}
	return nil, infraerrors.ServiceUnavailable("INVOICE_DOWNLOAD_FAILED", "invoice PDF download failed")
}

func (s *InvoiceService) doJSON(ctx context.Context, method, path string, payload any, out any) error {
	var encoded []byte
	var err error
	if payload != nil {
		encoded, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("encode invoice request: %w", err)
		}
	}
	for attempt := 0; attempt < 2; attempt++ {
		token, err := s.getAccessToken(ctx, attempt > 0)
		if err != nil {
			return err
		}
		var body io.Reader
		if encoded != nil {
			body = bytes.NewReader(encoded)
		}
		req, err := http.NewRequestWithContext(ctx, method, s.endpoint(path), body)
		if err != nil {
			return fmt.Errorf("create invoice request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Accept", "application/json")
		if encoded != nil {
			req.Header.Set("Content-Type", "application/json")
		}
		resp, err := s.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("invoice request failed: %w", err)
		}
		result, err := decodeInvoiceEnvelope(resp, out)
		_ = resp.Body.Close()
		if err == nil {
			return nil
		}
		if attempt == 0 && isInvoiceAuthFailure(resp.StatusCode, result) {
			s.invalidateToken(token)
			continue
		}
		return err
	}
	return infraerrors.ServiceUnavailable("INVOICE_AUTH_FAILED", "invoice service authentication failed")
}

func (s *InvoiceService) getAccessToken(ctx context.Context, force bool) (string, error) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()
	if !force && s.accessToken != "" && time.Now().Before(s.tokenExpiry) {
		return s.accessToken, nil
	}
	form := url.Values{
		"grant_type":    {"client_credentials"},
		"client_id":     {s.config.ClientID},
		"client_secret": {s.config.ClientSecret},
		"scope":         {invoiceScope},
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.endpoint("/api/oauth/token"), strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("create invoice token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request invoice token: %w", err)
	}
	defer resp.Body.Close()
	body, err := readLimitedBody(resp.Body, invoiceJSONBodyLimit)
	if err != nil {
		return "", fmt.Errorf("read invoice token response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return "", decodeInvoiceErrorBody(resp.StatusCode, body)
	}
	var token invoiceTokenResponse
	if err := json.Unmarshal(body, &token); err != nil || token.AccessToken == "" || token.ExpiresIn <= 0 {
		return "", &invoiceUpstreamError{StatusCode: http.StatusBadGateway, Code: "OAUTH_INVALID_RESPONSE", Message: "invoice token response is invalid"}
	}
	refreshSkew := time.Duration(token.ExpiresIn/10) * time.Second
	if refreshSkew < 5*time.Second {
		refreshSkew = 5 * time.Second
	}
	if refreshSkew > 60*time.Second {
		refreshSkew = 60 * time.Second
	}
	s.accessToken = token.AccessToken
	s.tokenExpiry = time.Now().Add(time.Duration(token.ExpiresIn)*time.Second - refreshSkew)
	return s.accessToken, nil
}

func (s *InvoiceService) invalidateToken(token string) {
	s.tokenMu.Lock()
	defer s.tokenMu.Unlock()
	if s.accessToken == token {
		s.accessToken = ""
		s.tokenExpiry = time.Time{}
	}
}

func (s *InvoiceService) endpoint(path string) string {
	return strings.TrimRight(s.config.BaseURL, "/") + path
}

func decodeInvoiceEnvelope(resp *http.Response, out any) (*invoiceUpstreamError, error) {
	body, err := readLimitedBody(resp.Body, invoiceJSONBodyLimit)
	if err != nil {
		return nil, fmt.Errorf("read invoice response: %w", err)
	}
	var envelope invoiceEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		upstream := &invoiceUpstreamError{StatusCode: resp.StatusCode, Code: "INVOICE_INVALID_RESPONSE", Message: "invoice service returned invalid JSON"}
		return upstream, upstream
	}
	code := fmt.Sprint(envelope.Code)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 || code != "0" {
		upstream := &invoiceUpstreamError{StatusCode: resp.StatusCode, Code: code, Message: envelope.Message, RequestID: envelope.RequestID}
		if upstream.Code == "<nil>" || upstream.Code == "" {
			upstream.Code = "INVOICE_INVALID_RESPONSE"
		}
		return upstream, upstream
	}
	if out != nil && len(envelope.Data) > 0 && string(envelope.Data) != "null" {
		if err := json.Unmarshal(envelope.Data, out); err != nil {
			upstream := &invoiceUpstreamError{StatusCode: http.StatusBadGateway, Code: "INVOICE_INVALID_RESPONSE", Message: "invoice service response data is invalid", RequestID: envelope.RequestID}
			return upstream, upstream
		}
	}
	return nil, nil
}

func decodeInvoiceUpstreamError(resp *http.Response) error {
	body, err := readLimitedBody(resp.Body, invoiceJSONBodyLimit)
	if err != nil {
		return fmt.Errorf("read invoice error response: %w", err)
	}
	return decodeInvoiceErrorBody(resp.StatusCode, body)
}

func decodeInvoiceErrorBody(statusCode int, body []byte) error {
	var envelope invoiceEnvelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return &invoiceUpstreamError{StatusCode: statusCode, Code: "INVOICE_UPSTREAM_ERROR", Message: "invoice service request failed"}
	}
	code := fmt.Sprint(envelope.Code)
	if code == "<nil>" || code == "" {
		code = "INVOICE_UPSTREAM_ERROR"
	}
	return &invoiceUpstreamError{StatusCode: statusCode, Code: code, Message: envelope.Message, RequestID: envelope.RequestID}
}

func isInvoiceAuthFailure(status int, upstream *invoiceUpstreamError) bool {
	if status == http.StatusUnauthorized || status == http.StatusForbidden {
		return true
	}
	if upstream == nil {
		return false
	}
	switch upstream.Code {
	case "AUTH_INVALID_TOKEN", "AUTH_INSUFFICIENT_SCOPE", "AUTH_FORBIDDEN":
		return true
	default:
		return false
	}
}

func invoicePublicError(err error) error {
	var upstream *invoiceUpstreamError
	if !errors.As(err, &upstream) {
		return infraerrors.ServiceUnavailable("INVOICE_SERVICE_UNAVAILABLE", "invoice service is temporarily unavailable").WithCause(err)
	}
	status := upstream.StatusCode
	switch {
	case strings.HasPrefix(upstream.Code, "OAUTH_") || strings.HasPrefix(upstream.Code, "AUTH_"):
		status = http.StatusServiceUnavailable
	case status == http.StatusTooManyRequests:
		status = http.StatusTooManyRequests
	case status >= 500 || status <= 0:
		status = http.StatusBadGateway
	case status == http.StatusUnauthorized || status == http.StatusForbidden:
		status = http.StatusServiceUnavailable
	case status < 400:
		status = http.StatusBadGateway
	}
	metadata := map[string]string{}
	if upstream.RequestID != "" {
		metadata["request_id"] = upstream.RequestID
	}
	return infraerrors.New(status, upstream.Code, "invoice service request failed").WithMetadata(metadata).WithCause(err)
}

func normalizeInvoiceOrderIDs(orderIDs []int64) ([]int64, error) {
	if len(orderIDs) == 0 || len(orderIDs) > invoiceMaxOrders {
		return nil, infraerrors.BadRequest("INVOICE_INVALID_ORDERS", "select between 1 and 20 orders")
	}
	seen := make(map[int64]struct{}, len(orderIDs))
	result := make([]int64, 0, len(orderIDs))
	for _, id := range orderIDs {
		if id <= 0 {
			return nil, infraerrors.BadRequest("INVOICE_INVALID_ORDERS", "order IDs must be positive")
		}
		if _, ok := seen[id]; ok {
			return nil, infraerrors.BadRequest("INVOICE_DUPLICATE_ORDER", "duplicate order selected")
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	sort.Slice(result, func(i, j int) bool { return result[i] < result[j] })
	return result, nil
}

func sameInvoiceOrderIDs(left, right []int64) bool {
	if len(left) != len(right) {
		return false
	}
	left = append([]int64(nil), left...)
	right = append([]int64(nil), right...)
	sort.Slice(left, func(i, j int) bool { return left[i] < left[j] })
	sort.Slice(right, func(i, j int) bool { return right[i] < right[j] })
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func normalizeInvoiceApplyRequest(input InvoiceApplyRequest) InvoiceApplyRequest {
	input.BuyerType = strings.ToLower(strings.TrimSpace(input.BuyerType))
	input.Title = strings.TrimSpace(input.Title)
	input.TaxpayerID = strings.TrimSpace(input.TaxpayerID)
	input.BuyerAddress = strings.TrimSpace(input.BuyerAddress)
	input.BuyerPhone = strings.TrimSpace(input.BuyerPhone)
	input.BuyerBank = strings.TrimSpace(input.BuyerBank)
	input.BuyerBankAccount = strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(input.BuyerBankAccount))
	input.RecipientEmail = strings.TrimSpace(input.RecipientEmail)
	return input
}

func validateInvoiceApplyRequest(input InvoiceApplyRequest) error {
	if input.BuyerType != "individual" && input.BuyerType != "company" {
		return infraerrors.BadRequest("INVOICE_INVALID_FORM", "buyer type must be individual or company")
	}
	if input.Title == "" || len([]rune(input.Title)) > 255 {
		return infraerrors.BadRequest("INVOICE_INVALID_FORM", "invoice title is required and must not exceed 255 characters")
	}
	if input.TaxpayerID == "" || len(input.TaxpayerID) > 32 {
		return infraerrors.BadRequest("INVOICE_INVALID_FORM", "taxpayer ID is required and must not exceed 32 characters")
	}
	if len([]rune(input.BuyerAddress)) > 255 || len([]rune(input.BuyerBank)) > 255 {
		return infraerrors.BadRequest("INVOICE_INVALID_FORM", "invoice address or bank name is too long")
	}
	if len([]rune(input.BuyerPhone)) > 64 || len([]rune(input.RecipientEmail)) > 255 {
		return infraerrors.BadRequest("INVOICE_INVALID_FORM", "invoice phone or recipient email is too long")
	}
	if input.BuyerBankAccount != "" {
		if len(input.BuyerBankAccount) < 8 || len(input.BuyerBankAccount) > 32 || !allDigits(input.BuyerBankAccount) {
			return infraerrors.BadRequest("INVOICE_INVALID_FORM", "bank account must contain 8 to 32 digits")
		}
	}
	parsed, err := mail.ParseAddress(input.RecipientEmail)
	if err != nil || !strings.EqualFold(parsed.Address, input.RecipientEmail) {
		return infraerrors.BadRequest("INVOICE_INVALID_FORM", "recipient email is invalid")
	}
	return nil
}

func invoiceDraftResponse(draft *dbent.InvoiceApplication) *InvoiceDraftResponse {
	return &InvoiceDraftResponse{
		DraftID: draft.ID, OrderIDs: append([]int64(nil), draft.OrderIds...),
		NeedPayTax: draft.NeedPayTax, Validation: draft.ValidationSnapshot,
		TaxOrderNos: append([]string(nil), draft.TaxOrderNos...),
	}
}

func invoiceApplicationResponse(application *dbent.InvoiceApplication) *InvoiceApplicationResponse {
	return &InvoiceApplicationResponse{
		ID: application.ID, ExternalID: invoiceStringValue(application.ExternalID),
		OrderIDs: append([]int64(nil), application.OrderIds...), OrderNos: append([]string(nil), application.OrderNos...),
		NeedPayTax: application.NeedPayTax, TaxOrderNos: append([]string(nil), application.TaxOrderNos...),
		Status: application.Status, Title: invoiceStringValue(application.Title), RecipientEmail: invoiceStringValue(application.RecipientEmail),
		TotalAmount: application.TotalAmount, Currency: application.Currency, RequestID: invoiceStringValue(application.RequestID),
		ErrorCode: invoiceStringValue(application.ErrorCode), CreatedAt: application.CreatedAt, UpdatedAt: application.UpdatedAt,
	}
}

func taxOrderBelongsToValidation(validation map[string]any, taxOrderNo string) bool {
	if taxOrderNo == "" {
		return false
	}
	if mapString(validation, "taxOrderNo") == taxOrderNo {
		return true
	}
	payments, ok := validation["taxPayments"].(map[string]any)
	if !ok {
		return false
	}
	for _, raw := range payments {
		paymentData, ok := raw.(map[string]any)
		if ok && mapString(paymentData, "taxOrderNo") == taxOrderNo {
			return true
		}
	}
	return false
}

func mapString(values map[string]any, key string) string {
	if values == nil {
		return ""
	}
	value, ok := values[key]
	if !ok || value == nil {
		return ""
	}
	if text, ok := value.(string); ok {
		return text
	}
	return fmt.Sprint(value)
}

func mapIdentifier(values map[string]any, key string) string {
	value := mapString(values, key)
	return strings.TrimSuffix(value, ".0")
}

func mapStringSlice(values map[string]any, key string) []string {
	raw, ok := values[key]
	if !ok || raw == nil {
		return nil
	}
	switch typed := raw.(type) {
	case []string:
		return append([]string(nil), typed...)
	case []any:
		result := make([]string, 0, len(typed))
		for _, value := range typed {
			text := strings.TrimSpace(fmt.Sprint(value))
			if text != "" {
				result = append(result, text)
			}
		}
		return result
	default:
		return nil
	}
}

func invoiceOrderSetKey(orderNos []string) string {
	if len(orderNos) == 0 {
		return ""
	}
	normalized := make([]string, 0, len(orderNos))
	for _, orderNo := range orderNos {
		orderNo = strings.TrimSpace(orderNo)
		if orderNo == "" {
			return ""
		}
		normalized = append(normalized, orderNo)
	}
	sort.Strings(normalized)
	return strings.Join(normalized, "\x00")
}

func mapBool(values map[string]any, key string) bool {
	value, ok := values[key]
	if !ok {
		return false
	}
	switch typed := value.(type) {
	case bool:
		return typed
	case string:
		parsed, _ := strconv.ParseBool(typed)
		return parsed
	default:
		return false
	}
}

func setNonEmpty(values map[string]any, key, value string) {
	if value != "" {
		values[key] = value
	}
}

func defaultString(value, fallback string) string {
	if strings.TrimSpace(value) == "" {
		return fallback
	}
	return value
}

func invoiceStringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func nonEmptyStringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}

func appendUniqueString(values []string, value string) []string {
	result := append([]string(nil), values...)
	if !containsString(result, value) {
		result = append(result, value)
	}
	return result
}

func allDigits(value string) bool {
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return false
		}
	}
	return true
}

func readLimitedBody(reader io.Reader, limit int64) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(reader, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("response body exceeds %d bytes", limit)
	}
	return body, nil
}

type limitedReadCloser struct {
	io.Reader
	io.Closer
}
