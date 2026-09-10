package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestSmartRouteRequestDescriptorJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/v1/responses", io.NopCloser(stringReader(`{"model":"gpt-5","stream":true}`)))
	req.Header.Set("Content-Type", "application/json")
	descriptor, err := smartRouteRequestDescriptor(req)
	require.NoError(t, err)
	require.Equal(t, "gpt-5", descriptor.Model)
	require.True(t, descriptor.Streaming)
	require.Equal(t, "text", descriptor.Kind)
	body, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.JSONEq(t, `{"model":"gpt-5","stream":true}`, string(body))
}

func TestSmartRouteRequestDescriptorGeminiAndImage(t *testing.T) {
	gemini := httptest.NewRequest(http.MethodPost, "/v1beta/models/gemini-2.5-pro:generateContent", nil)
	descriptor, err := smartRouteRequestDescriptor(gemini)
	require.NoError(t, err)
	require.Equal(t, "gemini-2.5-pro", descriptor.Model)

	image := httptest.NewRequest(http.MethodPost, "/v1/images/edits", nil)
	descriptor, err = smartRouteRequestDescriptor(image)
	require.NoError(t, err)
	require.Equal(t, "image", descriptor.Kind)
}

func TestSmartRouteRequestDescriptorMultipartImageRestoresBody(t *testing.T) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	require.NoError(t, writer.WriteField("model", "gpt-image-1"))
	file, err := writer.CreateFormFile("image", "input.png")
	require.NoError(t, err)
	_, err = file.Write([]byte("image-bytes"))
	require.NoError(t, err)
	require.NoError(t, writer.Close())
	original := append([]byte(nil), body.Bytes()...)

	req := httptest.NewRequest(http.MethodPost, "/v1/images/edits", bytes.NewReader(original))
	req.Header.Set("Content-Type", writer.FormDataContentType())
	descriptor, err := smartRouteRequestDescriptor(req)
	require.NoError(t, err)
	require.Equal(t, "gpt-image-1", descriptor.Model)
	require.Equal(t, "image", descriptor.Kind)
	restored, err := io.ReadAll(req.Body)
	require.NoError(t, err)
	require.Equal(t, original, restored)
}

func TestSmartRouteRequestDescriptorDetectsClaudeCode(t *testing.T) {
	body := `{
		"model":"claude-sonnet-4-20250514",
		"system":[{"type":"text","text":"You are Claude Code, Anthropic's official CLI for Claude."}],
		"metadata":{"user_id":"user_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa_account__session_12345678-1234-1234-1234-123456789abc"}
	}`
	req := httptest.NewRequest(http.MethodPost, "/v1/messages", stringReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "claude-cli/1.0.1")
	req.Header.Set("X-App", "claude-code")
	req.Header.Set("anthropic-beta", "message-batches-2024-09-24")
	req.Header.Set("anthropic-version", "2023-06-01")

	descriptor, err := smartRouteRequestDescriptor(req)
	require.NoError(t, err)
	require.True(t, descriptor.ClaudeCode)
}

func TestAbortSmartRouteModelUnsupportedUsesEndpointProtocol(t *testing.T) {
	gin.SetMode(gin.TestMode)
	tests := []struct {
		path           string
		wantTopLevel   string
		wantNestedCode string
	}{
		{path: "/v1/messages", wantTopLevel: "error"},
		{path: "/v1/chat/completions", wantNestedCode: "model_not_found"},
	}
	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			writer := httptest.NewRecorder()
			ctx, _ := gin.CreateTestContext(writer)
			ctx.Request = httptest.NewRequest(http.MethodPost, test.path, nil)
			abortSmartRouteResolveError(ctx, service.ErrSmartRouteModelUnsupported)
			require.Equal(t, http.StatusNotFound, writer.Code)
			var body map[string]any
			require.NoError(t, json.Unmarshal(writer.Body.Bytes(), &body))
			if test.wantTopLevel != "" {
				require.Equal(t, test.wantTopLevel, body["type"])
			}
			if test.wantNestedCode != "" {
				errorBody, ok := body["error"].(map[string]any)
				require.True(t, ok)
				require.Equal(t, test.wantNestedCode, errorBody["code"])
			}
		})
	}
}

type staticStringReader string

func (r *staticStringReader) Read(p []byte) (int, error) {
	if len(*r) == 0 {
		return 0, io.EOF
	}
	n := copy(p, *r)
	*r = (*r)[n:]
	return n, nil
}

func stringReader(value string) io.Reader {
	reader := staticStringReader(value)
	return &reader
}
