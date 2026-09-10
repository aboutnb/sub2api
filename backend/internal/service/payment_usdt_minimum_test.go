package service

import (
	"testing"

	infraerrors "github.com/Wei-Shaw/sub2api/internal/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestValidateUSDTMinimumPaymentAmount(t *testing.T) {
	t.Parallel()

	require.NoError(t, validateUSDTMinimumPaymentAmount("alipay", 1, 50))
	require.NoError(t, validateUSDTMinimumPaymentAmount("usdt_trc20", 49, 0))
	require.NoError(t, validateUSDTMinimumPaymentAmount("usdt_trc20", 50, 50))
	require.NoError(t, validateUSDTMinimumPaymentAmount("USDT-TRC20", 51, 50))

	err := validateUSDTMinimumPaymentAmount("usdt.tron", 49.99, 50)
	require.Error(t, err)
	require.Equal(t, "PAYMENT_AMOUNT_BELOW_USDT_MINIMUM", infraerrors.FromError(err).Reason)
}
