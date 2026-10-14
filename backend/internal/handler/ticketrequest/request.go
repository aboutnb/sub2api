// Package ticketrequest bounds the small JSON and pagination used by support tickets.
package ticketrequest

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"unicode/utf8"

	"github.com/Wei-Shaw/sub2api/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

const MaxBodyBytes = 64 << 10

func Bind(c *gin.Context, target any) bool {
	// Read before decoding so unknown fields and trailing data cannot bypass the cap.
	raw, err := io.ReadAll(http.MaxBytesReader(c.Writer, c.Request.Body, MaxBodyBytes))
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		response.ErrorWithDetails(c, http.StatusRequestEntityTooLarge, "Ticket request is too large", "TICKET_REQUEST_TOO_LARGE", nil)
		return false
	}
	if err != nil || !utf8.Valid(raw) {
		response.BadRequest(c, "Invalid request body")
		return false
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		response.BadRequest(c, "Invalid request body")
		return false
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		response.BadRequest(c, "Invalid request body")
		return false
	}
	return true
}

func Pagination(c *gin.Context) (page, size int, ok bool) {
	page, size = 1, 20
	for _, field := range []struct {
		value  string
		max    int
		target *int
	}{
		{c.Query("page"), 1000000, &page},
		{c.DefaultQuery("page_size", c.Query("limit")), 100, &size},
	} {
		if field.value == "" {
			continue
		}
		value, err := strconv.Atoi(field.value)
		if err != nil || value < 1 || value > field.max {
			response.ErrorWithDetails(c, http.StatusBadRequest, "Invalid ticket pagination", "TICKET_QUERY_INVALID", nil)
			return 0, 0, false
		}
		*field.target = value
	}
	return page, size, true
}
