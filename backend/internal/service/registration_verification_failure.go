package service

import "errors"

// VerificationFailure marks a confirmed invalid, unexpired proof. It preserves
// the public error and errors.Is behaviour while distinguishing provider errors
// and ordinary expiry, which must never count towards source bans.
type VerificationFailure struct{ Cause error }

func (e *VerificationFailure) Error() string { return e.Cause.Error() }
func (e *VerificationFailure) Unwrap() error { return e.Cause }
func IsRegistrationVerificationFailure(err error) bool {
	var failure *VerificationFailure
	return errors.As(err, &failure)
}
