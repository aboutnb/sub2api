package ticketrequest

import (
	"encoding/json"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func TestTicketRequestBoundaries(t *testing.T) {
	for _, tc := range []struct {
		name, body string
		code       int
	}{
		{"normal", `{"body":"hello"}`, 200},
		{"escaped unicode boundary", `{"body":"` + strings.Repeat(`\ud83d\ude80`, 4000) + `"}`, 200},
		{"too large unknown field", `{"body":"ok","padding":"` + strings.Repeat("x", MaxBodyBytes) + `"}`, 413},
		{"spoofed author", `{"body":"ok","author_role":"admin"}`, 400},
		{"trailing object", `{"body":"ok"}{}`, 400},
		{"malformed", `{"body":`, 400},
		{"invalid utf8", "{\"body\":\"\xff\"}", 400},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest("POST", "/tickets", strings.NewReader(tc.body))
			var target struct {
				Body string `json:"body"`
			}
			if Bind(c, &target) {
				c.Status(200)
			}
			require.Equal(t, tc.code, rec.Code)
			if tc.code >= 400 {
				var payload map[string]any
				require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &payload))
				require.NotContains(t, payload["message"], "padding")
			}
		})
	}
}

func TestTicketPaginationRejectsOverflow(t *testing.T) {
	for _, query := range []string{"page=9223372036854775807", "page=999999999999999999999999999", "page=-1", "page=abc", "page_size=101", "limit=9999999"} {
		t.Run(query, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = httptest.NewRequest("GET", "/tickets?"+query, nil)
			_, _, ok := Pagination(c)
			require.False(t, ok)
			require.Equal(t, 400, rec.Code)
		})
	}
}
