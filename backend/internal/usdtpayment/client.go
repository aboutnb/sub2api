package usdtpayment

import (
	"bytes"
	"context"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

const maxUpstreamResponseBytes = 1 << 20

type Client struct {
	config   config.USDTPaymentConfig
	resolver ConfigResolver
	// Legacy fields are retained for package-local tests and older constructors.
	apiBase       string
	publicBaseURL string
	keyID         string
	secret        string
	legacyToken   string
	httpClient    *http.Client
}

func NewClient(cfg *config.Config) *Client {
	timeout := 6 * time.Second
	if cfg != nil && cfg.USDTPayment.RequestTimeoutSeconds > 0 {
		timeout = time.Duration(cfg.USDTPayment.RequestTimeoutSeconds) * time.Second
	}
	client := &Client{httpClient: &http.Client{Timeout: timeout}}
	if cfg != nil {
		client.config = cfg.USDTPayment
		client.apiBase = strings.TrimRight(cfg.USDTPayment.APIBase, "/")
		client.publicBaseURL = strings.TrimRight(cfg.USDTPayment.PublicBaseURL, "/")
		client.keyID = cfg.USDTPayment.KeyID
		client.secret = cfg.USDTPayment.APISecret
		client.legacyToken = cfg.USDTPayment.LegacyToken
	}
	return client
}

func (c *Client) SetConfigResolver(resolver ConfigResolver) {
	if c != nil {
		c.resolver = resolver
	}
}

type upstreamEnvelope[T any] struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Data    T      `json:"data"`
}

type createUpstreamRequest struct {
	OrderID        string `json:"order_id"`
	Amount         string `json:"amount"`
	Fiat           string `json:"fiat"`
	TradeType      string `json:"trade_type"`
	NotifyURL      string `json:"notify_url"`
	RedirectURL    string `json:"redirect_url"`
	Name           string `json:"name"`
	TimeoutSeconds int    `json:"timeout_seconds"`
	// Rate is optional for older BEpusdt forks. The integrated merchant API
	// uses it to freeze the same rate returned by ExchangeRate.
	Rate string `json:"rate,omitempty"`
}

func (c *Client) CreateOrder(ctx context.Context, request createUpstreamRequest) (*UpstreamOrder, error) {
	cfg, err := c.effectiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	var envelope upstreamEnvelope[UpstreamOrder]
	if err := c.call(ctx, cfg, http.MethodPost, "/api/v1/merchant/order/create", request, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != "ok" {
		return nil, fmt.Errorf("BEpusdt create returned code %q: %s", envelope.Code, envelope.Message)
	}
	c.rewritePaymentURL(&envelope.Data, cfg.PublicBaseURL)
	return &envelope.Data, nil
}

func (c *Client) QueryOrder(ctx context.Context, merchantOrderID, providerTradeID string) (*UpstreamOrder, error) {
	cfg, err := c.effectiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	request := map[string]string{"order_id": merchantOrderID, "trade_id": providerTradeID}
	var envelope upstreamEnvelope[UpstreamOrder]
	if err := c.call(ctx, cfg, http.MethodPost, "/api/v1/merchant/order/query", request, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != "ok" {
		return nil, fmt.Errorf("BEpusdt query returned code %q: %s", envelope.Code, envelope.Message)
	}
	c.rewritePaymentURL(&envelope.Data, cfg.PublicBaseURL)
	return &envelope.Data, nil
}

func (c *Client) Capabilities(ctx context.Context, networks []string) ([]Capability, error) {
	cfg, err := c.effectiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	var envelope upstreamEnvelope[struct {
		Networks []Capability `json:"networks"`
	}]
	if err := c.call(ctx, cfg, http.MethodPost, "/api/v1/merchant/capabilities", map[string]any{"networks": networks}, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != "ok" {
		return nil, fmt.Errorf("BEpusdt capabilities returned code %q: %s", envelope.Code, envelope.Message)
	}
	return envelope.Data.Networks, nil
}

func (c *Client) ExchangeRate(ctx context.Context) (*RateQuote, error) {
	cfg, err := c.effectiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	var envelope upstreamEnvelope[RateQuote]
	if err := c.call(ctx, cfg, http.MethodPost, "/api/v1/merchant/rate", map[string]string{
		"crypto": "USDT",
		"fiat":   "CNY",
	}, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != "ok" {
		return nil, fmt.Errorf("BEpusdt rate returned code %q: %s", envelope.Code, envelope.Message)
	}
	return &envelope.Data, nil
}

func (c *Client) Readiness(ctx context.Context) (map[string]any, error) {
	cfg, err := c.effectiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	var envelope upstreamEnvelope[map[string]any]
	if err := c.call(ctx, cfg, http.MethodGet, "/api/v1/merchant/readiness", nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

func (c *Client) effectiveConfig(ctx context.Context) (config.USDTPaymentConfig, error) {
	if c == nil {
		return config.USDTPaymentConfig{}, fmt.Errorf("BEpusdt client is unavailable")
	}
	if c.resolver != nil {
		resolved, err := c.resolver.EffectiveConfig(ctx)
		if err != nil {
			return config.USDTPaymentConfig{}, err
		}
		return resolved, nil
	}
	resolved := c.config
	if resolved.APIBase == "" {
		resolved.APIBase = c.apiBase
	}
	if resolved.PublicBaseURL == "" {
		resolved.PublicBaseURL = c.publicBaseURL
	}
	if resolved.KeyID == "" {
		resolved.KeyID = c.keyID
	}
	if resolved.APISecret == "" {
		resolved.APISecret = c.secret
	}
	if resolved.LegacyToken == "" {
		resolved.LegacyToken = c.legacyToken
	}
	// Older BEpusdt builds use the legacy API token as the merchant HMAC
	// secret when no dedicated HMAC secret is configured. Keep that deployment
	// compatible while preferring the explicitly configured secret.
	if resolved.APISecret == "" {
		resolved.APISecret = resolved.LegacyToken
	}
	return resolved, nil
}

type cashierCreateData struct {
	Fiat           string `json:"fiat"`
	TradeID        string `json:"trade_id"`
	OrderID        string `json:"order_id"`
	Status         int    `json:"status"`
	Amount         string `json:"amount"`
	ExpirationTime int64  `json:"expiration_time"`
	PaymentURL     string `json:"payment_url"`
}

type cashierInfoData struct {
	Network      any    `json:"network"`
	TradeID      string `json:"trade_id"`
	OrderID      string `json:"order_id"`
	TradeType    string `json:"trade_type"`
	Status       int    `json:"status"`
	Money        string `json:"money"`
	ActualAmount string `json:"actual_amount"`
	Token        string `json:"token"`
	Fiat         string `json:"fiat"`
	ExpiredAt    int64  `json:"expired_at"`
	CreatedAt    int64  `json:"created_at"`
	TradeURL     string `json:"trade_url"`
	BlockTx      string `json:"block_transaction_id"`
}

func (c *Client) CreateCashierOrder(ctx context.Context, request map[string]any) (*cashierCreateData, error) {
	cfg, err := c.effectiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		StatusCode int               `json:"status_code"`
		Message    string            `json:"message"`
		Data       cashierCreateData `json:"data"`
	}
	if err := c.legacyCall(ctx, cfg, http.MethodPost, "/api/v1/order/create-order", request, &envelope); err != nil {
		return nil, err
	}
	if envelope.StatusCode != 200 {
		return nil, fmt.Errorf("BEpusdt cashier create returned %d: %s", envelope.StatusCode, envelope.Message)
	}
	if envelope.Data.TradeID == "" || envelope.Data.OrderID == "" {
		return nil, errors.New("BEpusdt cashier response is missing order identity")
	}
	// BEpusdt may derive this URL from the container Host header. Always use
	// the configured public base so browsers do not receive an internal Docker
	// hostname.
	if cfg.PublicBaseURL != "" {
		envelope.Data.PaymentURL = strings.TrimRight(cfg.PublicBaseURL, "/") + "/pay/checkout/" + url.PathEscape(envelope.Data.TradeID)
	}
	return &envelope.Data, nil
}

func (c *Client) CashierInfo(ctx context.Context, tradeID string) (*cashierInfoData, error) {
	cfg, err := c.effectiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	var envelope struct {
		StatusCode int             `json:"status_code"`
		Message    string          `json:"message"`
		Data       cashierInfoData `json:"data"`
	}
	if err := c.legacyCall(ctx, cfg, http.MethodPost, "/api/v1/pay/info", map[string]any{"trade_id": tradeID}, &envelope); err != nil {
		return nil, err
	}
	if envelope.StatusCode != 200 {
		return nil, fmt.Errorf("BEpusdt cashier info returned %d: %s", envelope.StatusCode, envelope.Message)
	}
	return &envelope.Data, nil
}

// CancelOrder stops a pending order in BEpusdt. The native endpoint is also
// valid for merchant-created orders because both flows share the same order
// store and legacy signing contract.
func (c *Client) CancelOrder(ctx context.Context, tradeID string) error {
	tradeID = strings.TrimSpace(tradeID)
	if tradeID == "" {
		return errors.New("BEpusdt trade id is empty")
	}
	cfg, err := c.effectiveConfig(ctx)
	if err != nil {
		return err
	}
	var envelope struct {
		StatusCode int    `json:"status_code"`
		Message    string `json:"message"`
	}
	if err := c.legacyCall(ctx, cfg, http.MethodPost, "/api/v1/order/cancel-transaction", map[string]any{"trade_id": tradeID}, &envelope); err != nil {
		return err
	}
	if envelope.StatusCode == http.StatusOK {
		return nil
	}
	// Cancellation is idempotent from Sub2API's perspective. If BEpusdt no
	// longer has the order, there is no upstream payment left to reconcile.
	message := strings.TrimSpace(envelope.Message)
	if strings.Contains(strings.ToLower(message), "order not found") || strings.Contains(message, "订单不存在") {
		return nil
	}
	if message == "" {
		message = "unknown upstream cancellation error"
	}
	return fmt.Errorf("BEpusdt cancel returned %d: %s", envelope.StatusCode, message)
}

func (c *Client) legacyCall(ctx context.Context, cfg config.USDTPaymentConfig, method, path string, input map[string]any, output any) error {
	request := make(map[string]any, len(input)+1)
	for key, value := range input {
		request[key] = value
	}
	request["signature"] = epusdtSign(request, cfg.LegacyToken)
	body, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("marshal BEpusdt cashier request: %w", err)
	}
	requestCtx := ctx
	cancel := func() {}
	if cfg.RequestTimeoutSeconds > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	}
	defer cancel()
	req, err := http.NewRequestWithContext(requestCtx, method, strings.TrimRight(cfg.APIBase, "/")+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build BEpusdt cashier request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call BEpusdt cashier: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read BEpusdt cashier response: %w", err)
	}
	if len(responseBody) > maxUpstreamResponseBytes {
		return errors.New("BEpusdt cashier response too large")
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("BEpusdt cashier HTTP %d", resp.StatusCode)
	}
	if err := json.Unmarshal(responseBody, output); err != nil {
		return fmt.Errorf("decode BEpusdt cashier response: %w", err)
	}
	return nil
}

func epusdtSign(data map[string]any, token string) string {
	keys := make([]string, 0, len(data))
	for key := range data {
		if key != "signature" {
			keys = append(keys, key)
		}
	}
	sort.Strings(keys)
	unsigned := ""
	for _, key := range keys {
		value := data[key]
		if value == nil || value == "" {
			continue
		}
		unsigned += key + "=" + fmt.Sprintf("%v", value) + "&"
	}
	unsigned = strings.TrimSuffix(unsigned, "&") + token
	sum := md5.Sum([]byte(unsigned))
	return fmt.Sprintf("%x", sum)
}

func (c *Client) call(ctx context.Context, cfg config.USDTPaymentConfig, method, path string, input, output any) error {
	requestCtx := ctx
	cancel := func() {}
	if cfg.RequestTimeoutSeconds > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, time.Duration(cfg.RequestTimeoutSeconds)*time.Second)
	}
	defer cancel()
	body := []byte{}
	var err error
	if input != nil {
		body, err = json.Marshal(input)
		if err != nil {
			return fmt.Errorf("marshal BEpusdt request: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(requestCtx, method, strings.TrimRight(cfg.APIBase, "/")+path, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build BEpusdt request: %w", err)
	}
	nonce, err := secureNonce()
	if err != nil {
		return err
	}
	timestamp := strconv.FormatInt(time.Now().Unix(), 10)
	digest := sha256Hex(body)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(headerKeyID, cfg.KeyID)
	req.Header.Set(headerTimestamp, timestamp)
	req.Header.Set(headerNonce, nonce)
	req.Header.Set(headerDigest, digest)
	req.Header.Set(headerSignature, hmacV2Sign(merchantSecret(cfg), method, path, timestamp, nonce, digest))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call BEpusdt: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxUpstreamResponseBytes+1))
	if err != nil {
		return fmt.Errorf("read BEpusdt response: %w", err)
	}
	if len(responseBody) > maxUpstreamResponseBytes {
		return fmt.Errorf("BEpusdt response exceeds %d bytes", maxUpstreamResponseBytes)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var upstreamError struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		}
		_ = json.Unmarshal(responseBody, &upstreamError)
		return fmt.Errorf("BEpusdt HTTP %d (%s): %s", resp.StatusCode, upstreamError.Code, upstreamError.Message)
	}
	if err := json.Unmarshal(responseBody, output); err != nil {
		return fmt.Errorf("decode BEpusdt response: %w", err)
	}
	return nil
}

func merchantSecret(cfg config.USDTPaymentConfig) string {
	if strings.TrimSpace(cfg.APISecret) != "" {
		return cfg.APISecret
	}
	return cfg.LegacyToken
}

func (c *Client) rewritePaymentURL(order *UpstreamOrder, publicBaseURL string) {
	if order == nil || publicBaseURL == "" || order.TradeID == "" {
		return
	}
	order.PaymentURL = strings.TrimRight(publicBaseURL, "/") + "/pay/checkout/" + url.PathEscape(order.TradeID)
}
