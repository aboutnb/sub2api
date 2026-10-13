package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/payment"
)

func TestEasyPayPopupBuildsGMUsdtSelector(t *testing.T) {
	provider, err := NewEasyPay("gm", map[string]string{
		"pid":           "1000",
		"pkey":          "gm-secret",
		"apiBase":       "https://gm.example.com/payments/epay/v1/order/create-transaction",
		"notifyUrl":     "https://flowai.example.com/api/v1/payment/webhook/easypay",
		"returnUrl":     "https://flowai.example.com/payment/result",
		"paymentMode":   paymentModePopup,
		"customMethods": `[{"type":"usdt_trc20","upstreamType":"usdt.tron","displayName":"USDT-TRC20"}]`,
	})
	if err != nil {
		t.Fatalf("NewEasyPay returned error: %v", err)
	}

	result, err := provider.CreatePayment(context.Background(), payment.CreatePaymentRequest{
		OrderID:     "order-gm-usdt-1",
		Amount:      "10.00",
		PaymentType: "usdt_trc20",
		Subject:     "USDT top up",
		NotifyURL:   "https://flowai.example.com/api/v1/payment/webhook/easypay",
		ReturnURL:   "https://flowai.example.com/payment/result",
	})
	if err != nil {
		t.Fatalf("CreatePayment returned error: %v", err)
	}

	parsed, err := url.Parse(result.PayURL)
	if err != nil {
		t.Fatalf("parse PayURL: %v", err)
	}
	if parsed.Path != "/payments/epay/v1/order/create-transaction/submit.php" {
		t.Fatalf("PayURL path = %q, want GM EPay submit.php path", parsed.Path)
	}

	query := parsed.Query()
	if query.Get("type") != "usdt.tron" {
		t.Fatalf("type = %q, want usdt.tron", query.Get("type"))
	}
	if query.Get("pid") != "1000" {
		t.Fatalf("pid = %q, want 1000", query.Get("pid"))
	}

	signed := make(map[string]string, len(query))
	for key, values := range query {
		if len(values) > 0 {
			signed[key] = values[0]
		}
	}
	if got, want := query.Get("sign"), easyPaySign(signed, "gm-secret"); got != want {
		t.Fatalf("sign = %q, want %q", got, want)
	}
}

func TestEasyPayGMVerifiesEpayGetCallback(t *testing.T) {
	t.Parallel()

	provider, err := NewEasyPay("gm", map[string]string{
		"pid":           "1000",
		"pkey":          "gm-secret",
		"apiBase":       "https://gm.example.com/payments/epay/v1/order/create-transaction",
		"notifyUrl":     "https://flowai.example.com/api/v1/payment/webhook/easypay",
		"returnUrl":     "https://flowai.example.com/payment/result",
		"customMethods": `[{"type":"usdt_trc20","upstreamType":"usdt.tron","displayName":"USDT-TRC20"}]`,
	})
	if err != nil {
		t.Fatalf("NewEasyPay returned error: %v", err)
	}

	values := url.Values{
		"pid":          {"1000"},
		"trade_no":     {"gm-trade-1"},
		"out_trade_no": {"sub2_gm_order_1"},
		"type":         {"usdt.tron"},
		"name":         {"USDT top up"},
		"money":        {"10.0000"},
		"trade_status": {tradeStatusSuccess},
		"sign_type":    {signTypeMD5},
	}
	signed := make(map[string]string, len(values))
	for key, items := range values {
		if len(items) > 0 {
			signed[key] = items[0]
		}
	}
	values.Set("sign", easyPaySign(signed, "gm-secret"))

	notification, err := provider.VerifyNotification(context.Background(), values.Encode(), nil)
	if err != nil {
		t.Fatalf("VerifyNotification returned error: %v", err)
	}
	if notification == nil {
		t.Fatal("VerifyNotification returned nil notification")
	}
	if notification.TradeNo != "gm-trade-1" {
		t.Fatalf("trade number = %q, want gm-trade-1", notification.TradeNo)
	}
	if notification.OrderID != "sub2_gm_order_1" {
		t.Fatalf("order ID = %q, want sub2_gm_order_1", notification.OrderID)
	}
	if notification.Amount != 10 {
		t.Fatalf("amount = %v, want 10", notification.Amount)
	}
	if notification.Status != payment.ProviderStatusSuccess {
		t.Fatalf("status = %q, want %q", notification.Status, payment.ProviderStatusSuccess)
	}
	if notification.Metadata["pid"] != "1000" {
		t.Fatalf("metadata pid = %q, want 1000", notification.Metadata["pid"])
	}
	if notification.Metadata["payment_type"] != "usdt_trc20" {
		t.Fatalf("metadata payment type = %q, want usdt_trc20", notification.Metadata["payment_type"])
	}
}

func TestEasyPayRejectsCrossOrderSignatureReplay(t *testing.T) {
	t.Parallel()

	provider, err := NewEasyPay("gm", map[string]string{
		"pid":           "1001",
		"pkey":          "gm-secret",
		"apiBase":       "https://gm.example.com/payments/epay/v1/order/create-transaction",
		"notifyUrl":     "https://flowai.example.com/api/v1/payment/webhook/easypay",
		"returnUrl":     "https://flowai.example.com/payment/result",
		"customMethods": `[{"type":"usdt_bep20","upstreamType":"usdt.binance","displayName":"USDT-BEP20"}]`,
	})
	if err != nil {
		t.Fatalf("NewEasyPay returned error: %v", err)
	}

	creation := map[string]string{
		"money":        "101.50",
		"name":         "100.00 FA",
		"notify_url":   "https://flowai.example.com/api/v1/payment/webhook/easypay",
		"out_trade_no": "sub2_second",
		"pid":          "1001",
		"return_url":   "https://&out_trade_no=sub2_target&trade_no=forged-trade&trade_status=TRADE_SUCCESS&type=@flowai.example.com/payment/result?order_id=293&out_trade_no=sub2_second&status=success",
		"type":         "usdt.binance",
	}
	creationSign := easyPaySign(creation, "gm-secret")
	forged := url.Values{
		"money":        {creation["money"]},
		"name":         {"100.00 FA&notify_url=https://flowai.example.com/api/v1/payment/webhook/easypay&out_trade_no=sub2_second&pid=1001&return_url=https://"},
		"out_trade_no": {"sub2_target"},
		"trade_no":     {"forged-trade"},
		"trade_status": {tradeStatusSuccess},
		"type":         {"@flowai.example.com/payment/result?order_id=293&out_trade_no=sub2_second&status=success&type=usdt.binance"},
		"sign":         {creationSign},
		"sign_type":    {signTypeMD5},
	}
	if got := easyPaySign(mapFromURLValues(forged), "gm-secret"); got != creationSign {
		t.Fatalf("test fixture does not reproduce signature collision: got %q, want %q", got, creationSign)
	}

	if _, err := provider.VerifyNotification(context.Background(), forged.Encode(), nil); err == nil {
		t.Fatal("VerifyNotification accepted a cross-order order-creation signature replay")
	}
}

func TestEasyPayRejectsDuplicateCallbackParameter(t *testing.T) {
	t.Parallel()

	provider, err := NewEasyPay("gm", map[string]string{
		"pid":       "1001",
		"pkey":      "gm-secret",
		"apiBase":   "https://gm.example.com",
		"notifyUrl": "https://flowai.example.com/api/v1/payment/webhook/easypay",
		"returnUrl": "https://flowai.example.com/payment/result",
	})
	if err != nil {
		t.Fatalf("NewEasyPay returned error: %v", err)
	}

	raw := "pid=1001&pid=1002&trade_no=t1&out_trade_no=o1&type=alipay&money=1.00&trade_status=TRADE_SUCCESS&sign_type=MD5&sign=invalid"
	if _, err := provider.VerifyNotification(context.Background(), raw, nil); err == nil {
		t.Fatal("VerifyNotification accepted a duplicate callback parameter")
	}
}

func mapFromURLValues(values url.Values) map[string]string {
	params := make(map[string]string, len(values))
	for key, items := range values {
		if len(items) == 1 {
			params[key] = items[0]
		}
	}
	return params
}

func TestEasyPayConfirmsUSDTNotificationWithSignedOrderQuery(t *testing.T) {
	t.Parallel()

	queryFields := map[string]string{
		"code":                 "1",
		"msg":                  "success",
		"pid":                  "1000",
		"out_trade_no":         "sub2_gm_order_1",
		"trade_no":             "gm-trade-1",
		"type":                 "usdt.tron",
		"money":                "10.0000",
		"trade_status":         tradeStatusSuccess,
		"status":               "1",
		"block_transaction_id": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}
	server := newEasyPayGMQueryServer(t, queryFields, false)
	provider := newEasyPayGMConfirmationProvider(t, server.URL)

	confirmed, err := provider.ConfirmNotification(context.Background(), &payment.PaymentNotification{
		TradeNo: "gm-trade-1",
		OrderID: "sub2_gm_order_1",
		Amount:  10,
		Status:  payment.ProviderStatusSuccess,
		Metadata: map[string]string{
			"pid":          "1000",
			"payment_type": "usdt_trc20",
		},
	})
	if err != nil {
		t.Fatalf("ConfirmNotification returned error: %v", err)
	}
	if confirmed == nil || confirmed.TradeNo != "gm-trade-1" || confirmed.OrderID != "sub2_gm_order_1" {
		t.Fatalf("confirmed notification = %+v", confirmed)
	}
	if confirmed.Metadata["query_signature_verified"] != "true" {
		t.Fatal("signed query response was not marked verified")
	}
}

func TestEasyPayRejectsUSDTNotificationWhenOrderQueryDoesNotMatch(t *testing.T) {
	t.Parallel()

	base := map[string]string{
		"code":                 "1",
		"msg":                  "success",
		"pid":                  "1000",
		"out_trade_no":         "sub2_gm_order_1",
		"trade_no":             "gm-trade-1",
		"type":                 "usdt.tron",
		"money":                "10.0000",
		"trade_status":         tradeStatusSuccess,
		"status":               "1",
		"block_transaction_id": "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
	}
	tests := []struct {
		name          string
		mutate        func(map[string]string)
		corruptSign   bool
		callbackMoney float64
	}{
		{name: "pending", mutate: func(fields map[string]string) { fields["trade_status"] = "WAITING"; fields["status"] = "0" }},
		{name: "wrong order", mutate: func(fields map[string]string) { fields["out_trade_no"] = "sub2_other" }},
		{name: "wrong trade", mutate: func(fields map[string]string) { fields["trade_no"] = "gm-other" }},
		{name: "wrong type", mutate: func(fields map[string]string) { fields["type"] = "usdt.binance" }},
		{name: "missing chain hash", mutate: func(fields map[string]string) { fields["block_transaction_id"] = "" }},
		{name: "invalid chain hash", mutate: func(fields map[string]string) { fields["block_transaction_id"] = "not-a-chain-hash" }},
		{name: "amount mismatch", callbackMoney: 9},
		{name: "invalid response signature", corruptSign: true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			fields := make(map[string]string, len(base))
			for key, value := range base {
				fields[key] = value
			}
			if tc.mutate != nil {
				tc.mutate(fields)
			}
			server := newEasyPayGMQueryServer(t, fields, tc.corruptSign)
			provider := newEasyPayGMConfirmationProvider(t, server.URL)
			callbackMoney := tc.callbackMoney
			if callbackMoney == 0 {
				callbackMoney = 10
			}
			_, err := provider.ConfirmNotification(context.Background(), &payment.PaymentNotification{
				TradeNo: "gm-trade-1",
				OrderID: "sub2_gm_order_1",
				Amount:  callbackMoney,
				Status:  payment.ProviderStatusSuccess,
				Metadata: map[string]string{
					"pid":          "1000",
					"payment_type": "usdt_trc20",
				},
			})
			if err == nil {
				t.Fatal("ConfirmNotification accepted mismatched upstream evidence")
			}
		})
	}
}

func TestEasyPayDoesNotQueryForRMBNotification(t *testing.T) {
	t.Parallel()
	provider, err := NewEasyPay("rmb", map[string]string{
		"pid":       "1000",
		"pkey":      "gm-secret",
		"apiBase":   "https://unreachable.invalid",
		"notifyUrl": "https://flowai.example.com/api/v1/payment/webhook/easypay",
		"returnUrl": "https://flowai.example.com/payment/result",
	})
	if err != nil {
		t.Fatalf("NewEasyPay returned error: %v", err)
	}
	notification := &payment.PaymentNotification{
		TradeNo: "rmb-trade-1", OrderID: "sub2_rmb_1", Amount: 100,
		Status:   payment.ProviderStatusSuccess,
		Metadata: map[string]string{"pid": "1000", "payment_type": payment.TypeAlipay},
	}
	confirmed, err := provider.ConfirmNotification(context.Background(), notification)
	if err != nil {
		t.Fatalf("ConfirmNotification returned error for RMB callback: %v", err)
	}
	if confirmed != notification {
		t.Fatal("RMB callback should pass through without an upstream USDT query")
	}
}

func newEasyPayGMConfirmationProvider(t *testing.T, origin string) *EasyPay {
	t.Helper()
	provider, err := NewEasyPay("gm", map[string]string{
		"pid":           "1000",
		"pkey":          "gm-secret",
		"apiBase":       origin + "/payments/epay/v1/order/create-transaction",
		"notifyUrl":     "https://flowai.example.com/api/v1/payment/webhook/easypay",
		"returnUrl":     "https://flowai.example.com/payment/result",
		"customMethods": `[{"type":"usdt_trc20","upstreamType":"usdt.tron","displayName":"USDT-TRC20"},{"type":"usdt_bep20","upstreamType":"usdt.binance","displayName":"USDT-BEP20"}]`,
	})
	if err != nil {
		t.Fatalf("NewEasyPay returned error: %v", err)
	}
	return provider
}

func newEasyPayGMQueryServer(t *testing.T, fields map[string]string, corruptSignature bool) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/payments/epay/v1/order/create-transaction/api.php" {
			t.Errorf("query path = %q", r.URL.Path)
		}
		if err := r.ParseForm(); err != nil {
			t.Errorf("ParseForm: %v", err)
		}
		if r.PostForm.Get("act") != "order" || r.PostForm.Get("pid") != "1000" || r.PostForm.Get("key") != "gm-secret" || r.PostForm.Get("out_trade_no") != "sub2_gm_order_1" {
			t.Errorf("unexpected query form: %v", r.PostForm)
		}

		response := make(map[string]interface{}, len(fields)+2)
		for key, value := range fields {
			switch key {
			case "code", "status":
				parsed, _ := strconv.Atoi(value)
				response[key] = parsed
			default:
				response[key] = value
			}
		}
		signature := easyPaySign(fields, "gm-secret")
		if corruptSignature {
			signature = "invalid"
		}
		response["sign"] = signature
		response["sign_type"] = signTypeMD5
		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(response); err != nil {
			t.Errorf("encode response: %v", err)
		}
	}))
	t.Cleanup(server.Close)
	return server
}
