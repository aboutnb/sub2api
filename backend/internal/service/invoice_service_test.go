package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/enttest"
	"github.com/Wei-Shaw/sub2api/internal/config"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"

	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	_ "modernc.org/sqlite"
)

func TestInvoiceClientTokenFormCacheAndValidation(t *testing.T) {
	t.Parallel()

	var tokenCalls atomic.Int32
	var validationCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/oauth/token":
			tokenCalls.Add(1)
			if got := r.Header.Get("Content-Type"); got != "application/x-www-form-urlencoded" {
				t.Errorf("token content type = %q", got)
			}
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse token form: %v", err)
			}
			want := url.Values{
				"grant_type":    {"client_credentials"},
				"client_id":     {"client-id"},
				"client_secret": {"server-secret"},
				"scope":         {invoiceScope},
			}
			if r.Form.Encode() != want.Encode() {
				t.Errorf("token form = %q, want %q", r.Form.Encode(), want.Encode())
			}
			writeInvoiceTestJSON(t, w, http.StatusOK, map[string]any{
				"access_token": "cached-token", "token_type": "Bearer", "expires_in": 900, "scope": invoiceScope,
			})
		case "/api/v1/invoice-orders/validate":
			validationCalls.Add(1)
			if got := r.Header.Get("Authorization"); got != "Bearer cached-token" {
				t.Errorf("authorization = %q", got)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode validation request: %v", err)
			}
			if validationCalls.Load() == 1 {
				taxOrderNos, ok := payload["taxOrderNos"].([]any)
				if payload["needPayTax"] != true || !ok || len(taxOrderNos) != 1 {
					t.Errorf("tax validation payload = %#v", payload)
				}
				writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{
					"totalAmount": "102.00", "currency": "CNY", "taxAmount": "6.12", "taxPaidAmount": "6.12", "taxDueAmount": "0.00",
				})
				return
			}
			if _, found := payload["needPayTax"]; found {
				t.Errorf("no-tax validation unexpectedly sent needPayTax: %#v", payload)
			}
			if _, found := payload["taxOrderNos"]; found {
				t.Errorf("no-tax validation unexpectedly sent taxOrderNos: %#v", payload)
			}
			writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{
				// Merchant-side fee data must not become a user-paid invoice amount.
				"totalAmount": "102.00", "currency": "CNY", "taxAmount": "6.12", "invoiceAmount": "108.12",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := newInvoiceHTTPTestService(server)
	for _, testCase := range []struct {
		needPayTax  bool
		taxOrderNos []string
		wantAmount  string
	}{
		{needPayTax: true, taxOrderNos: []string{"INV-TAX-1"}, wantAmount: "108.12"},
		{needPayTax: false, wantAmount: "102.00"},
	} {
		result, err := service.validateRemoteOrders(context.Background(), []string{"ORDER-1"}, testCase.needPayTax, testCase.taxOrderNos)
		if err != nil {
			t.Fatalf("validate orders: %v", err)
		}
		if got := mapString(result, "totalAmount"); got != "102.00" {
			t.Fatalf("totalAmount = %q", got)
		}
		if got := mapString(result, "invoiceAmount"); got != testCase.wantAmount {
			t.Fatalf("invoiceAmount = %q, want %s", got, testCase.wantAmount)
		}
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Fatalf("token calls = %d, want 1", got)
	}
	if got := validationCalls.Load(); got != 2 {
		t.Fatalf("validation calls = %d, want 2", got)
	}
}

func TestSetInvoiceAmountUsesExactDecimalArithmetic(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		needPayTax bool
		validation map[string]any
		want       string
	}{
		{name: "whole amounts", needPayTax: true, validation: map[string]any{"totalAmount": "100.00", "taxAmount": "6.00", "taxPaidAmount": "0.00", "taxDueAmount": "6.00"}, want: "106.00"},
		{name: "cent precision", needPayTax: true, validation: map[string]any{"totalAmount": "102.00", "taxAmount": "6.12", "taxPaidAmount": "6.12", "taxDueAmount": "0.00"}, want: "108.12"},
		{name: "platform fee is excluded", needPayTax: false, validation: map[string]any{"totalAmount": "100.00", "taxAmount": "6.00", "invoiceAmount": "106.00"}, want: "100.00"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := setInvoiceAmount(tt.validation, tt.needPayTax); err != nil {
				t.Fatalf("set invoice amount: %v", err)
			}
			if got := mapString(tt.validation, "invoiceAmount"); got != tt.want {
				t.Fatalf("invoiceAmount = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestSetInvoiceAmountRejectsInconsistentTaxAmounts(t *testing.T) {
	t.Parallel()

	tests := []map[string]any{
		{"totalAmount": "100.00", "taxAmount": "6.00", "taxPaidAmount": "7.00", "taxDueAmount": "0.00"},
		{"totalAmount": "100.00", "taxAmount": "6.00", "taxPaidAmount": "2.00", "taxDueAmount": "5.00"},
		{"totalAmount": "100.00", "taxAmount": "6.00", "taxPaidAmount": "0.00", "taxDueAmount": "6.00", "invoiceAmount": "105.99"},
		{"totalAmount": "100.00", "taxAmount": "-6.00", "taxPaidAmount": "0.00", "taxDueAmount": "-6.00"},
	}
	for _, validation := range tests {
		if err := setInvoiceAmount(validation, true); err == nil {
			t.Fatalf("expected inconsistent tax validation to be rejected: %#v", validation)
		}
	}
}

func TestInvoiceClientRefreshesInvalidTokenOnce(t *testing.T) {
	t.Parallel()

	var tokenCalls atomic.Int32
	var apiCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/oauth/token" {
			call := tokenCalls.Add(1)
			writeInvoiceTestJSON(t, w, http.StatusOK, map[string]any{
				"access_token": "token-" + string(rune('0'+call)), "expires_in": 900,
			})
			return
		}
		call := apiCalls.Add(1)
		if call == 1 {
			writeInvoiceTestJSON(t, w, http.StatusUnauthorized, map[string]any{
				"code": "AUTH_INVALID_TOKEN", "message": "expired", "requestId": "req-auth",
			})
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer token-2" {
			t.Errorf("retry authorization = %q", got)
		}
		writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{"paid": true})
	}))
	defer server.Close()

	status, err := newInvoiceHTTPTestService(server).checkRemoteTaxStatus(context.Background(), "INV-TAX-1")
	if err != nil {
		t.Fatalf("check tax status: %v", err)
	}
	if !mapBool(status, "paid") {
		t.Fatal("expected paid status")
	}
	if tokenCalls.Load() != 2 || apiCalls.Load() != 2 {
		t.Fatalf("calls token=%d api=%d, want 2/2", tokenCalls.Load(), apiCalls.Load())
	}
}

func TestInvoiceTaxPaymentStatusPollsUntilPaid(t *testing.T) {
	var statusCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/oauth/token":
			writeInvoiceTestJSON(t, w, http.StatusOK, map[string]any{"access_token": "token", "expires_in": 900})
		case "/api/v1/invoice-tax-payments/status":
			call := statusCalls.Add(1)
			if call > 2 {
				t.Errorf("tax status was polled %d times after payment was confirmed", call)
			}
			writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{"paid": call == 2})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	status, paid, err := newInvoiceHTTPTestService(server).pollRemoteTaxPaymentStatus(context.Background(), "INV-TAX-POLL")
	if err != nil {
		t.Fatalf("poll tax status: %v", err)
	}
	if !paid || !mapBool(status, "paid") {
		t.Fatalf("poll result = %#v, paid=%v; want paid=true", status, paid)
	}
	if got := statusCalls.Load(); got != 2 {
		t.Fatalf("tax status calls = %d, want 2", got)
	}
}

func TestInvoiceTaxPaymentStatusDoesNotRetryUpstreamErrors(t *testing.T) {
	var statusCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/oauth/token":
			writeInvoiceTestJSON(t, w, http.StatusOK, map[string]any{"access_token": "token", "expires_in": 900})
		case "/api/v1/invoice-tax-payments/status":
			statusCalls.Add(1)
			writeInvoiceTestJSON(t, w, http.StatusTooManyRequests, map[string]any{
				"code": "RATE_LIMITED", "message": "try again later", "requestId": "req-tax-rate",
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	_, paid, err := newInvoiceHTTPTestService(server).pollRemoteTaxPaymentStatus(context.Background(), "INV-TAX-429")
	if err == nil {
		t.Fatal("expected upstream rate-limit error")
	}
	if paid {
		t.Fatal("rate-limited tax status must not be treated as paid")
	}
	if got := statusCalls.Load(); got != 1 {
		t.Fatalf("tax status calls = %d, want 1 after upstream error", got)
	}
}

func TestInvoiceClientCancelAndPDF(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/oauth/token":
			writeInvoiceTestJSON(t, w, http.StatusOK, map[string]any{"access_token": "token", "expires_in": 900})
		case r.Method == http.MethodPost && r.URL.EscapedPath() == "/api/v1/invoices/application-1/cancel":
			writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{"id": "application-1", "status": "canceled"})
		case r.Method == http.MethodGet && r.URL.EscapedPath() == "/api/v1/invoices/application-1/pdf":
			if got := r.Header.Get("Accept"); got != "application/pdf" {
				t.Errorf("PDF accept = %q", got)
			}
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Disposition", `attachment; filename="invoice.pdf"`)
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("%PDF-1.7 test"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := newInvoiceHTTPTestService(server)
	canceled, err := service.cancelRemoteApplication(context.Background(), "application-1")
	if err != nil {
		t.Fatalf("cancel invoice: %v", err)
	}
	if got := mapString(canceled, "status"); got != "canceled" {
		t.Fatalf("cancel status = %q", got)
	}

	pdf, err := service.downloadRemotePDF(context.Background(), "application-1")
	if err != nil {
		t.Fatalf("download PDF: %v", err)
	}
	defer func() { _ = pdf.Body.Close() }()
	body, err := io.ReadAll(pdf.Body)
	if err != nil {
		t.Fatalf("read PDF: %v", err)
	}
	if string(body) != "%PDF-1.7 test" || pdf.ContentType != "application/pdf" {
		t.Fatalf("unexpected PDF response: type=%q body=%q", pdf.ContentType, body)
	}
}

func TestInvoiceClientRedactsCredentialsFromTokenError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		writeInvoiceTestJSON(t, w, http.StatusUnauthorized, map[string]any{
			"code": "OAUTH_INVALID_CLIENT", "message": "client secret server-secret rejected", "requestId": "req-secret",
		})
	}))
	defer server.Close()

	_, err := newInvoiceHTTPTestService(server).getAccessToken(context.Background(), false)
	if err == nil {
		t.Fatal("expected token error")
	}
	if strings.Contains(err.Error(), "server-secret") {
		t.Fatalf("credential leaked in error: %v", err)
	}
}

func TestDecodeInvoiceEnvelopeRequiresExplicitZeroCode(t *testing.T) {
	t.Parallel()

	resp := &http.Response{
		StatusCode: http.StatusOK,
		Body:       io.NopCloser(strings.NewReader(`{"message":"ok","data":{"paid":true}}`)),
	}
	var result map[string]any
	upstream, err := decodeInvoiceEnvelope(resp, &result)
	if err == nil {
		t.Fatal("expected an envelope without code=0 to fail")
	}
	if upstream == nil || upstream.Code != "INVOICE_INVALID_RESPONSE" {
		t.Fatalf("upstream error = %#v", upstream)
	}
}

func TestInvoicePublicErrorRedactsMessageAndNormalizesSuccessStatus(t *testing.T) {
	t.Parallel()

	publicErr := invoicePublicError(&invoiceUpstreamError{
		StatusCode: http.StatusOK,
		Code:       "INVOICE_ORDER_NOT_ELIGIBLE",
		Message:    "provider database details must stay private",
		RequestID:  "req-public-error",
	})
	if got := infraerrors.Code(publicErr); got != http.StatusBadGateway {
		t.Fatalf("public status = %d, want %d", got, http.StatusBadGateway)
	}
	if got := infraerrors.Message(publicErr); got != "invoice service request failed" {
		t.Fatalf("public message = %q", got)
	}
	if strings.Contains(publicErr.Error(), "provider database details") {
		t.Fatalf("upstream message leaked in public error: %v", publicErr)
	}
	if got := infraerrors.FromError(publicErr).Metadata["request_id"]; got != "req-public-error" {
		t.Fatalf("request ID = %q", got)
	}
}

func TestInvoicePublicErrorPreservesTaxConflictCodes(t *testing.T) {
	t.Parallel()

	for _, code := range []string{
		"INVOICE_TAX_PAYMENT_EXCEEDS_REQUIRED",
		"INVOICE_TAX_ORDER_ALREADY_CLAIMED",
		"INVOICE_TAX_PAYMENT_NOT_ELIGIBLE",
	} {
		t.Run(code, func(t *testing.T) {
			err := invoicePublicError(&invoiceUpstreamError{
				StatusCode: http.StatusConflict,
				Code:       code,
				Message:    "provider conflict details",
				RequestID:  "req-tax-conflict",
			})
			if got := infraerrors.Code(err); got != http.StatusConflict {
				t.Fatalf("public status = %d, want %d", got, http.StatusConflict)
			}
			if got := infraerrors.Reason(err); got != code {
				t.Fatalf("public reason = %q, want %q", got, code)
			}
			if got := infraerrors.FromError(err).Metadata["request_id"]; got != "req-tax-conflict" {
				t.Fatalf("request ID = %q", got)
			}
			if strings.Contains(err.Error(), "provider conflict details") {
				t.Fatalf("upstream tax conflict leaked in public error: %v", err)
			}
		})
	}
}

func TestNormalizeInvoiceApplyRequest(t *testing.T) {
	t.Parallel()

	input := normalizeInvoiceApplyRequest(InvoiceApplyRequest{
		BuyerType:        " COMPANY ",
		Title:            " Example Ltd. ",
		TaxpayerID:       " TAX-1 ",
		BuyerBankAccount: "6222-0000 0000",
		RecipientEmail:   " invoice@example.test ",
	})
	if input.BuyerType != "company" || input.Title != "Example Ltd." || input.BuyerBankAccount != "622200000000" {
		t.Fatalf("normalized input = %#v", input)
	}
	if err := validateInvoiceApplyRequest(input); err != nil {
		t.Fatalf("normalized input should be valid: %v", err)
	}

	input.BuyerPhone = strings.Repeat("1", 65)
	if err := validateInvoiceApplyRequest(input); infraerrors.Reason(err) != "INVOICE_INVALID_FORM" {
		t.Fatalf("oversized phone reason = %q", infraerrors.Reason(err))
	}
}

func TestInvoiceApplicationOwnershipIsLocal(t *testing.T) {
	db, err := sql.Open("sqlite", "file:invoice-ownership?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	application, err := client.InvoiceApplication.Create().
		SetUserID(101).
		SetOrderIds([]int64{1}).
		SetOrderNos([]string{"ORDER-1"}).
		Save(context.Background())
	if err != nil {
		t.Fatalf("create application: %v", err)
	}

	service := &InvoiceService{entClient: client}
	if _, err := service.getOwnedApplication(context.Background(), 101, application.ID); err != nil {
		t.Fatalf("owner should access application: %v", err)
	}
	_, err = service.getOwnedApplication(context.Background(), 202, application.ID)
	if got := infraerrors.Reason(err); got != "NOT_FOUND" {
		t.Fatalf("cross-user error reason = %q, want NOT_FOUND", got)
	}
}

func TestInvoiceDraftRecoveryClaimsOrdersAndPreservesTaxDrafts(t *testing.T) {
	db, err := sql.Open("sqlite", "file:invoice-drafts?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	ctx := context.Background()

	noTaxDraft, err := client.InvoiceApplication.Create().
		SetUserID(101).
		SetOrderIds([]int64{7, 8}).
		SetOrderNos([]string{"ORDER-7", "ORDER-8"}).
		SetNeedPayTax(false).
		SetValidationSnapshot(map[string]any{"totalAmount": "100.00"}).
		Save(ctx)
	if err != nil {
		t.Fatalf("create no-tax draft: %v", err)
	}

	service := &InvoiceService{
		entClient: client,
		config: config.InvoiceIntegrationConfig{
			Enabled: true, BaseURL: "https://invoice.example.test", ClientID: "client", ClientSecret: "secret",
		},
	}
	recovered, err := service.CurrentDraft(ctx, 101)
	if err != nil {
		t.Fatalf("recover draft: %v", err)
	}
	if recovered == nil || recovered.DraftID != noTaxDraft.ID || !sameInvoiceOrderIDs(recovered.OrderIDs, []int64{7, 8}) {
		t.Fatalf("unexpected recovered draft: %#v", recovered)
	}
	if err := service.ensureOrdersNotClaimed(ctx, []int64{7}, 0); infraerrors.Reason(err) != "INVOICE_ORDER_ALREADY_CLAIMED" {
		t.Fatalf("active draft did not claim source order: %v", err)
	}
	if err := service.AbandonDraft(ctx, 101, noTaxDraft.ID); err != nil {
		t.Fatalf("abandon no-tax draft: %v", err)
	}
	if recovered, err := service.CurrentDraft(ctx, 101); err != nil || recovered != nil {
		t.Fatalf("abandoned draft should not be active: draft=%#v err=%v", recovered, err)
	}

	taxDraft, err := client.InvoiceApplication.Create().
		SetUserID(101).
		SetOrderIds([]int64{9}).
		SetOrderNos([]string{"ORDER-9"}).
		SetNeedPayTax(true).
		SetTaxOrderNos([]string{"INV-TAX-9"}).
		Save(ctx)
	if err != nil {
		t.Fatalf("create tax draft: %v", err)
	}
	if _, err := taxDraft.Update().SetStatus("failed").Save(ctx); err != nil {
		t.Fatalf("mark tax draft failed: %v", err)
	}
	if err := service.AbandonDraft(ctx, 101, taxDraft.ID); infraerrors.Reason(err) != "INVOICE_TAX_DRAFT_MUST_RESUME" {
		t.Fatalf("tax draft abandon error = %v", err)
	}
	if recovered, err := service.CurrentDraft(ctx, 101); err != nil || recovered == nil || recovered.DraftID != taxDraft.ID {
		t.Fatalf("tax draft was not preserved: draft=%#v err=%v", recovered, err)
	}
	if err := service.ensureOrdersNotClaimed(ctx, []int64{9}, 0); infraerrors.Reason(err) != "INVOICE_ORDER_ALREADY_CLAIMED" {
		t.Fatalf("failed tax draft did not retain source-order claim: %v", err)
	}
}

func TestInvoiceServiceEndToEndBothTaxModes(t *testing.T) {
	db, err := sql.Open("sqlite", "file:invoice-e2e?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(entsql.OpenDB(dialect.SQLite, db))))
	ctx := context.Background()
	user, err := client.User.Create().SetEmail("invoice-e2e@example.test").SetPasswordHash("test-hash").Save(ctx)
	if err != nil {
		t.Fatalf("create user: %v", err)
	}
	noTaxOrder := createInvoiceE2EOrder(t, ctx, client, user, "ORDER-NO-TAX")
	taxOrder := createInvoiceE2EOrder(t, ctx, client, user, "ORDER-TAX")

	var noTaxValidationCalls atomic.Int32
	var taxValidationCalls atomic.Int32
	var tokenCalls atomic.Int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/oauth/token":
			tokenCalls.Add(1)
			if err := r.ParseForm(); err != nil {
				t.Fatalf("parse token request: %v", err)
			}
			if r.Form.Get("scope") != invoiceScope || r.Form.Get("grant_type") != "client_credentials" {
				t.Errorf("unexpected token request: %q", r.Form.Encode())
			}
			writeInvoiceTestJSON(t, w, http.StatusOK, map[string]any{"access_token": "e2e-token", "expires_in": 900, "scope": invoiceScope})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/invoice-orders/validate":
			assertInvoiceE2EAuthorization(t, r)
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode validation request: %v", err)
			}
			orderNos := mapStringSlice(payload, "orderNos")
			if len(orderNos) != 1 {
				t.Fatalf("validation orderNos = %#v", orderNos)
			}
			switch orderNos[0] {
			case "ORDER-NO-TAX":
				noTaxValidationCalls.Add(1)
				if _, found := payload["needPayTax"]; found {
					t.Errorf("no-tax validation unexpectedly included needPayTax: %#v", payload)
				}
				writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{
					"orders": []map[string]any{{"orderNo": "ORDER-NO-TAX", "amount": "100.00", "currency": "CNY"}},
					// Provider bookkeeping may include the merchant-side fee. The
					// customer-facing invoice amount must remain the order amount.
					"totalAmount": "100.00", "currency": "CNY", "taxAmount": "6.00", "invoiceAmount": "106.00",
				})
			case "ORDER-TAX":
				taxValidationCalls.Add(1)
				if payload["needPayTax"] != true {
					t.Errorf("tax validation missing needPayTax: %#v", payload)
				}
				paidTaxOrders := mapStringSlice(payload, "taxOrderNos")
				data := map[string]any{
					"orders":      []map[string]any{{"orderNo": "ORDER-TAX", "amount": "100.00", "currency": "CNY"}},
					"totalAmount": "100.00", "currency": "CNY", "taxAmount": "6.00",
				}
				if len(paidTaxOrders) == 0 {
					data["taxPaidAmount"] = "0.00"
					data["taxDueAmount"] = "6.00"
					data["taxPayments"] = map[string]any{
						"alipay": map[string]any{"taxOrderNo": "INV-TAX-ALIPAY", "payUrl": "https://cashier.example.test/alipay"},
						"wxpay":  map[string]any{"taxOrderNo": "INV-TAX-WXPAY", "payUrl": "https://cashier.example.test/wxpay"},
					}
				} else if len(paidTaxOrders) == 1 && paidTaxOrders[0] == "INV-TAX-ALIPAY" {
					data["taxPaidAmount"] = "6.00"
					data["taxDueAmount"] = "0.00"
				} else {
					t.Fatalf("unexpected paid tax orders: %#v", paidTaxOrders)
				}
				writeInvoiceTestEnvelope(t, w, http.StatusOK, data)
			default:
				t.Fatalf("unexpected invoice order: %q", orderNos[0])
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/invoice-tax-payments/status":
			assertInvoiceE2EAuthorization(t, r)
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode tax status request: %v", err)
			}
			if payload["taxOrderNo"] != "INV-TAX-ALIPAY" {
				t.Errorf("tax order = %q", payload["taxOrderNo"])
			}
			writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{"paid": true})
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/invoices":
			assertInvoiceE2EAuthorization(t, r)
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode application request: %v", err)
			}
			orderNos := mapStringSlice(payload, "orderNos")
			if len(orderNos) != 1 {
				t.Fatalf("application orderNos = %#v", orderNos)
			}
			switch orderNos[0] {
			case "ORDER-NO-TAX":
				if _, found := payload["needPayTax"]; found {
					t.Errorf("no-tax application unexpectedly included needPayTax: %#v", payload)
				}
				writeInvoiceTestEnvelope(t, w, http.StatusCreated, map[string]any{"id": "APP-NO-TAX", "status": "pending"})
			case "ORDER-TAX":
				if payload["needPayTax"] != true || len(mapStringSlice(payload, "taxOrderNos")) != 1 {
					t.Errorf("tax application payload = %#v", payload)
				}
				writeInvoiceTestEnvelope(t, w, http.StatusCreated, map[string]any{"id": "APP-TAX", "status": "pending"})
			default:
				t.Fatalf("unexpected application order: %q", orderNos[0])
			}
		case r.Method == http.MethodPost && r.URL.Path == "/api/v1/invoices/APP-NO-TAX/cancel":
			assertInvoiceE2EAuthorization(t, r)
			writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{"id": "APP-NO-TAX", "status": "canceled"})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/invoices":
			assertInvoiceE2EAuthorization(t, r)
			writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{
				"items": []map[string]any{
					{"id": "APP-NO-TAX", "status": "canceled", "orderNos": []string{"ORDER-NO-TAX"}},
					{"id": "APP-TAX", "status": "completed", "orderNos": []string{"ORDER-TAX"}},
				},
				"total": 2, "page": 1, "pageSize": 100,
			})
		case r.Method == http.MethodGet && r.URL.Path == "/api/v1/invoices/APP-TAX/pdf":
			assertInvoiceE2EAuthorization(t, r)
			w.Header().Set("Content-Type", "application/pdf")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("%PDF-1.7 invoice e2e"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := newInvoiceHTTPTestService(server)
	service.entClient = client

	noTaxDraft, err := service.ValidateOrders(ctx, user.ID, []int64{noTaxOrder.ID}, false)
	if err != nil {
		t.Fatalf("validate platform-paid invoice: %v", err)
	}
	if noTaxDraft.NeedPayTax || mapString(noTaxDraft.Validation, "invoiceAmount") != "100.00" {
		t.Fatalf("unexpected platform-paid draft: %#v", noTaxDraft)
	}
	recovered, err := service.CurrentDraft(ctx, user.ID)
	if err != nil || recovered == nil || recovered.DraftID != noTaxDraft.DraftID {
		t.Fatalf("platform-paid draft did not recover: draft=%#v err=%v", recovered, err)
	}
	reused, err := service.ValidateOrders(ctx, user.ID, []int64{noTaxOrder.ID}, false)
	if err != nil || reused.DraftID != noTaxDraft.DraftID || noTaxValidationCalls.Load() != 1 {
		t.Fatalf("same selection did not reuse draft: draft=%#v err=%v calls=%d", reused, err, noTaxValidationCalls.Load())
	}
	if _, err := service.ValidateOrders(ctx, user.ID, []int64{taxOrder.ID}, true); infraerrors.Reason(err) != "INVOICE_DRAFT_ACTIVE" {
		t.Fatalf("different selection with an active draft = %v", err)
	}

	noTaxApplication, err := service.Apply(ctx, user.ID, noTaxDraft.DraftID, invoiceE2EBuyer())
	if err != nil || noTaxApplication.ExternalID != "APP-NO-TAX" || noTaxApplication.TotalAmount != "100.00" {
		t.Fatalf("submit platform-paid invoice: application=%#v err=%v", noTaxApplication, err)
	}
	if _, err := service.Cancel(ctx, user.ID, noTaxApplication.ID); err != nil {
		t.Fatalf("cancel platform-paid invoice: %v", err)
	}

	taxDraft, err := service.ValidateOrders(ctx, user.ID, []int64{taxOrder.ID}, true)
	if err != nil {
		t.Fatalf("validate customer-paid invoice: %v", err)
	}
	if !taxDraft.NeedPayTax || mapString(taxDraft.Validation, "invoiceAmount") != "106.00" || mapString(taxDraft.Validation, "taxDueAmount") != "6.00" {
		t.Fatalf("unexpected customer-paid draft: %#v", taxDraft)
	}
	if !taxOrderBelongsToValidation(taxDraft.Validation, "INV-TAX-ALIPAY") || !taxOrderBelongsToValidation(taxDraft.Validation, "INV-TAX-WXPAY") {
		t.Fatalf("customer-paid draft did not retain both checkout channels: %#v", taxDraft.Validation)
	}
	resumedTaxDraft, err := service.ValidateOrders(ctx, user.ID, []int64{taxOrder.ID}, true)
	if err != nil || resumedTaxDraft.DraftID != taxDraft.DraftID || taxValidationCalls.Load() != 1 {
		t.Fatalf("tax draft did not resume safely: draft=%#v err=%v calls=%d", resumedTaxDraft, err, taxValidationCalls.Load())
	}
	taxStatus, err := service.CheckTaxPayment(ctx, user.ID, taxDraft.DraftID, "INV-TAX-ALIPAY")
	if err != nil || !taxStatus.Paid || !taxStatus.Ready || len(taxStatus.TaxOrderNos) != 1 {
		t.Fatalf("reconcile tax payment: status=%#v err=%v", taxStatus, err)
	}
	taxApplication, err := service.Apply(ctx, user.ID, taxDraft.DraftID, invoiceE2EBuyer())
	if err != nil || taxApplication.ExternalID != "APP-TAX" || taxApplication.TotalAmount != "106.00" {
		t.Fatalf("submit customer-paid invoice: application=%#v err=%v", taxApplication, err)
	}
	if _, err := service.ValidateOrders(ctx, user.ID, []int64{taxOrder.ID}, true); infraerrors.Reason(err) != "INVOICE_ORDER_ALREADY_CLAIMED" {
		t.Fatalf("submitted invoice order was not claimed: %v", err)
	}

	applications, total, err := service.ListApplications(ctx, user.ID, 1, 20)
	if err != nil || total != 2 || len(applications) != 2 {
		t.Fatalf("list applications: items=%#v total=%d err=%v", applications, total, err)
	}
	pdf, err := service.DownloadPDF(ctx, user.ID, taxApplication.ID)
	if err != nil {
		t.Fatalf("download completed PDF: %v", err)
	}
	defer func() { _ = pdf.Body.Close() }()
	pdfBody, err := io.ReadAll(pdf.Body)
	if err != nil || string(pdfBody) != "%PDF-1.7 invoice e2e" {
		t.Fatalf("downloaded PDF = %q, err=%v", pdfBody, err)
	}
	if tokenCalls.Load() != 1 {
		t.Fatalf("token calls = %d, want cached token", tokenCalls.Load())
	}
}

func createInvoiceE2EOrder(t *testing.T, ctx context.Context, client *dbent.Client, user *dbent.User, orderNo string) *dbent.PaymentOrder {
	t.Helper()
	order, err := client.PaymentOrder.Create().
		SetUserID(user.ID).
		SetUserEmail(user.Email).
		SetUserName("Invoice E2E").
		SetAmount(100).
		SetPayAmount(100).
		SetRechargeCode("invoice-e2e").
		SetOutTradeNo(orderNo).
		SetPaymentType(payment.TypeAlipay).
		SetPaymentTradeNo("TX-" + orderNo).
		SetStatus(payment.OrderStatusCompleted).
		SetExpiresAt(time.Now().Add(time.Hour)).
		SetClientIP("127.0.0.1").
		SetSrcHost("invoice-e2e.test").
		Save(ctx)
	if err != nil {
		t.Fatalf("create payment order %s: %v", orderNo, err)
	}
	return order
}

func invoiceE2EBuyer() InvoiceApplyRequest {
	return InvoiceApplyRequest{
		BuyerType: "company", Title: "Invoice E2E Ltd.", TaxpayerID: "91310000677833266F", RecipientEmail: "invoice-e2e@example.test",
	}
}

func assertInvoiceE2EAuthorization(t *testing.T, r *http.Request) {
	t.Helper()
	if got := r.Header.Get("Authorization"); got != "Bearer e2e-token" {
		t.Errorf("authorization = %q", got)
	}
}

func TestInvoiceListRecoversUnknownSubmissionByOrderSet(t *testing.T) {
	db, err := sql.Open("sqlite", "file:invoice-recovery?mode=memory&cache=shared&_fk=1")
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	defer func() { _ = db.Close() }()
	if _, err := db.Exec("PRAGMA foreign_keys = ON"); err != nil {
		t.Fatalf("enable sqlite foreign keys: %v", err)
	}
	drv := entsql.OpenDB(dialect.SQLite, db)
	client := enttest.NewClient(t, enttest.WithOptions(dbent.Driver(drv)))
	_, err = client.InvoiceApplication.Create().
		SetUserID(101).
		SetOrderIds([]int64{2, 3}).
		SetOrderNos([]string{"ORDER-3", "ORDER-2"}).
		SetStatus(invoiceStatusUnknown).
		Save(context.Background())
	if err != nil {
		t.Fatalf("create unknown application: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/oauth/token":
			writeInvoiceTestJSON(t, w, http.StatusOK, map[string]any{"access_token": "token", "expires_in": 900})
		case "/api/v1/invoices":
			writeInvoiceTestEnvelope(t, w, http.StatusOK, map[string]any{
				"items": []map[string]any{{
					"id": 901, "status": "completed", "orderNos": []string{"ORDER-2", "ORDER-3"},
				}},
				"total": 1, "page": 1, "pageSize": 100,
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	service := newInvoiceHTTPTestService(server)
	service.entClient = client
	items, total, err := service.ListApplications(context.Background(), 101, 1, 20)
	if err != nil {
		t.Fatalf("list applications: %v", err)
	}
	if total != 1 || len(items) != 1 {
		t.Fatalf("list result total=%d items=%d", total, len(items))
	}
	if items[0].ExternalID != "901" || items[0].Status != "completed" {
		t.Fatalf("submission not recovered: external_id=%q status=%q", items[0].ExternalID, items[0].Status)
	}
}

func TestNormalizeInvoiceOrderIDs(t *testing.T) {
	t.Parallel()

	ids, err := normalizeInvoiceOrderIDs([]int64{9, 2, 5})
	if err != nil {
		t.Fatalf("normalize order IDs: %v", err)
	}
	if got := []int64{2, 5, 9}; len(ids) != len(got) || ids[0] != got[0] || ids[1] != got[1] || ids[2] != got[2] {
		t.Fatalf("normalized IDs = %#v", ids)
	}
	if _, err := normalizeInvoiceOrderIDs([]int64{1, 1}); infraerrors.Reason(err) != "INVOICE_DUPLICATE_ORDER" {
		t.Fatalf("duplicate reason = %q", infraerrors.Reason(err))
	}
}

func newInvoiceHTTPTestService(server *httptest.Server) *InvoiceService {
	client := server.Client()
	client.Timeout = 2 * time.Second
	return &InvoiceService{
		config: config.InvoiceIntegrationConfig{
			Enabled: true, BaseURL: server.URL, ClientID: "client-id", ClientSecret: "server-secret", TimeoutSeconds: 2,
		},
		httpClient: client,
	}
}

func writeInvoiceTestEnvelope(t *testing.T, w http.ResponseWriter, status int, data any) {
	t.Helper()
	writeInvoiceTestJSON(t, w, status, map[string]any{"code": 0, "message": "ok", "data": data})
}

func writeInvoiceTestJSON(t *testing.T, w http.ResponseWriter, status int, payload any) {
	t.Helper()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(payload); err != nil {
		t.Errorf("encode test response: %v", err)
	}
}
