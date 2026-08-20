package usdtpayment

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func TestHMACV2AndPayloadSignature(t *testing.T) {
	body := []byte(`{"amount":"102.00"}`)
	digest := sha256Hex(body)
	signature := hmacV2Sign("secret", http.MethodPost, "/api/v1/merchant/order/create", "1700000000", "nonce-1234567890", digest)
	if !equalHex(signature, hmacV2Sign("secret", "post", "/api/v1/merchant/order/create", "1700000000", "nonce-1234567890", digest)) {
		t.Fatal("method case changed HMAC signature")
	}
	if equalHex(signature, hmacV2Sign("secret", http.MethodPost, "/api/v1/merchant/order/query", "1700000000", "nonce-1234567890", digest)) {
		t.Fatal("path change was accepted by HMAC")
	}
	payloadSignature := hmacHex("secret", body)
	if !equalHex(payloadSignature, hmacHex("secret", body)) || equalHex(payloadSignature, hmacHex("secret", []byte(`{"amount":"101.00"}`))) {
		t.Fatal("payload HMAC did not bind to the body")
	}
}

func TestVerifyWebhookRejectsTamperedBody(t *testing.T) {
	service := &Service{config: testUSDTConfig()}
	payload := WebhookPayload{
		EventID: "evt_1", EventType: "payment.succeeded", OccurredAt: time.Now().Unix(),
		OrderID: "sub2_order", TradeID: "trade_1", Fiat: "CNY", Amount: "102.00", Crypto: "USDT",
		ActualAmount: "14.5714", TradeType: "usdt.trc20", Network: "tron", Token: "TAddress",
		BlockTransactionID: "0xhash", TransferAt: time.Now().Unix(), BlockNumber: 123,
		SignatureVersion: "v2", KeyID: "key-1",
	}
	unsigned, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	payload.Signature = hmacHex(service.config.APISecret, unsigned)
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	digest := sha256Hex(raw)
	headers := map[string]string{
		strings.ToLower(headerKeyID):     service.config.KeyID,
		strings.ToLower(headerTimestamp): timestamp,
		strings.ToLower(headerNonce):     "nonce-1234567890123456",
		strings.ToLower(headerDigest):    digest,
		strings.ToLower(headerSignature): hmacV2Sign(service.config.APISecret, http.MethodPost, "/api/v1/usdt/webhook/bepusdt", timestamp, "nonce-1234567890123456", digest),
	}
	if _, err := service.VerifyWebhook(raw, headers, "/api/v1/usdt/webhook/bepusdt", time.Now()); err != nil {
		t.Fatalf("valid webhook rejected: %v", err)
	}
	tampered := bytes.Replace(raw, []byte(`"amount":"102.00"`), []byte(`"amount":"103.00"`), 1)
	if _, err := service.VerifyWebhook(tampered, headers, "/api/v1/usdt/webhook/bepusdt", time.Now()); err == nil {
		t.Fatal("tampered webhook was accepted")
	}
}

func TestClientSignsServerRequest(t *testing.T) {
	var gotHeaders http.Header
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotHeaders = r.Header.Clone()
		_, _ = w.Write([]byte(`{"code":"ok","data":{"networks":[]}}`))
	}))
	defer server.Close()
	client := &Client{apiBase: server.URL, keyID: "key-1", secret: "secret", httpClient: server.Client()}
	if _, err := client.Capabilities(t.Context(), []string{"tron"}); err != nil {
		t.Fatalf("capabilities request failed: %v", err)
	}
	bodyDigest := gotHeaders.Get(headerDigest)
	if bodyDigest == "" || gotHeaders.Get(headerKeyID) != "key-1" || gotHeaders.Get(headerNonce) == "" {
		t.Fatalf("missing signed headers: %v", gotHeaders)
	}
	if len(bodyDigest) != sha256.Size*2 || !equalHex(gotHeaders.Get(headerSignature), hmacV2Sign("secret", http.MethodPost, "/api/v1/merchant/capabilities", gotHeaders.Get(headerTimestamp), gotHeaders.Get(headerNonce), bodyDigest)) {
		t.Fatal("request signature does not match headers")
	}
}

func testUSDTConfig() config.USDTPaymentConfig {
	return config.USDTPaymentConfig{
		Enabled: true, KeyID: "key-1", APISecret: "secret", Fiat: "CNY", WebhookClockSkewSeconds: 300,
	}
}
