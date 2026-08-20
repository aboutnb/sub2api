package usdtpayment

import "github.com/google/wire"

var ProviderSet = wire.NewSet(
	NewClient,
	NewRepository,
	NewService,
)
