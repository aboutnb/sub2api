package service

import "context"

// IsModelSupportedForSmartRoute reuses the gateway's canonical account model
// eligibility rules (including Antigravity, Bedrock, OAuth normalization and
// passthrough behavior). Smart routing must use the same predicate as the
// scheduler so a candidate is not selected only to fail later in forwarding.
func IsModelSupportedForSmartRoute(ctx context.Context, account *Account, requestedModel string) bool {
	if account == nil {
		return false
	}
	return (&GatewayService{}).isModelSupportedByAccountWithContext(ctx, account, requestedModel)
}
