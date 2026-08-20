package usdtpayment

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

const maxUpstreamResponseBytes = 1 << 20

type Client struct {
	apiBase       string
	publicBaseURL string
	keyID         string
	secret        string
	httpClient    *http.Client
}

func NewClient(cfg *config.Config) *Client {
	timeout := 6 * time.Second
	if cfg != nil && cfg.USDTPayment.RequestTimeoutSeconds > 0 {
		timeout = time.Duration(cfg.USDTPayment.RequestTimeoutSeconds) * time.Second
	}
	client := &Client{httpClient: &http.Client{Timeout: timeout}}
	if cfg != nil {
		client.apiBase = strings.TrimRight(cfg.USDTPayment.APIBase, "/")
		client.publicBaseURL = strings.TrimRight(cfg.USDTPayment.PublicBaseURL, "/")
		client.keyID = cfg.USDTPayment.KeyID
		client.secret = cfg.USDTPayment.APISecret
	}
	return client
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
	var envelope upstreamEnvelope[UpstreamOrder]
	if err := c.call(ctx, http.MethodPost, "/api/v1/merchant/order/create", request, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != "ok" {
		return nil, fmt.Errorf("BEpusdt create returned code %q: %s", envelope.Code, envelope.Message)
	}
	c.rewritePaymentURL(&envelope.Data)
	return &envelope.Data, nil
}

func (c *Client) QueryOrder(ctx context.Context, merchantOrderID, providerTradeID string) (*UpstreamOrder, error) {
	request := map[string]string{"order_id": merchantOrderID, "trade_id": providerTradeID}
	var envelope upstreamEnvelope[UpstreamOrder]
	if err := c.call(ctx, http.MethodPost, "/api/v1/merchant/order/query", request, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != "ok" {
		return nil, fmt.Errorf("BEpusdt query returned code %q: %s", envelope.Code, envelope.Message)
	}
	c.rewritePaymentURL(&envelope.Data)
	return &envelope.Data, nil
}

func (c *Client) Capabilities(ctx context.Context, networks []string) ([]Capability, error) {
	var envelope upstreamEnvelope[struct {
		Networks []Capability `json:"networks"`
	}]
	if err := c.call(ctx, http.MethodPost, "/api/v1/merchant/capabilities", map[string]any{"networks": networks}, &envelope); err != nil {
		return nil, err
	}
	if envelope.Code != "ok" {
		return nil, fmt.Errorf("BEpusdt capabilities returned code %q: %s", envelope.Code, envelope.Message)
	}
	return envelope.Data.Networks, nil
}

func (c *Client) ExchangeRate(ctx context.Context) (*RateQuote, error) {
	var envelope upstreamEnvelope[RateQuote]
	if err := c.call(ctx, http.MethodPost, "/api/v1/merchant/rate", map[string]string{
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
	var envelope upstreamEnvelope[map[string]any]
	if err := c.call(ctx, http.MethodGet, "/api/v1/merchant/readiness", nil, &envelope); err != nil {
		return nil, err
	}
	return envelope.Data, nil
}

func (c *Client) call(ctx context.Context, method, path string, input, output any) error {
	body := []byte{}
	var err error
	if input != nil {
		body, err = json.Marshal(input)
		if err != nil {
			return fmt.Errorf("marshal BEpusdt request: %w", err)
		}
	}
	req, err := http.NewRequestWithContext(ctx, method, c.apiBase+path, bytes.NewReader(body))
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
	req.Header.Set(headerKeyID, c.keyID)
	req.Header.Set(headerTimestamp, timestamp)
	req.Header.Set(headerNonce, nonce)
	req.Header.Set(headerDigest, digest)
	req.Header.Set(headerSignature, hmacV2Sign(c.secret, method, path, timestamp, nonce, digest))
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call BEpusdt: %w", err)
	}
	defer resp.Body.Close()
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

func (c *Client) rewritePaymentURL(order *UpstreamOrder) {
	if order == nil || c.publicBaseURL == "" || order.TradeID == "" {
		return
	}
	order.PaymentURL = c.publicBaseURL + "/pay/checkout/" + url.PathEscape(order.TradeID)
}
