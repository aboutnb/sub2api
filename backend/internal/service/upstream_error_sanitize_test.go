package service

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func newSanitizeTestContext() *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest("POST", "https://proxy.flowai.cyou/v1/messages", nil)
	c.Request.Host = "proxy.flowai.cyou"
	return c
}

func TestSanitizeUpstreamErrorMessageForClientPreservesCurrentHost(t *testing.T) {
	c := newSanitizeTestContext()
	message := "request failed at https://upstream.example/v1?access_token=secret; retry via https://proxy.flowai.cyou/v1, status 429; error.type=api_error; use response.id"

	got := SanitizeUpstreamErrorMessageForClient(c, message)

	require.Contains(t, got, "https://proxy.flowai.cyou/v1")
	require.Contains(t, got, "status 429")
	require.Contains(t, got, "error.type=api_error")
	require.Contains(t, got, "response.id")
	require.NotContains(t, got, "upstream.example")
	require.NotContains(t, got, "access_token=secret")
	require.Contains(t, got, upstreamAddressRedaction)
}

func TestSanitizeUpstreamErrorMessageForClientFiltersNetworkAddresses(t *testing.T) {
	c := newSanitizeTestContext()
	message := "connect 203.0.113.8:8443 or [2001:db8::8]:443; provider api.openai.com; error.type=server_error"

	got := SanitizeUpstreamErrorMessageForClient(c, message)

	require.NotContains(t, got, "203.0.113.8")
	require.NotContains(t, got, "2001:db8::8")
	require.NotContains(t, got, "api.openai.com")
	require.Contains(t, got, "error.type=server_error")
}

func TestSanitizeUpstreamErrorBodyForClientPreservesJSONShapeAndNumbers(t *testing.T) {
	c := newSanitizeTestContext()
	body := []byte(`{"error":{"type":"api_error","message":"balance unavailable at https://upstream.example/v1?key=secret"},"status":429,"retryable":false,"items":["api.openai.com",3]}`)

	got := SanitizeUpstreamErrorBodyForClient(c, body)
	var decoded map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(got)))
	decoder.UseNumber()
	require.NoError(t, decoder.Decode(&decoded))

	require.Equal(t, json.Number("429"), decoded["status"])
	require.Equal(t, false, decoded["retryable"])
	items, ok := decoded["items"].([]any)
	require.True(t, ok)
	require.Len(t, items, 2)
	require.Equal(t, json.Number("3"), items[1])
	errorBody, ok := decoded["error"].(map[string]any)
	require.True(t, ok)
	message, ok := errorBody["message"].(string)
	require.True(t, ok)
	require.NotContains(t, message, "upstream.example")
	require.NotContains(t, string(got), "api.openai.com")
}

func TestSanitizeUpstreamErrorSSELineForClientOnlyChangesErrorData(t *testing.T) {
	c := newSanitizeTestContext()
	normal := `data: {"type":"response.output_text.delta","delta":"https://upstream.example should remain model text"}`
	errorLine := `data: {"type":"error","error":{"message":"upstream https://upstream.example/v1"}}`

	require.Equal(t, normal, SanitizeUpstreamErrorSSELineForClient(c, normal))
	got := SanitizeUpstreamErrorSSELineForClient(c, errorLine)
	require.NotContains(t, got, "upstream.example")
	require.Contains(t, got, upstreamAddressRedaction)
}
