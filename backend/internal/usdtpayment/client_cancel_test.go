package usdtpayment

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func TestClientCancelOrder(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		message    string
		wantErr    bool
	}{
		{name: "success", statusCode: http.StatusOK},
		{name: "already missing is idempotent", statusCode: 400, message: "订单不存在"},
		{name: "upstream failure", statusCode: 400, message: "订单已支付", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var gotPath string
			var gotRequest map[string]any
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotPath = r.URL.Path
				if err := json.NewDecoder(r.Body).Decode(&gotRequest); err != nil {
					t.Fatalf("decode request: %v", err)
				}
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"status_code":` + strconv.Itoa(tt.statusCode) + `,"message":"` + tt.message + `"}`))
			}))
			defer server.Close()

			cfg := config.USDTPaymentConfig{APIBase: server.URL, LegacyToken: "legacy-token"}
			client := &Client{config: cfg, httpClient: server.Client()}
			err := client.CancelOrder(t.Context(), "trade-123")
			if (err != nil) != tt.wantErr {
				t.Fatalf("CancelOrder() error = %v, wantErr %v", err, tt.wantErr)
			}
			if gotPath != "/api/v1/order/cancel-transaction" {
				t.Fatalf("request path = %q", gotPath)
			}
			if gotRequest["trade_id"] != "trade-123" {
				t.Fatalf("trade_id = %v", gotRequest["trade_id"])
			}
			if gotRequest["signature"] == "" {
				t.Fatal("request signature is missing")
			}
		})
	}
}
