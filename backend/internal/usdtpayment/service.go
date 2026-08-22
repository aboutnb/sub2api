package usdtpayment

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Wei-Shaw/sub2api/internal/config"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/shopspring/decimal"
)

const (
	webhookPath        = "/api/v1/usdt/webhook/bepusdt"
	reconcileLease     = 15 * time.Second
	maxWebhookBodySize = 1 << 20
)

type Service struct {
	config       config.USDTPaymentConfig
	client       *Client
	repository   *Repository
	bridge       PaymentBridge
	modeProvider interface{ GetUSDTPaymentCheckoutMode(context.Context) string }
	stop         chan struct{}
	stopOnce     sync.Once
	wg           sync.WaitGroup
}

// SetCheckoutModeProvider connects the DB-backed admin setting without making
// the USDT module depend on the broader settings service at construction time.
func (s *Service) SetCheckoutModeProvider(provider interface{ GetUSDTPaymentCheckoutMode(context.Context) string }) {
	if s != nil {
		s.modeProvider = provider
	}
}

func (s *Service) checkoutMode(ctx context.Context) string {
	if s != nil && s.modeProvider != nil {
		return s.modeProvider.GetUSDTPaymentCheckoutMode(ctx)
	}
	if strings.EqualFold(strings.TrimSpace(s.config.CheckoutMode), "cashier") {
		return "cashier"
	}
	return "fixed"
}

func NewService(cfg *config.Config, client *Client, repository *Repository, bridge PaymentBridge) *Service {
	service := &Service{client: client, repository: repository, bridge: bridge, stop: make(chan struct{})}
	if cfg != nil {
		service.config = cfg.USDTPayment
	}
	if service.config.Enabled {
		service.wg.Add(1)
		go service.runReconciler()
	}
	return service
}

func (s *Service) Stop() {
	if s == nil {
		return
	}
	s.stopOnce.Do(func() { close(s.stop) })
	s.wg.Wait()
}

func (s *Service) Enabled() bool {
	return s != nil && s.config.Enabled
}

func (s *Service) CheckoutMode(ctx context.Context) string { return s.checkoutMode(ctx) }

func (s *Service) Capabilities(ctx context.Context) ([]Capability, error) {
	if !s.Enabled() {
		return []Capability{}, nil
	}
	items, err := s.client.Capabilities(ctx, s.config.EnabledNetworks)
	if err != nil {
		return nil, err
	}
	allowed := s.allowedNetworks()
	filtered := make([]Capability, 0, len(items))
	for _, item := range items {
		network := strings.ToLower(strings.TrimSpace(item.Network))
		if allowed[network] && item.Crypto == "USDT" && item.TradeType == networkTradeTypes[network] {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}

func (s *Service) ExchangeRate(ctx context.Context) (*RateQuote, error) {
	if !s.Enabled() {
		return nil, infraerrors.Forbidden("USDT_PAYMENT_DISABLED", "USDT payment is disabled")
	}
	quote, err := s.client.ExchangeRate(ctx)
	if err != nil {
		return nil, fmt.Errorf("get BEpusdt exchange rate: %w", err)
	}
	if quote == nil || quote.Crypto != "USDT" || quote.Fiat != s.config.Fiat || !positiveDecimal(quote.Rate) {
		return nil, errors.New("BEpusdt returned an invalid USDT exchange rate")
	}
	return quote, nil
}

func (s *Service) CreateOrder(ctx context.Context, userID int64, req CreateRequest, clientIP, sourceHost, sourceURL, locale string) (*CheckoutOrder, error) {
	if !s.Enabled() {
		return nil, infraerrors.Forbidden("USDT_PAYMENT_DISABLED", "USDT payment is disabled")
	}
	if unit := strings.TrimSpace(req.AmountUnit); unit != "" && !strings.EqualFold(unit, "USDT") {
		return nil, infraerrors.BadRequest("USDT_AMOUNT_UNIT_INVALID", "USDT orders must use USDT amounts")
	}
	if s.checkoutMode(ctx) == "cashier" {
		return s.createCashierOrder(ctx, userID, req, clientIP, sourceHost, sourceURL, locale)
	}
	network := strings.ToLower(strings.TrimSpace(req.Network))
	tradeType, ok := networkTradeTypes[network]
	if !ok || !s.allowedNetworks()[network] {
		return nil, infraerrors.BadRequest("USDT_NETWORK_DISABLED", "selected USDT network is not enabled")
	}
	capabilities, err := s.Capabilities(ctx)
	if err != nil {
		return nil, fmt.Errorf("check BEpusdt capabilities: %w", err)
	}
	ready := false
	var unavailableReason string
	for _, capability := range capabilities {
		if strings.EqualFold(capability.Network, network) {
			ready = capability.AcceptingOrders
			unavailableReason = capability.Reason
			break
		}
	}
	if !ready {
		if unavailableReason == "" {
			unavailableReason = "network is not reported by BEpusdt"
		}
		return nil, infraerrors.ServiceUnavailable("USDT_NETWORK_NOT_READY", unavailableReason)
	}
	rateQuote, err := s.ExchangeRate(ctx)
	if err != nil {
		return nil, infraerrors.ServiceUnavailable("USDT_RATE_UNAVAILABLE", err.Error())
	}
	requestedCrypto := decimal.NewFromFloat(req.Amount).Truncate(8)
	if req.Amount <= 0 || requestedCrypto.LessThanOrEqual(decimal.Zero) {
		return nil, infraerrors.BadRequest("USDT_AMOUNT_INVALID", "USDT amount must be greater than zero")
	}
	rate, err := decimal.NewFromString(rateQuote.Rate)
	if err != nil || rate.LessThanOrEqual(decimal.Zero) {
		return nil, infraerrors.ServiceUnavailable("USDT_RATE_UNAVAILABLE", "BEpusdt returned an invalid exchange rate")
	}
	fiatAmount := requestedCrypto.Mul(rate).Round(8)
	if fiatAmount.LessThanOrEqual(decimal.Zero) {
		return nil, infraerrors.BadRequest("USDT_AMOUNT_INVALID", "USDT amount is too small for CNY settlement")
	}
	prepared, err := s.bridge.PrepareUSDTOrder(ctx, PrepareOrderRequest{
		UserID: userID, Amount: fiatAmount.InexactFloat64(), TargetPayAmount: fiatAmount.InexactFloat64(), OrderType: req.OrderType, PlanID: req.PlanID,
		ClientIP: clientIP, SourceHost: sourceHost, SourceURL: sourceURL,
		PaymentSource: req.PaymentSource, Locale: locale,
	})
	if err != nil {
		return nil, err
	}
	redirectURL := strings.TrimSpace(req.ReturnURL)
	if !validReturnURL(redirectURL) {
		redirectURL = s.config.PublicCallbackBaseURL + "/payment"
	}
	upstream, err := s.client.CreateOrder(ctx, createUpstreamRequest{
		OrderID: prepared.MerchantOrderID, Amount: prepared.FiatAmount, Fiat: s.config.Fiat,
		TradeType: tradeType, NotifyURL: s.config.PublicCallbackBaseURL + webhookPath,
		RedirectURL: redirectURL, Name: "Sub2API USDT payment", TimeoutSeconds: s.config.OrderTimeoutSeconds,
		Rate: rateQuote.Rate,
	})
	if err != nil {
		_ = s.bridge.FailUSDTOrderBeforeQuote(ctx, prepared.ID, err)
		return nil, fmt.Errorf("create BEpusdt order: %w", err)
	}
	quote, err := s.quoteFromUpstream(prepared.ID, upstream, network, tradeType, prepared.FiatAmount, requestedCrypto.String(), rateQuote.Rate)
	if err != nil {
		_ = s.bridge.FailUSDTOrderBeforeQuote(ctx, prepared.ID, err)
		return nil, err
	}
	if err := s.repository.SaveQuote(ctx, quote); err != nil {
		_ = s.bridge.FailUSDTOrderBeforeQuote(ctx, prepared.ID, err)
		return nil, err
	}
	return checkoutFromQuote(quote, prepared.BaseAmount, prepared.PayAmount, prepared.FeeRate, "PENDING"), nil
}

func (s *Service) createCashierOrder(ctx context.Context, userID int64, req CreateRequest, clientIP, sourceHost, sourceURL, locale string) (*CheckoutOrder, error) {
	if strings.TrimSpace(s.config.LegacyToken) == "" {
		return nil, infraerrors.ServiceUnavailable("USDT_CASHIER_NOT_CONFIGURED", "BEpusdt legacy API token is not configured")
	}
	if req.Amount <= 0 {
		return nil, infraerrors.BadRequest("USDT_AMOUNT_INVALID", "USDT amount must be greater than zero")
	}
	rateQuote, err := s.ExchangeRate(ctx)
	if err != nil {
		return nil, infraerrors.ServiceUnavailable("USDT_RATE_UNAVAILABLE", err.Error())
	}
	rate, err := decimal.NewFromString(rateQuote.Rate)
	if err != nil || rate.LessThanOrEqual(decimal.Zero) {
		return nil, infraerrors.ServiceUnavailable("USDT_RATE_UNAVAILABLE", "BEpusdt returned an invalid exchange rate")
	}
	requestedCrypto := decimal.NewFromFloat(req.Amount).Truncate(8)
	fiatAmount := requestedCrypto.Mul(rate).Round(8)
	prepared, err := s.bridge.PrepareUSDTOrder(ctx, PrepareOrderRequest{
		UserID: userID, Amount: fiatAmount.InexactFloat64(), TargetPayAmount: fiatAmount.InexactFloat64(),
		OrderType: req.OrderType, PlanID: req.PlanID, ClientIP: clientIP, SourceHost: sourceHost,
		SourceURL: sourceURL, PaymentSource: req.PaymentSource, Locale: locale,
	})
	if err != nil {
		return nil, err
	}
	redirectURL := strings.TrimSpace(req.ReturnURL)
	if !validReturnURL(redirectURL) {
		redirectURL = s.config.PublicCallbackBaseURL + "/payment"
	}
	// Use the same public callback path as fixed mode; the handler dispatches
	// HMAC-v2 versus legacy MD5 by request headers/body signature.
	notifyURL := s.config.PublicCallbackBaseURL + webhookPath
	upstream, err := s.client.CreateCashierOrder(ctx, map[string]any{
		"order_id": prepared.MerchantOrderID, "notify_url": notifyURL, "redirect_url": redirectURL,
		"amount": fiatAmount.InexactFloat64(), "name": "Sub2API USDT payment", "fiat": s.config.Fiat,
		"timeout": s.config.OrderTimeoutSeconds,
	})
	if err != nil {
		_ = s.bridge.FailUSDTOrderBeforeQuote(ctx, prepared.ID, err)
		return nil, fmt.Errorf("create BEpusdt cashier order: %w", err)
	}
	createdAt := time.Now().UTC()
	expiresAt := createdAt.Add(time.Duration(s.config.OrderTimeoutSeconds) * time.Second)
	if upstream.ExpirationTime > 0 {
		expiresAt = createdAt.Add(time.Duration(upstream.ExpirationTime) * time.Second)
	}
	quote := &Quote{
		PaymentOrderID: prepared.ID, MerchantOrderID: upstream.OrderID, ProviderTradeID: upstream.TradeID,
		FiatCurrency: s.config.Fiat, FiatAmount: canonicalDecimal(fiatAmount.String()), CryptoCurrency: "USDT",
		Network: "pending", TradeType: "pending", CryptoAmount: canonicalDecimal(requestedCrypto.String()),
		ExchangeRate: canonicalDecimal(rateQuote.Rate), ReceivingAddress: "pending", PaymentURL: upstream.PaymentURL,
		UpstreamCreatedAt: createdAt, UpstreamExpiresAt: expiresAt, ProviderStatus: "waiting",
	}
	if err := s.repository.SaveQuote(ctx, quote); err != nil {
		_ = s.bridge.FailUSDTOrderBeforeQuote(ctx, prepared.ID, err)
		return nil, err
	}
	return checkoutFromQuote(quote, prepared.BaseAmount, prepared.PayAmount, prepared.FeeRate, "PENDING"), nil
}

func (s *Service) quoteFromUpstream(paymentOrderID int64, upstream *UpstreamOrder, network, tradeType, fiatAmount, expectedCryptoAmount, expectedRate string) (*Quote, error) {
	if upstream == nil {
		return nil, errors.New("BEpusdt returned an empty order")
	}
	if upstream.OrderID == "" || upstream.TradeID == "" || upstream.PaymentURL == "" {
		return nil, errors.New("BEpusdt quote is missing order identity or payment URL")
	}
	if !equalDecimal(upstream.Amount, fiatAmount) || !strings.EqualFold(upstream.Fiat, s.config.Fiat) {
		return nil, fmt.Errorf("BEpusdt quote fiat mismatch: expected %s %s, got %s %s", fiatAmount, s.config.Fiat, upstream.Amount, upstream.Fiat)
	}
	if upstream.Crypto != "USDT" || upstream.Network != network || upstream.TradeType != tradeType {
		return nil, errors.New("BEpusdt quote currency or network mismatch")
	}
	if !positiveDecimal(upstream.ActualAmount) || !positiveDecimal(upstream.ExchangeRate) || strings.TrimSpace(upstream.Token) == "" {
		return nil, errors.New("BEpusdt quote contains an invalid amount, rate, or receiving address")
	}
	if expectedCryptoAmount != "" && !equalDecimal(upstream.ActualAmount, expectedCryptoAmount) {
		return nil, fmt.Errorf("BEpusdt quote USDT amount mismatch: expected %s, got %s", expectedCryptoAmount, upstream.ActualAmount)
	}
	if expectedRate != "" && !equalDecimal(upstream.ExchangeRate, expectedRate) {
		return nil, fmt.Errorf("BEpusdt quote exchange rate changed: expected %s, got %s", expectedRate, upstream.ExchangeRate)
	}
	createdAt := time.Unix(upstream.CreatedAt, 0)
	expiresAt := time.Unix(upstream.ExpiresAt, 0)
	if upstream.CreatedAt <= 0 || upstream.ExpiresAt <= upstream.CreatedAt || time.Until(expiresAt) <= 0 {
		return nil, errors.New("BEpusdt quote contains an invalid expiry window")
	}
	return &Quote{
		PaymentOrderID: paymentOrderID, MerchantOrderID: upstream.OrderID, ProviderTradeID: upstream.TradeID,
		FiatCurrency: strings.ToUpper(upstream.Fiat), FiatAmount: canonicalDecimal(upstream.Amount),
		CryptoCurrency: "USDT", Network: network, TradeType: tradeType,
		CryptoAmount: canonicalDecimal(upstream.ActualAmount), ExchangeRate: canonicalDecimal(upstream.ExchangeRate),
		ReceivingAddress: upstream.Token, PaymentURL: upstream.PaymentURL,
		UpstreamCreatedAt: createdAt, UpstreamExpiresAt: expiresAt, ProviderStatus: upstream.StatusName,
	}, nil
}

func (s *Service) GetOrder(ctx context.Context, userID, paymentOrderID int64) (*CheckoutOrder, error) {
	quote, status, amount, payAmount, feeRate, err := s.repository.GetQuoteForUser(ctx, paymentOrderID, userID)
	if err != nil {
		if isNotFound(err) {
			return nil, infraerrors.NotFound("USDT_ORDER_NOT_FOUND", "USDT payment order not found")
		}
		return nil, err
	}
	if quote.ProviderStatus == "waiting" || quote.ProviderStatus == "confirming" {
		_ = s.reconcileQuote(ctx, quote)
		quote, status, amount, payAmount, feeRate, err = s.repository.GetQuoteForUser(ctx, paymentOrderID, userID)
		if err != nil {
			return nil, err
		}
	}
	return checkoutFromQuote(quote, amount, payAmount, feeRate, status), nil
}

func checkoutFromQuote(q *Quote, amount, payAmount, feeRate float64, status string) *CheckoutOrder {
	mode := "fixed"
	if q.PaymentMode == "cashier" || q.TradeType == "pending" {
		mode = "cashier"
	}
	return &CheckoutOrder{
		OrderID: q.PaymentOrderID, OutTradeNo: q.MerchantOrderID, Amount: amount, PayAmount: payAmount,
		FeeRate: feeRate, Status: status, PaymentType: PaymentType, PaymentMode: mode,
		FiatCurrency: q.FiatCurrency, FiatAmount: q.FiatAmount, CryptoCurrency: q.CryptoCurrency,
		CryptoAmount: q.CryptoAmount, Network: q.Network, TradeType: q.TradeType,
		ReceivingAddress: q.ReceivingAddress, ExchangeRate: q.ExchangeRate, PaymentURL: q.PaymentURL,
		ExpiresAt: q.UpstreamExpiresAt, TransactionHash: q.TransactionHash, ChainTransferAt: q.ChainTransferAt,
	}
}

func (s *Service) VerifyWebhook(raw []byte, headers map[string]string, requestPath string, now time.Time) (*WebhookPayload, error) {
	if !s.Enabled() {
		return nil, errors.New("USDT payment is disabled")
	}
	if len(raw) == 0 || len(raw) > maxWebhookBodySize {
		return nil, errors.New("invalid webhook body size")
	}
	keyID := headers[strings.ToLower(headerKeyID)]
	timestamp := headers[strings.ToLower(headerTimestamp)]
	nonce := headers[strings.ToLower(headerNonce)]
	digest := strings.ToLower(headers[strings.ToLower(headerDigest)])
	signature := strings.ToLower(headers[strings.ToLower(headerSignature)])
	if keyID != s.config.KeyID || timestamp == "" || len(nonce) < 16 || digest == "" || signature == "" {
		return nil, errors.New("missing or invalid BEpusdt HMAC headers")
	}
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || absInt64(now.Unix()-ts) > int64(s.config.WebhookClockSkewSeconds) {
		return nil, errors.New("BEpusdt webhook timestamp outside allowed window")
	}
	if digest != sha256Hex(raw) {
		return nil, errors.New("BEpusdt webhook body digest mismatch")
	}
	if !equalHex(signature, hmacV2Sign(s.config.APISecret, "POST", requestPath, timestamp, nonce, digest)) {
		return nil, errors.New("BEpusdt webhook request signature mismatch")
	}
	var payload WebhookPayload
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&payload); err != nil {
		return nil, fmt.Errorf("decode BEpusdt webhook: %w", err)
	}
	if decoder.Decode(&struct{}{}) == nil {
		return nil, errors.New("BEpusdt webhook contains multiple JSON values")
	}
	payloadSignature := payload.Signature
	payload.Signature = ""
	unsigned, err := json.Marshal(payload)
	if err != nil || !equalHex(payloadSignature, hmacHex(s.config.APISecret, unsigned)) {
		return nil, errors.New("BEpusdt webhook payload signature mismatch")
	}
	payload.Signature = payloadSignature
	if payload.SignatureVersion != "v2" || payload.KeyID != s.config.KeyID || payload.EventType != "payment.succeeded" || payload.EventID == "" {
		return nil, errors.New("unsupported BEpusdt webhook event")
	}
	return &payload, nil
}

// HandleLegacyWebhook accepts BEpusdt's native checkout callback. It verifies
// the original MD5 body signature, enriches the callback with /pay/info so the
// selected network is frozen, and then enters the same idempotent webhook path.
func (s *Service) HandleLegacyWebhook(ctx context.Context, raw []byte) error {
	var body map[string]any
	if err := json.Unmarshal(raw, &body); err != nil {
		return errors.New("invalid native BEpusdt webhook JSON")
	}
	signature, _ := body["signature"].(string)
	if signature == "" || !strings.EqualFold(signature, epusdtSign(body, s.config.LegacyToken)) {
		return errors.New("invalid native BEpusdt webhook signature")
	}
	tradeID, _ := body["trade_id"].(string)
	orderID, _ := body["order_id"].(string)
	status := intFromAny(body["status"])
	if tradeID == "" || orderID == "" || status != 2 {
		return errors.New("unsupported native BEpusdt webhook")
	}
	info, err := s.client.CashierInfo(ctx, tradeID)
	if err != nil {
		return err
	}
	quote, err := s.repository.GetQuoteByMerchantOrderID(ctx, orderID)
	if err != nil {
		return err
	}
	actualAmount, _ := body["actual_amount"].(string)
	if actualAmount == "" {
		actualAmount = info.ActualAmount
	}
	token, _ := body["token"].(string)
	if token == "" {
		token = info.Token
	}
	network := normalizeCashierNetwork(info.Network, info.TradeType)
	if network == "" || info.TradeType == "" || token == "" {
		return errors.New("native BEpusdt webhook is missing selected network proof")
	}
	expiresAt := quote.UpstreamExpiresAt
	if info.ExpiredAt > 0 {
		expiresAt = time.Unix(info.ExpiredAt, 0)
	}
	if err := s.repository.UpdateCashierSelection(ctx, quote, network, info.TradeType, info.Money, actualAmount, token, "succeeded", expiresAt); err != nil {
		return err
	}
	txHash, _ := body["block_transaction_id"].(string)
	occurred := time.Now().Unix()
	payload := WebhookPayload{
		EventID: "native:" + tradeID + ":" + txHash, EventType: "payment.succeeded", OccurredAt: occurred,
		OrderID: orderID, TradeID: tradeID, Fiat: info.Fiat, Amount: info.Money, Crypto: "USDT", ActualAmount: actualAmount,
		TradeType: info.TradeType, Network: network, Token: token,
		BlockTransactionID: txHash, TransferAt: occurred, SignatureVersion: "legacy", KeyID: "native",
	}
	return s.HandleWebhook(ctx, payload, raw)
}

func intFromAny(value any) int {
	switch v := value.(type) {
	case float64:
		return int(v)
	case int:
		return v
	case string:
		n, _ := strconv.Atoi(v)
		return n
	}
	return 0
}

func (s *Service) HandleWebhook(ctx context.Context, payload WebhookPayload, raw []byte) error {
	inserted, err := s.repository.InsertWebhookEvent(ctx, payload, raw)
	if err != nil {
		return err
	}
	if !inserted {
		return nil
	}
	err = s.processWebhookPayload(ctx, payload)
	if markErr := s.repository.MarkWebhookEventByEventID(ctx, payload.EventID, err); markErr != nil && err == nil {
		return markErr
	}
	return err
}

func (s *Service) processWebhookPayload(ctx context.Context, payload WebhookPayload) error {
	quote, err := s.repository.GetQuoteByMerchantOrderID(ctx, payload.OrderID)
	if err != nil {
		return fmt.Errorf("find USDT quote for webhook: %w", err)
	}
	proof := &UpstreamOrder{
		OrderID: payload.OrderID, TradeID: payload.TradeID, Status: 2, StatusName: "succeeded",
		Fiat: payload.Fiat, Amount: payload.Amount, Crypto: payload.Crypto, ActualAmount: payload.ActualAmount,
		TradeType: payload.TradeType, Network: payload.Network, Token: payload.Token,
		BlockTransactionID: payload.BlockTransactionID, TransferAt: payload.TransferAt,
		Confirmation: map[string]any{"confirmed": true, "block_number": payload.BlockNumber},
	}
	return s.confirmProof(ctx, quote, proof)
}

func (s *Service) reconcileQuote(ctx context.Context, quote *Quote) error {
	if quote.TradeType == "pending" || quote.Network == "pending" {
		info, err := s.client.CashierInfo(ctx, quote.ProviderTradeID)
		if err != nil {
			_ = s.repository.RecordReconcile(ctx, quote.ID, quote.ProviderStatus, time.Now().Add(s.reconcileDelay(quote.ReconcileAttempts)), err)
			return err
		}
		status := mapCashierStatus(info.Status)
		if info.TradeType != "" && info.Network != "" && info.Token != "" {
			tradeType := info.TradeType
			network := normalizeCashierNetwork(info.Network, tradeType)
			expiresAt := quote.UpstreamExpiresAt
			if info.ExpiredAt > 0 {
				expiresAt = time.Unix(info.ExpiredAt, 0)
			}
			if err := s.repository.UpdateCashierSelection(ctx, quote, network, tradeType, info.Money, info.ActualAmount, info.Token, status, expiresAt); err != nil {
				return err
			}
			quote.Network, quote.TradeType, quote.FiatAmount, quote.CryptoAmount, quote.ReceivingAddress, quote.ProviderStatus, quote.UpstreamExpiresAt = network, tradeType, canonicalDecimal(info.Money), canonicalDecimal(info.ActualAmount), info.Token, status, expiresAt
		}
		if status != "succeeded" {
			return s.repository.RecordReconcile(ctx, quote.ID, status, time.Now().Add(s.reconcileDelay(quote.ReconcileAttempts)), nil)
		}
		if strings.TrimSpace(info.BlockTx) == "" {
			// Native /pay/info intentionally does not provide a chain proof. Wait
			// for the signed legacy callback carrying block_transaction_id.
			return s.repository.RecordReconcile(ctx, quote.ID, "confirming", time.Now().Add(2*time.Second), nil)
		}
		upstream := &UpstreamOrder{OrderID: info.OrderID, TradeID: info.TradeID, Status: info.Status, StatusName: status, Fiat: info.Fiat, Amount: info.Money, Crypto: "USDT", ActualAmount: info.ActualAmount, TradeType: info.TradeType, Network: normalizeCashierNetwork(info.Network, info.TradeType), Token: info.Token, BlockTransactionID: info.BlockTx, TransferAt: time.Now().Unix(), Confirmation: map[string]any{"confirmed": true}}
		if err := s.confirmProof(ctx, quote, upstream); err != nil {
			_ = s.repository.RecordReconcile(ctx, quote.ID, quote.ProviderStatus, time.Now().Add(2*time.Second), err)
			return err
		}
		return nil
	}
	upstream, err := s.client.QueryOrder(ctx, quote.MerchantOrderID, quote.ProviderTradeID)
	if err != nil {
		next := time.Now().Add(s.reconcileDelay(quote.ReconcileAttempts))
		_ = s.repository.RecordReconcile(ctx, quote.ID, quote.ProviderStatus, next, err)
		return err
	}
	switch upstream.StatusName {
	case "succeeded":
		if err := s.confirmProof(ctx, quote, upstream); err != nil {
			_ = s.repository.RecordReconcile(ctx, quote.ID, quote.ProviderStatus, time.Now().Add(2*time.Second), err)
			return err
		}
		return nil
	case "waiting", "confirming", "expired":
		status := upstream.StatusName
		if status == "expired" && time.Now().Before(quote.UpstreamExpiresAt.Add(time.Duration(s.config.LatePaymentWindowMinutes)*time.Minute)) {
			status = "waiting"
		}
		return s.repository.RecordReconcile(ctx, quote.ID, status, time.Now().Add(s.reconcileDelay(quote.ReconcileAttempts)), nil)
	default:
		return s.repository.RecordReconcile(ctx, quote.ID, upstream.StatusName, time.Now().Add(30*time.Second), nil)
	}
}

func mapCashierStatus(status int) string {
	switch status {
	case 2:
		return "succeeded"
	case 3:
		return "expired"
	case 1:
		return "waiting"
	default:
		return "waiting"
	}
}

func normalizeCashierNetwork(network any, tradeType string) string {
	n := ""
	switch value := network.(type) {
	case string:
		n = strings.ToLower(strings.TrimSpace(value))
	case map[string]any:
		if name, ok := value["network"].(string); ok {
			n = strings.ToLower(strings.TrimSpace(name))
		}
		if n == "" {
			if alias, ok := value["alias"].(string); ok {
				n = strings.ToLower(strings.TrimSpace(alias))
			}
		}
	}
	for key, value := range networkTradeTypes {
		if value == strings.ToLower(strings.TrimSpace(tradeType)) {
			if n == "" {
				return key
			}
			break
		}
	}
	return n
}

func (s *Service) confirmProof(ctx context.Context, quote *Quote, proof *UpstreamOrder) error {
	if quote == nil || proof == nil {
		return errors.New("USDT payment proof is empty")
	}
	if proof.OrderID != quote.MerchantOrderID || proof.TradeID != quote.ProviderTradeID {
		return errors.New("USDT payment proof order identity mismatch")
	}
	if !equalDecimal(proof.Amount, quote.FiatAmount) || !equalDecimal(proof.ActualAmount, quote.CryptoAmount) {
		return errors.New("USDT payment proof amount mismatch")
	}
	if !strings.EqualFold(proof.Fiat, quote.FiatCurrency) || proof.Crypto != quote.CryptoCurrency ||
		proof.Network != quote.Network || proof.TradeType != quote.TradeType || proof.Token != quote.ReceivingAddress {
		return errors.New("USDT payment proof frozen quote mismatch")
	}
	if proof.StatusName != "succeeded" || !confirmationIsTrue(proof.Confirmation) {
		return errors.New("USDT payment is not confirmed")
	}
	txHash := strings.TrimSpace(proof.BlockTransactionID)
	transferAt := time.Unix(proof.TransferAt, 0)
	if txHash == "" || proof.TransferAt <= 0 {
		return errors.New("USDT payment proof is missing transaction hash or transfer time")
	}
	if transferAt.Before(quote.UpstreamCreatedAt) || transferAt.After(quote.UpstreamExpiresAt) {
		return errors.New("USDT transfer occurred outside the frozen payment window")
	}
	recoveryDeadline := quote.UpstreamExpiresAt.Add(time.Duration(s.config.LatePaymentWindowMinutes) * time.Minute)
	if time.Now().After(recoveryDeadline) {
		return errors.New("USDT payment recovery window has elapsed")
	}
	blockNumber := confirmationBlockNumber(proof.Confirmation)
	if err := s.repository.ReserveConfirmation(ctx, quote, txHash, transferAt, blockNumber); err != nil {
		return err
	}
	quote.ProviderStatus = "confirming"
	quote.TransactionHash = &txHash
	quote.ChainTransferAt = &transferAt
	quote.BlockNumber = &blockNumber
	// Fulfill the local order before marking the quote as succeeded. If the
	// balance/subscription write fails, the quote remains in confirming state
	// and a webhook or polling retry can safely run the idempotent bridge again.
	if err := s.bridge.ConfirmUSDTOrder(ctx, ConfirmOrderRequest{
		PaymentOrderID: quote.PaymentOrderID, MerchantOrderID: quote.MerchantOrderID,
		ProviderTradeID: quote.ProviderTradeID, TransactionHash: txHash, FiatAmount: quote.FiatAmount,
		TransferAt: transferAt, QuoteExpiresAt: quote.UpstreamExpiresAt, RecoveryDeadline: recoveryDeadline,
	}); err != nil {
		return err
	}
	return s.repository.RecordConfirmed(ctx, quote, txHash, transferAt, blockNumber)
}

func (s *Service) runReconciler() {
	defer s.wg.Done()
	interval := time.Duration(s.config.ReconcileIntervalSeconds) * time.Second
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	s.runReconcileBatch()
	for {
		select {
		case <-ticker.C:
			s.runReconcileBatch()
		case <-s.stop:
			return
		}
	}
}

func (s *Service) runReconcileBatch() {
	ctx, cancel := context.WithTimeout(context.Background(), 12*time.Second)
	defer cancel()
	events, err := s.repository.ClaimWebhookEvents(ctx, s.config.ReconcileBatchSize)
	if err != nil {
		slog.Error("[USDT] claim webhook events", "error", err)
	} else {
		for _, event := range events {
			processErr := s.processWebhookPayload(ctx, event.Payload)
			_ = s.repository.CompleteWebhookEvent(ctx, event.ID, processErr)
		}
	}
	quotes, err := s.repository.ClaimDueQuotes(ctx, s.config.ReconcileBatchSize, reconcileLease)
	if err != nil {
		slog.Error("[USDT] claim reconciliation quotes", "error", err)
		return
	}
	for _, quote := range quotes {
		if err := s.reconcileQuote(ctx, quote); err != nil {
			slog.Warn("[USDT] reconcile quote", "order", quote.MerchantOrderID, "error", err)
		}
	}
}

func (s *Service) reconcileDelay(attempt int) time.Duration {
	seconds := s.config.ReconcileIntervalSeconds
	if attempt > 1 {
		seconds *= attempt
	}
	if seconds > 30 {
		seconds = 30
	}
	return time.Duration(seconds) * time.Second
}

func (s *Service) allowedNetworks() map[string]bool {
	allowed := make(map[string]bool, len(s.config.EnabledNetworks))
	for _, network := range s.config.EnabledNetworks {
		allowed[strings.ToLower(strings.TrimSpace(network))] = true
	}
	return allowed
}

func positiveDecimal(value string) bool {
	d, err := decimal.NewFromString(strings.TrimSpace(value))
	return err == nil && d.GreaterThan(decimal.Zero)
}

func equalDecimal(a, b string) bool {
	left, err := decimal.NewFromString(strings.TrimSpace(a))
	if err != nil {
		return false
	}
	right, err := decimal.NewFromString(strings.TrimSpace(b))
	return err == nil && left.Equal(right)
}

func canonicalDecimal(value string) string {
	d, err := decimal.NewFromString(strings.TrimSpace(value))
	if err != nil {
		return ""
	}
	return d.String()
}

func confirmationIsTrue(value map[string]any) bool {
	confirmed, ok := value["confirmed"].(bool)
	return ok && confirmed
}

func confirmationBlockNumber(value map[string]any) int64 {
	switch block := value["block_number"].(type) {
	case float64:
		return int64(block)
	case int64:
		return block
	case int:
		return int64(block)
	case json.Number:
		result, _ := block.Int64()
		return result
	}
	return 0
}

func validReturnURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.IsAbs() && parsed.Hostname() != "" && (parsed.Scheme == "https" || parsed.Scheme == "http")
}

func absInt64(value int64) int64 {
	if value < 0 {
		return -value
	}
	return value
}
