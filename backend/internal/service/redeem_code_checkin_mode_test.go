package service

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRedeemCodeCheckinMode(t *testing.T) {
	tests := []struct {
		name string
		code RedeemCode
		want string
	}{
		{name: "lucky", code: RedeemCode{Type: RedeemTypeCheckin, Notes: CheckinHistoryModeNote("lucky")}, want: "lucky"},
		{name: "normal", code: RedeemCode{Type: RedeemTypeCheckin, Notes: CheckinHistoryModeNote("normal")}, want: "normal"},
		{name: "other redeem type", code: RedeemCode{Type: RedeemTypeBalance, Notes: CheckinHistoryModeNote("lucky")}},
		{name: "invalid mode", code: RedeemCode{Type: RedeemTypeCheckin, Notes: CheckinHistoryModeNote("invalid")}},
		{name: "legacy empty note", code: RedeemCode{Type: RedeemTypeCheckin}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			require.Equal(t, test.want, test.code.CheckinMode())
		})
	}
}
