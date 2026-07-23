package responseheaders

import (
	"net/http"
	"testing"

	"github.com/Wei-Shaw/sub2api/internal/config"
)

func TestFilterHeadersDisabledUsesDefaultAllowlist(t *testing.T) {
	src := http.Header{}
	src.Add("Content-Type", "application/json")
	src.Add("X-Request-Id", "req-123")
	src.Add("X-Test", "ok")
	src.Add("Connection", "keep-alive")
	src.Add("Content-Length", "123")

	cfg := config.ResponseHeaderConfig{
		Enabled:     false,
		ForceRemove: []string{"x-request-id"},
	}

	filtered := FilterHeaders(src, CompileHeaderFilter(cfg))
	if filtered.Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type passthrough, got %q", filtered.Get("Content-Type"))
	}
	if filtered.Get("X-Request-Id") != "req-123" {
		t.Fatalf("expected X-Request-Id allowed, got %q", filtered.Get("X-Request-Id"))
	}
	if filtered.Get("X-Test") != "" {
		t.Fatalf("expected X-Test removed, got %q", filtered.Get("X-Test"))
	}
	if filtered.Get("Connection") != "" {
		t.Fatalf("expected Connection to be removed, got %q", filtered.Get("Connection"))
	}
	if filtered.Get("Content-Length") != "" {
		t.Fatalf("expected Content-Length to be removed, got %q", filtered.Get("Content-Length"))
	}
}

func TestFilterHeadersEnabledUsesAllowlist(t *testing.T) {
	src := http.Header{}
	src.Add("Content-Type", "application/json")
	src.Add("X-Extra", "ok")
	src.Add("X-Remove", "nope")
	src.Add("X-Blocked", "nope")

	cfg := config.ResponseHeaderConfig{
		Enabled:           true,
		AdditionalAllowed: []string{"x-extra"},
		ForceRemove:       []string{"x-remove"},
	}

	filtered := FilterHeaders(src, CompileHeaderFilter(cfg))
	if filtered.Get("Content-Type") != "application/json" {
		t.Fatalf("expected Content-Type allowed, got %q", filtered.Get("Content-Type"))
	}
	if filtered.Get("X-Extra") != "ok" {
		t.Fatalf("expected X-Extra allowed, got %q", filtered.Get("X-Extra"))
	}
	if filtered.Get("X-Remove") != "" {
		t.Fatalf("expected X-Remove removed, got %q", filtered.Get("X-Remove"))
	}
	if filtered.Get("X-Blocked") != "" {
		t.Fatalf("expected X-Blocked removed, got %q", filtered.Get("X-Blocked"))
	}
}

func TestWriteFilteredErrorHeadersRemovesAddressBearingHeaders(t *testing.T) {
	src := http.Header{
		"Content-Type":     []string{"application/json"},
		"Location":         []string{"https://upstream.example/reauth"},
		"WWW-Authenticate": []string{"Bearer realm=https://upstream.example"},
		"Retry-After":      []string{"30"},
	}
	dst := http.Header{}

	WriteFilteredErrorHeaders(dst, src, CompileHeaderFilter(config.ResponseHeaderConfig{}))

	if dst.Get("Content-Type") != "application/json" {
		t.Fatalf("expected content type to remain, got %q", dst.Get("Content-Type"))
	}
	if dst.Get("Location") != "" || dst.Get("WWW-Authenticate") != "" {
		t.Fatalf("address-bearing error headers must be removed: %#v", dst)
	}
	if dst.Get("Retry-After") != "30" {
		t.Fatalf("expected Retry-After to remain, got %q", dst.Get("Retry-After"))
	}
}
