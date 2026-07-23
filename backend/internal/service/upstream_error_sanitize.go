package service

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/url"
	"regexp"
	"strings"

	"github.com/gin-gonic/gin"
	"golang.org/x/net/publicsuffix"
)

const upstreamAddressRedaction = "[upstream address redacted]"

var (
	sensitiveQueryParamRegex = regexp.MustCompile(`(?i)([?&](?:key|client_secret|access_token|refresh_token)=)[^&"\s]+`)
	absoluteURLRegex         = regexp.MustCompile(`(?i)(?:https?|wss?)://[^\s<>"']+`)
	protocolRelativeRegex    = regexp.MustCompile(`(?i)//(?:\[[0-9a-f:.]+\]|[a-z0-9][a-z0-9.-]*)(?::[0-9]{1,5})?(?:/[^\s<>"']*)?`)
	ipv4AddressRegex         = regexp.MustCompile(`\b(?:(?:25[0-5]|2[0-4][0-9]|1?[0-9]{1,2})\.){3}(?:25[0-5]|2[0-4][0-9]|1?[0-9]{1,2})(?::[0-9]{1,5})?\b`)
	bracketedIPv6Regex       = regexp.MustCompile(`\[[0-9a-fA-F:.]+\](?::[0-9]{1,5})?`)
	domainAddressRegex       = regexp.MustCompile(`(?i)\b(?:[a-z0-9](?:[a-z0-9-]{0,62}[a-z0-9])?\.)+[a-z]{2,63}(?::[0-9]{1,5})?\b`)
	dottedFieldSuffixes      = map[string]struct{}{
		"code": {}, "content": {}, "data": {}, "delta": {}, "details": {},
		"email": {}, "id": {}, "index": {}, "message": {}, "model": {},
		"name": {}, "object": {}, "reason": {}, "role": {}, "status": {},
		"type": {}, "usage": {},
	}
)

// sanitizeUpstreamErrorMessage removes query credentials for internal diagnostics.
// Client responses must additionally use SanitizeUpstreamErrorMessageForClient.
func sanitizeUpstreamErrorMessage(msg string) string {
	if msg == "" {
		return msg
	}
	return sensitiveQueryParamRegex.ReplaceAllString(msg, `$1***`)
}

// SanitizeUpstreamErrorMessageForClient preserves the current request host while
// removing external network addresses from an upstream-provided error message.
func SanitizeUpstreamErrorMessageForClient(c *gin.Context, msg string) string {
	if msg == "" {
		return msg
	}

	currentHosts := currentRequestHosts(c)
	msg = sanitizeUpstreamErrorMessage(msg)
	msg = absoluteURLRegex.ReplaceAllStringFunc(msg, func(candidate string) string {
		return redactURLCandidate(candidate, currentHosts)
	})
	msg = protocolRelativeRegex.ReplaceAllStringFunc(msg, func(candidate string) string {
		return redactURLCandidate(candidate, currentHosts)
	})
	msg = bracketedIPv6Regex.ReplaceAllStringFunc(msg, func(candidate string) string {
		if currentHosts.contains(addressHostname(candidate)) {
			return candidate
		}
		return upstreamAddressRedaction
	})
	msg = ipv4AddressRegex.ReplaceAllStringFunc(msg, func(candidate string) string {
		if currentHosts.contains(addressHostname(candidate)) {
			return candidate
		}
		return upstreamAddressRedaction
	})
	msg = domainAddressRegex.ReplaceAllStringFunc(msg, func(candidate string) string {
		if !isPublicDomainAddress(candidate) || currentHosts.contains(addressHostname(candidate)) {
			return candidate
		}
		return upstreamAddressRedaction
	})
	return msg
}

// SanitizeUpstreamErrorBodyForClient recursively sanitizes string values in a
// JSON error body. Non-JSON errors are sanitized as plain text. The original
// bytes are returned when no client-visible value changed.
func SanitizeUpstreamErrorBodyForClient(c *gin.Context, body []byte) []byte {
	if len(body) == 0 {
		return body
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.UseNumber()
	var decoded any
	if err := decoder.Decode(&decoded); err != nil {
		return []byte(SanitizeUpstreamErrorMessageForClient(c, string(body)))
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		return []byte(SanitizeUpstreamErrorMessageForClient(c, string(body)))
	}

	sanitized, changed := sanitizeUpstreamJSONValue(c, decoded)
	if !changed {
		return body
	}
	encoded, err := json.Marshal(sanitized)
	if err != nil {
		return []byte(SanitizeUpstreamErrorMessageForClient(c, string(body)))
	}
	return encoded
}

// SanitizeUpstreamErrorSSELineForClient sanitizes only SSE data lines that
// carry a structured error event. Normal model output remains byte-for-byte
// unchanged.
func SanitizeUpstreamErrorSSELineForClient(c *gin.Context, line string) string {
	lineEnd := ""
	content := line
	if strings.HasSuffix(content, "\r\n") {
		lineEnd = "\r\n"
		content = strings.TrimSuffix(content, "\r\n")
	} else if strings.HasSuffix(content, "\n") || strings.HasSuffix(content, "\r") {
		lineEnd = content[len(content)-1:]
		content = content[:len(content)-1]
	}
	if !strings.HasPrefix(content, "data:") {
		return line
	}
	payload := strings.TrimSpace(strings.TrimPrefix(content, "data:"))
	if payload == "" || payload == "[DONE]" {
		return line
	}
	sanitized := sanitizeUpstreamErrorSSEDataForClient(c, []byte(payload))
	if bytes.Equal(sanitized, []byte(payload)) {
		return line
	}
	return "data: " + string(sanitized) + lineEnd
}

func sanitizeUpstreamErrorSSEDataForClient(c *gin.Context, payload []byte) []byte {
	decoder := json.NewDecoder(bytes.NewReader(payload))
	var decoded map[string]any
	if err := decoder.Decode(&decoded); err != nil || decoded == nil {
		return payload
	}
	if !isUpstreamErrorJSONValue(decoded) {
		return payload
	}
	return SanitizeUpstreamErrorBodyForClient(c, payload)
}

func isUpstreamErrorJSONValue(value map[string]any) bool {
	if eventType, ok := value["type"].(string); ok {
		eventType = strings.TrimSpace(strings.ToLower(eventType))
		if eventType == "error" || eventType == "response.failed" {
			return true
		}
	}
	_, hasError := value["error"]
	return hasError
}

func sanitizeUpstreamJSONValue(c *gin.Context, value any) (any, bool) {
	switch typed := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		changed := false
		for key, child := range typed {
			sanitizedChild, childChanged := sanitizeUpstreamJSONValue(c, child)
			out[key] = sanitizedChild
			changed = changed || childChanged
		}
		return out, changed
	case []any:
		out := make([]any, len(typed))
		changed := false
		for i, child := range typed {
			sanitizedChild, childChanged := sanitizeUpstreamJSONValue(c, child)
			out[i] = sanitizedChild
			changed = changed || childChanged
		}
		return out, changed
	case string:
		sanitized := SanitizeUpstreamErrorMessageForClient(c, typed)
		return sanitized, sanitized != typed
	default:
		return value, false
	}
}

type requestHostSet map[string]struct{}

func currentRequestHosts(c *gin.Context) requestHostSet {
	hosts := requestHostSet{}
	if c == nil || c.Request == nil {
		return hosts
	}
	hosts.add(c.Request.Host)
	if c.Request.URL != nil {
		hosts.add(c.Request.URL.Host)
	}
	return hosts
}

func (hosts requestHostSet) add(raw string) {
	host := normalizeAddressHostname(raw)
	if host != "" {
		hosts[host] = struct{}{}
	}
}

func (hosts requestHostSet) contains(host string) bool {
	host = normalizeAddressHostname(host)
	if host == "" {
		return false
	}
	_, ok := hosts[host]
	return ok
}

func redactURLCandidate(candidate string, currentHosts requestHostSet) string {
	core, suffix := trimAddressPunctuation(candidate)
	parseTarget := core
	if strings.HasPrefix(parseTarget, "//") {
		parseTarget = "https:" + parseTarget
	}
	parsed, err := url.Parse(parseTarget)
	if err != nil || parsed.Hostname() == "" {
		return candidate
	}
	if currentHosts.contains(parsed.Hostname()) {
		return candidate
	}
	return upstreamAddressRedaction + suffix
}

func trimAddressPunctuation(candidate string) (string, string) {
	core := candidate
	for len(core) > 0 {
		last := core[len(core)-1]
		if !strings.ContainsRune(".,;!?)]}", rune(last)) {
			break
		}
		core = core[:len(core)-1]
	}
	return core, candidate[len(core):]
}

func addressHostname(candidate string) string {
	candidate, _ = trimAddressPunctuation(strings.TrimSpace(candidate))
	if strings.Contains(candidate, "://") || strings.HasPrefix(candidate, "//") {
		parseTarget := candidate
		if strings.HasPrefix(parseTarget, "//") {
			parseTarget = "https:" + parseTarget
		}
		if parsed, err := url.Parse(parseTarget); err == nil {
			return parsed.Hostname()
		}
	}
	return normalizeAddressHostname(candidate)
}

func isPublicDomainAddress(candidate string) bool {
	host := addressHostname(candidate)
	if host == "" || net.ParseIP(host) != nil {
		return false
	}
	labels := strings.Split(strings.TrimSuffix(host, "."), ".")
	if len(labels) == 2 {
		if _, isField := dottedFieldSuffixes[labels[1]]; isField {
			return false
		}
	}
	_, icann := publicsuffix.PublicSuffix(host)
	return icann
}

func normalizeAddressHostname(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	if host, _, err := net.SplitHostPort(raw); err == nil {
		return strings.ToLower(strings.Trim(host, "[]"))
	}
	if strings.HasPrefix(raw, "[") && strings.HasSuffix(raw, "]") {
		return strings.ToLower(strings.Trim(raw, "[]"))
	}
	if strings.Count(raw, ":") == 1 {
		if host, port, found := strings.Cut(raw, ":"); found && port != "" && allASCIIDigits(port) {
			return strings.ToLower(host)
		}
	}
	return strings.ToLower(strings.Trim(raw, "[]"))
}

func allASCIIDigits(value string) bool {
	if value == "" {
		return false
	}
	for i := 0; i < len(value); i++ {
		if value[i] < '0' || value[i] > '9' {
			return false
		}
	}
	return true
}
