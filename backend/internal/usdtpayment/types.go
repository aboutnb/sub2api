package usdtpayment

import (
	"context"
	"time"
)

const (
	PaymentType = "usdt"
	ProviderKey = "bepusdt"
)

var networkTradeTypes = map[string]string{
	"tron":     "usdt.trc20",
	"bsc":      "usdt.bep20",
	"ethereum": "usdt.erc20",
	"polygon":  "usdt.polygon",
	"arbitrum": "usdt.arbitrum",
	"solana":   "usdt.solana",
	"ton":      "usdt.ton",
	"aptos":    "usdt.aptos",
	"xlayer":   "usdt.xlayer",
	"plasma":   "usdt.plasma",
}

type PrepareOrderRequest struct {
	UserID int64
	Amount float64
	// TargetPayAmount is the exact CNY amount to send to BEpusdt when the
	// customer entered a crypto-denominated amount.
	TargetPayAmount float64
	OrderType       string
	PlanID          int64
	ClientIP        string
	SourceHost      string
	SourceURL       string
	PaymentSource   string
	Locale          string
}

type PreparedOrder struct {
	ID              int64
	MerchantOrderID string
	FiatAmount      string
	BaseAmount      float64
	PayAmount       float64
	FeeRate         float64
	OrderType       string
	ExpiresAt       time.Time
}

type ConfirmOrderRequest struct {
	PaymentOrderID   int64
	MerchantOrderID  string
	ProviderTradeID  string
	TransactionHash  string
	FiatAmount       string
	TransferAt       time.Time
	QuoteExpiresAt   time.Time
	RecoveryDeadline time.Time
}

// PaymentBridge is the only dependency on Sub2API's existing payment domain.
// BEpusdt never enters the RMB provider registry.
type PaymentBridge interface {
	PrepareUSDTOrder(context.Context, PrepareOrderRequest) (*PreparedOrder, error)
	ConfirmUSDTOrder(context.Context, ConfirmOrderRequest) error
	FailUSDTOrderBeforeQuote(context.Context, int64, error) error
}

type Capability struct {
	Crypto          string `json:"crypto"`
	Network         string `json:"network"`
	NetworkName     string `json:"network_name"`
	TradeType       string `json:"trade_type"`
	WalletCount     int    `json:"wallet_count"`
	RPCEndpointSet  bool   `json:"rpc_endpoint_set"`
	ScannerBlock    string `json:"scanner_block"`
	ScannerSuccess  string `json:"scanner_success"`
	LastScanAt      int64  `json:"last_scan_at"`
	AcceptingOrders bool   `json:"accepting_orders"`
	Reason          string `json:"reason,omitempty"`
}

// RateQuote is the provider's current fiat value for one unit of USDT.
// The rate is fetched immediately before order creation and the returned
// order must keep the same snapshot to prevent crypto/fiat drift.
type RateQuote struct {
	Crypto    string `json:"crypto"`
	Fiat      string `json:"fiat"`
	Rate      string `json:"rate"`
	UpdatedAt int64  `json:"updated_at"`
}

type UpstreamOrder struct {
	OrderID            string         `json:"order_id"`
	TradeID            string         `json:"trade_id"`
	Status             int            `json:"status"`
	StatusName         string         `json:"status_name"`
	Fiat               string         `json:"fiat"`
	Amount             string         `json:"amount"`
	Crypto             string         `json:"crypto"`
	ActualAmount       string         `json:"actual_amount"`
	ExchangeRate       string         `json:"exchange_rate"`
	TradeType          string         `json:"trade_type"`
	Network            string         `json:"network"`
	Token              string         `json:"token"`
	BlockTransactionID string         `json:"block_transaction_id"`
	TransferAt         int64          `json:"transfer_at"`
	CreatedAt          int64          `json:"created_at"`
	ExpiresAt          int64          `json:"expires_at"`
	PaymentURL         string         `json:"payment_url,omitempty"`
	Confirmation       map[string]any `json:"confirmation"`
}

type Quote struct {
	ID                  int64      `json:"-"`
	PaymentOrderID      int64      `json:"payment_order_id"`
	MerchantOrderID     string     `json:"out_trade_no"`
	ProviderTradeID     string     `json:"provider_trade_id"`
	FiatCurrency        string     `json:"fiat_currency"`
	FiatAmount          string     `json:"fiat_amount"`
	CryptoCurrency      string     `json:"crypto_currency"`
	Network             string     `json:"network"`
	TradeType           string     `json:"trade_type"`
	CryptoAmount        string     `json:"crypto_amount"`
	ExchangeRate        string     `json:"exchange_rate"`
	ReceivingAddress    string     `json:"receiving_address"`
	PaymentURL          string     `json:"payment_url"`
	UpstreamCreatedAt   time.Time  `json:"created_at"`
	UpstreamExpiresAt   time.Time  `json:"expires_at"`
	ProviderStatus      string     `json:"provider_status"`
	TransactionHash     *string    `json:"transaction_hash,omitempty"`
	ChainTransferAt     *time.Time `json:"chain_transfer_at,omitempty"`
	BlockNumber         *int64     `json:"block_number,omitempty"`
	LastReconciledAt    *time.Time `json:"last_reconciled_at,omitempty"`
	NextReconcileAt     time.Time  `json:"-"`
	ReconcileAttempts   int        `json:"-"`
	ReconcileLeaseUntil *time.Time `json:"-"`
	LastError           *string    `json:"-"`
	CreatedAt           time.Time  `json:"-"`
	UpdatedAt           time.Time  `json:"-"`
}

type WebhookPayload struct {
	EventID            string `json:"event_id"`
	EventType          string `json:"event_type"`
	OccurredAt         int64  `json:"occurred_at"`
	OrderID            string `json:"order_id"`
	TradeID            string `json:"trade_id"`
	Fiat               string `json:"fiat"`
	Amount             string `json:"amount"`
	Crypto             string `json:"crypto"`
	ActualAmount       string `json:"actual_amount"`
	TradeType          string `json:"trade_type"`
	Network            string `json:"network"`
	Token              string `json:"token"`
	BlockTransactionID string `json:"block_transaction_id"`
	TransferAt         int64  `json:"transfer_at"`
	BlockNumber        int64  `json:"block_number"`
	SignatureVersion   string `json:"signature_version"`
	KeyID              string `json:"key_id"`
	Signature          string `json:"signature,omitempty"`
}

type CreateRequest struct {
	Amount        float64 `json:"amount"` // USDT amount entered by the user
	AmountUnit    string  `json:"amount_unit"`
	Network       string  `json:"network" binding:"required"`
	OrderType     string  `json:"order_type"`
	PlanID        int64   `json:"plan_id"`
	ReturnURL     string  `json:"return_url"`
	PaymentSource string  `json:"payment_source"`
}

type CheckoutOrder struct {
	OrderID          int64      `json:"order_id"`
	OutTradeNo       string     `json:"out_trade_no"`
	Amount           float64    `json:"amount"`
	PayAmount        float64    `json:"pay_amount"`
	FeeRate          float64    `json:"fee_rate"`
	Status           string     `json:"status"`
	PaymentType      string     `json:"payment_type"`
	FiatCurrency     string     `json:"fiat_currency"`
	FiatAmount       string     `json:"fiat_amount"`
	CryptoCurrency   string     `json:"crypto_currency"`
	CryptoAmount     string     `json:"crypto_amount"`
	Network          string     `json:"network"`
	TradeType        string     `json:"trade_type"`
	ReceivingAddress string     `json:"receiving_address"`
	ExchangeRate     string     `json:"exchange_rate"`
	PaymentURL       string     `json:"payment_url"`
	ExpiresAt        time.Time  `json:"expires_at"`
	TransactionHash  *string    `json:"transaction_hash,omitempty"`
	ChainTransferAt  *time.Time `json:"chain_transfer_at,omitempty"`
}
