package middleware

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCheckinRoutesUseExplicitAuditActions(t *testing.T) {
	require.Equal(t, "user.checkin.claim", auditActionOverrides["POST /api/v1/user/checkin"])
	require.Equal(t, "admin.checkin.config.update", auditActionOverrides["PUT /api/v1/admin/checkin/config"])
}
