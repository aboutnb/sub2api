package service

import (
	"context"
	"fmt"
	"strings"
	"time"

	dbent "github.com/Wei-Shaw/sub2api/ent"
	"github.com/Wei-Shaw/sub2api/ent/paymentorder"
	"github.com/Wei-Shaw/sub2api/internal/payment"
	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/Wei-Shaw/sub2api/internal/usdtpayment"
	"github.com/shopspring/decimal"
)

// PrepareUSDTOrder reuses only Sub2API's local pricing, limits, and order
// creation. It deliberately does not select or invoke an RMB provider.
func (s *PaymentService) PrepareUSDTOrder(ctx context.Context, req usdtpayment.PrepareOrderRequest) (*usdtpayment.PreparedOrder, error) {
	orderType := strings.TrimSpace(req.OrderType)
	if orderType == "" {
		orderType = payment.OrderTypeBalance
	}
	localReq := CreateOrderRequest{
		UserID: req.UserID, Amount: req.Amount, PaymentType: usdtpayment.PaymentType,
		ClientIP: req.ClientIP, SrcHost: req.SourceHost, SrcURL: req.SourceURL,
		PaymentSource: req.PaymentSource, OrderType: orderType, PlanID: req.PlanID, Locale: req.Locale,
	}
	cfg, err := s.configService.GetPaymentConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("get payment config: %w", err)
	}
	if !cfg.Enabled {
		return nil, infraerrors.Forbidden("PAYMENT_DISABLED", "payment system is disabled")
	}
	plan, err := s.validateOrderInput(ctx, localReq, cfg)
	if err != nil {
		return nil, err
	}
	if err := s.checkCancelRateLimit(ctx, req.UserID, cfg); err != nil {
		return nil, err
	}
	user, err := s.userRepo.GetByID(ctx, req.UserID)
	if err != nil {
		return nil, fmt.Errorf("get user: %w", err)
	}
	if user.Status != payment.EntityStatusActive {
		return nil, infraerrors.Forbidden("USER_INACTIVE", "user account is disabled")
	}
	if s.notificationEmailService != nil {
		s.notificationEmailService.RememberRecipientLocale(ctx, req.UserID, user.Email, req.Locale)
	}
	baseAmount := req.Amount
	limitAmount := req.Amount
	if plan != nil {
		baseAmount = plan.Price
		limitAmount = plan.Price
	}
	// USDT is a separate crypto settlement rail. It never participates in the
	// generic fiat recharge fee settings; its optional incentive is applied only
	// to the credited balance below.
	feeRate := 0.0
	fiatAmount, payAmount, err := calculateCreateOrderPayAmountForOrderType(
		limitAmount, feeRate, payment.DefaultPaymentCurrency, orderType, cfg.SubscriptionUSDToCNYRate,
	)
	if req.TargetPayAmount > 0 {
		// USDT orders freeze the provider's CNY quote. Do not apply the RMB
		// fee a second time after the quote has already been converted. Derive
		// the credited base so the configured fee remains included in the
		// frozen provider amount.
		targetPay := decimal.NewFromFloat(req.TargetPayAmount).Round(8)
		fiatAmount = strings.TrimRight(strings.TrimRight(targetPay.StringFixed(8), "0"), ".")
		payAmount = targetPay.InexactFloat64()
		baseAmount = baseAmountForTargetPay(payAmount, feeRate)
	}
	if err != nil {
		return nil, err
	}
	orderAmount := baseAmount
	if orderType == payment.OrderTypeBalance {
		orderAmount = calculateUSDTBalanceCreditedAmount(baseAmount, cfg.USDTPaymentBonusPercent)
	}
	selection := &payment.InstanceSelection{
		ProviderKey: usdtpayment.ProviderKey, SupportedTypes: usdtpayment.PaymentType,
		PaymentMode: "redirect", Config: map[string]string{"currency": payment.DefaultPaymentCurrency},
	}
	order, err := s.createOrderInTx(ctx, localReq, user, plan, cfg, orderAmount, limitAmount, feeRate, payAmount, selection)
	if err != nil {
		return nil, err
	}
	return &usdtpayment.PreparedOrder{
		ID: order.ID, MerchantOrderID: order.OutTradeNo, FiatAmount: fiatAmount,
		BaseAmount: order.Amount, PayAmount: order.PayAmount, FeeRate: order.FeeRate,
		OrderType: order.OrderType, ExpiresAt: order.ExpiresAt,
	}, nil
}

func calculateUSDTBalanceCreditedAmount(baseAmount, bonusPercent float64) float64 {
	if baseAmount <= 0 || bonusPercent <= 0 {
		return baseAmount
	}
	bonus := decimal.NewFromFloat(bonusPercent).Div(decimal.NewFromInt(100))
	return decimal.NewFromFloat(baseAmount).Mul(decimal.NewFromInt(1).Add(bonus)).Round(2).InexactFloat64()
}

// baseAmountForTargetPay derives the balance principal from the frozen total.
// The provider total remains the exact high-precision target; the local
// balance principal is kept at the site's normal cent precision.
func baseAmountForTargetPay(target, feeRate float64) float64 {
	if target <= 0 || feeRate <= 0 {
		return target
	}
	return decimal.NewFromFloat(target).
		Div(decimal.NewFromFloat(1).Add(decimal.NewFromFloat(feeRate).Div(decimal.NewFromInt(100)))).
		Round(2).
		InexactFloat64()
}

// ConfirmUSDTOrder accepts only an already verified, frozen USDT proof. Amount
// comparison remains decimal-exact and does not use the RMB webhook tolerance.
func (s *PaymentService) ConfirmUSDTOrder(ctx context.Context, req usdtpayment.ConfirmOrderRequest) error {
	order, err := s.entClient.PaymentOrder.Get(ctx, req.PaymentOrderID)
	if err != nil {
		if dbent.IsNotFound(err) {
			return infraerrors.NotFound("USDT_ORDER_NOT_FOUND", "USDT payment order not found")
		}
		return err
	}
	if order.OutTradeNo != req.MerchantOrderID || order.PaymentType != usdtpayment.PaymentType ||
		strings.TrimSpace(psStringValue(order.ProviderKey)) != usdtpayment.ProviderKey {
		return fmt.Errorf("USDT payment order identity or provider mismatch")
	}
	expectedFiat := decimal.NewFromFloat(order.PayAmount).String()
	if !usdtBridgeEqualDecimal(expectedFiat, req.FiatAmount) {
		s.writeAuditLog(ctx, order.ID, "USDT_AMOUNT_MISMATCH", usdtpayment.ProviderKey, map[string]any{
			"expected": expectedFiat, "actual": req.FiatAmount, "tradeNo": req.ProviderTradeID,
		})
		return fmt.Errorf("USDT fiat amount mismatch: expected %s, got %s", expectedFiat, req.FiatAmount)
	}
	if req.TransactionHash == "" || req.TransferAt.IsZero() || req.QuoteExpiresAt.IsZero() || req.RecoveryDeadline.IsZero() {
		return fmt.Errorf("USDT payment proof is incomplete")
	}
	if req.TransferAt.After(req.QuoteExpiresAt) || time.Now().After(req.RecoveryDeadline) {
		return fmt.Errorf("USDT payment proof is outside the accepted time window")
	}
	if !order.ExpiresAt.Equal(req.QuoteExpiresAt) {
		return fmt.Errorf("USDT quote expiry does not match the local order")
	}
	switch order.Status {
	case OrderStatusCompleted, OrderStatusRefunded:
		return nil
	case OrderStatusPaid, OrderStatusRecharging, OrderStatusFailed:
		return s.executeFulfillment(ctx, order.ID)
	case OrderStatusPending, OrderStatusCancelled, OrderStatusExpired:
	default:
		return fmt.Errorf("USDT order cannot be confirmed from status %s", order.Status)
	}
	updated, err := s.entClient.PaymentOrder.Update().Where(
		paymentorder.IDEQ(order.ID),
		paymentorder.StatusIn(OrderStatusPending, OrderStatusCancelled, OrderStatusExpired),
	).SetStatus(OrderStatusPaid).
		SetPaymentTradeNo(req.TransactionHash).
		SetPaidAt(req.TransferAt).
		ClearFailedAt().
		ClearFailedReason().
		Save(ctx)
	if err != nil {
		return fmt.Errorf("mark USDT order paid: %w", err)
	}
	if updated == 0 {
		current, getErr := s.entClient.PaymentOrder.Get(ctx, order.ID)
		if getErr != nil {
			return getErr
		}
		if current.Status == OrderStatusCompleted || current.Status == OrderStatusRefunded {
			return nil
		}
		if current.Status == OrderStatusPaid || current.Status == OrderStatusRecharging || current.Status == OrderStatusFailed {
			return s.executeFulfillment(ctx, current.ID)
		}
		return fmt.Errorf("USDT order status changed to %s during confirmation", current.Status)
	}
	s.writeAuditLog(ctx, order.ID, "ORDER_PAID", usdtpayment.ProviderKey, map[string]any{
		"tradeNo": req.ProviderTradeID, "transactionHash": req.TransactionHash,
		"paidAmount": req.FiatAmount, "transferAt": req.TransferAt,
	})
	return s.executeFulfillment(ctx, order.ID)
}

func (s *PaymentService) FailUSDTOrderBeforeQuote(ctx context.Context, orderID int64, cause error) error {
	reason := "failed to create USDT quote"
	if cause != nil {
		reason = cause.Error()
	}
	_, err := s.entClient.PaymentOrder.Update().Where(
		paymentorder.IDEQ(orderID), paymentorder.StatusEQ(OrderStatusPending),
	).SetStatus(OrderStatusFailed).SetFailedAt(time.Now()).SetFailedReason(reason).Save(ctx)
	if err == nil {
		s.writeAuditLog(ctx, orderID, "USDT_QUOTE_FAILED", usdtpayment.ProviderKey, map[string]any{"reason": reason})
	}
	return err
}

func usdtBridgeEqualDecimal(left, right string) bool {
	a, err := decimal.NewFromString(strings.TrimSpace(left))
	if err != nil {
		return false
	}
	b, err := decimal.NewFromString(strings.TrimSpace(right))
	return err == nil && a.Equal(b)
}
