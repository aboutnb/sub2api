package dto

import (
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/stretchr/testify/require"
)

func TestRedeemCodeCheckinModeIsExposedWithoutInternalNote(t *testing.T) {
	code := &service.RedeemCode{
		Type:  service.RedeemTypeCheckin,
		Notes: service.CheckinHistoryModeNote("lucky"),
	}

	userDTO := RedeemCodeFromService(code)
	require.NotNil(t, userDTO.CheckinMode)
	require.Equal(t, "lucky", *userDTO.CheckinMode)
	require.Nil(t, userDTO.Notes)

	adminDTO := RedeemCodeFromServiceAdmin(code)
	require.NotNil(t, adminDTO.CheckinMode)
	require.Equal(t, "lucky", *adminDTO.CheckinMode)
	require.Empty(t, adminDTO.Notes)
}
