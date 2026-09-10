package service

import (
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestRedeemCodeExpiry(t *testing.T) {
	now := time.Now().UTC()
	past := now.Add(-time.Hour)
	future := now.Add(time.Hour)

	tests := []struct {
		name        string
		code        RedeemCode
		wantExpired bool
		wantCanUse  bool
	}{
		{
			name:        "unused without expiry can be used",
			code:        RedeemCode{Status: StatusUnused},
			wantExpired: false,
			wantCanUse:  true,
		},
		{
			name:        "unused before expiry can be used",
			code:        RedeemCode{Status: StatusUnused, ExpiresAt: &future},
			wantExpired: false,
			wantCanUse:  true,
		},
		{
			name:        "unused after expiry cannot be used",
			code:        RedeemCode{Status: StatusUnused, ExpiresAt: &past},
			wantExpired: true,
			wantCanUse:  false,
		},
		{
			name:        "explicit expired status is expired",
			code:        RedeemCode{Status: StatusExpired},
			wantExpired: true,
			wantCanUse:  false,
		},
		{
			name:        "used code remains used even after expiry time",
			code:        RedeemCode{Status: StatusUsed, ExpiresAt: &past},
			wantExpired: false,
			wantCanUse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.wantExpired, tt.code.IsExpiredAt(now))
			require.Equal(t, tt.wantCanUse, tt.code.CanUse())
		})
	}
}

func TestGenerateRedeemCodeFormat(t *testing.T) {
	code, err := GenerateRedeemCode()
	require.NoError(t, err)

	parts := strings.Split(code, "-")
	require.Len(t, parts, redeemCodeLength/redeemCodeGroupSize)

	joined := strings.Join(parts, "")
	require.Len(t, joined, redeemCodeLength)

	for _, part := range parts {
		require.Len(t, part, redeemCodeGroupSize)
		require.Equal(t, strings.ToUpper(part), part)
		for _, ch := range part {
			require.Containsf(t, redeemCodeAlphabet, string(ch), "unexpected character %q in %q", ch, code)
		}
	}
}
