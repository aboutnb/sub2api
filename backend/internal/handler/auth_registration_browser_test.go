package handler

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestRegistrationBrowserSignedCookie(t *testing.T) {
	h := newRegistrationChallengeTestHandler()
	c, recorder := newRegistrationChallengeTestContext()
	c.Request.Header.Set("X-Forwarded-Proto", "https")
	h.ensureRegistrationBrowser(c)
	cookies := recorder.Result().Cookies()
	require.Len(t, cookies, 1)
	cookie := cookies[0]
	require.True(t, cookie.HttpOnly)
	require.True(t, cookie.Secure)
	require.Equal(t, http.SameSiteLaxMode, cookie.SameSite)
	c.Request.AddCookie(cookie)
	hash := h.registrationBrowserHash(c)
	require.Len(t, hash, 64)
	// IP changes retain the browser identity, while a forged cookie is ignored.
	c.Request.RemoteAddr = "198.51.100.7:1234"
	require.Equal(t, hash, h.registrationBrowserHash(c))
	c.Request.Header.Set("Cookie", registrationBrowserCookie+"=x"+cookie.Value[1:])
	require.Empty(t, h.registrationBrowserHash(c))
}

func TestRegistrationFeatureBoundToChallenge(t *testing.T) {
	h := newRegistrationChallengeTestHandler()
	c, _ := newRegistrationChallengeTestContext()
	c.Request.Header.Set(registrationFeatureHeader, strings.Repeat("a", 64))
	submission := buildRegistrationChallengeSubmissionForTest(t, h, c, "user@example.com", "register", "")
	_, err := h.validateRegistrationChallenge(c, "register", "user@example.com", submission)
	require.NoError(t, err)
	c.Request.Header.Set(registrationFeatureHeader, strings.Repeat("b", 64))
	_, err = h.validateRegistrationChallenge(c, "register", "user@example.com", submission)
	require.ErrorIs(t, err, errRegistrationChallengeInvalid)
}

func TestRegistrationBrowserMissingOrInvalidSignalsDoNotRejectLegacyChallenge(t *testing.T) {
	h := newRegistrationChallengeTestHandler()
	c, _ := newRegistrationChallengeTestContext()
	c.Request.Header.Set("Cookie", registrationBrowserCookie+"=forged")
	c.Request.Header.Set(registrationFeatureHeader, strings.Repeat("z", 64))
	require.Empty(t, h.registrationBrowserHash(c))
	require.Empty(t, h.registrationFeatureHash(c))
	submission := buildRegistrationChallengeSubmissionForTest(t, h, c, "user@example.com", "register", "")
	_, err := h.validateRegistrationChallenge(c, "register", "user@example.com", submission)
	require.NoError(t, err)
}
