package middleware

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestIsAsyncImageTaskRead(t *testing.T) {
	require.True(t, isAsyncImageTaskRead(http.MethodGet, "/v1/images/tasks/imgtask_123"))
	require.True(t, isAsyncImageTaskRead(http.MethodGet, "/images/tasks/imgtask_123"))
	require.False(t, isAsyncImageTaskRead(http.MethodPost, "/v1/images/tasks/imgtask_123"))
	require.False(t, isAsyncImageTaskRead(http.MethodGet, "/v1/images/generations"))
}

func TestIsHistoricalBatchImageRead(t *testing.T) {
	if !isHistoricalBatchImageRead(http.MethodGet, "/v1/images/batches/batch_123") {
		t.Fatal("expected batch history GET to bypass live routing")
	}
	if !isHistoricalBatchImageRead(http.MethodGet, "/images/batches/batch_123/items") {
		t.Fatal("expected batch item history GET to bypass live routing")
	}
	if isHistoricalBatchImageRead(http.MethodGet, "/v1/images/batches/models") {
		t.Fatal("models discovery must continue through smart routing")
	}
	if isHistoricalBatchImageRead(http.MethodPost, "/v1/images/batches") {
		t.Fatal("batch submission must continue through smart routing")
	}
	if !isHistoricalBatchImageRead(http.MethodPost, "/v1/images/batches/batch_123/cancel") {
		t.Fatal("batch cancellation should remain owner-scoped")
	}
	if !isHistoricalBatchImageRead(http.MethodDelete, "/v1/images/batches/batch_123/outputs") {
		t.Fatal("batch output deletion should remain owner-scoped")
	}
}
