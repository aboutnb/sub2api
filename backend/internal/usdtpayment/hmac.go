package usdtpayment

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

const (
	headerKeyID     = "X-BEPUSDT-Key-Id"
	headerTimestamp = "X-BEPUSDT-Timestamp"
	headerNonce     = "X-BEPUSDT-Nonce"
	headerDigest    = "X-BEPUSDT-Content-SHA256"
	headerSignature = "X-BEPUSDT-Signature"
)

func sha256Hex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func hmacV2Sign(secret, method, path, timestamp, nonce, digest string) string {
	canonical := strings.Join([]string{strings.ToUpper(method), path, timestamp, nonce, strings.ToLower(digest)}, "\n")
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(canonical))
	return hex.EncodeToString(mac.Sum(nil))
}

func hmacHex(secret string, body []byte) string {
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write(body)
	return hex.EncodeToString(mac.Sum(nil))
}

func secureNonce() (string, error) {
	buf := make([]byte, 24)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate nonce: %w", err)
	}
	return hex.EncodeToString(buf), nil
}

func equalHex(actual, expected string) bool {
	a, err := hex.DecodeString(strings.TrimSpace(actual))
	if err != nil {
		return false
	}
	e, err := hex.DecodeString(strings.TrimSpace(expected))
	return err == nil && hmac.Equal(a, e)
}
