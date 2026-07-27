package admin

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUpdateCheckinConfigRequestAcceptsDecimalStringsAndNumbers(t *testing.T) {
	var request updateCheckinConfigRequest
	err := json.Unmarshal([]byte(`{
		"normal_enabled":true,
		"lucky_enabled":false,
		"normal_min":"0.01",
		"normal_max":0.05,
		"lucky_reward_type":"multiplier",
		"lucky_positive_probability":70,
		"lucky_min_multiplier":0.5,
		"lucky_max_multiplier":2,
		"lucky_amount_min":"-0.05",
		"lucky_amount_max":0.10,
		"change_reason":"test"
	}`), &request)

	require.NoError(t, err)
	require.NotNil(t, request.NormalEnabled)
	require.True(t, *request.NormalEnabled)
	require.NotNil(t, request.LuckyEnabled)
	require.False(t, *request.LuckyEnabled)
	require.Equal(t, "0.01", string(request.NormalMin))
	require.Equal(t, "0.05", string(request.NormalMax))
	require.Equal(t, "0.5", string(request.LuckyMinMultiply))
	require.Equal(t, "70", string(request.LuckyPositiveProbability))
	require.Equal(t, "2", string(request.LuckyMaxMultiply))
	require.Equal(t, "-0.05", string(request.LuckyAmountMin))
	require.Equal(t, "0.10", string(request.LuckyAmountMax))
}

func TestUpdateCheckinConfigRequestRejectsNonDecimalValues(t *testing.T) {
	var request updateCheckinConfigRequest
	err := json.Unmarshal([]byte(`{"normal_min":true}`), &request)

	require.Error(t, err)
}
