package audit

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestHTTPSink_Notify_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, http.MethodPost, r.Method)
		require.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sink := NewHTTPSink(server.URL)
	err := sink.Notify(Event{
		TS:        1,
		Metrics:   []string{"Alloc"},
		IPAddress: "192.168.0.107",
	})
	require.NoError(t, err)
}

func TestHTTPSink_Notify_BadStatus(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	sink := NewHTTPSink(server.URL)
	err := sink.Notify(Event{TS: 1, Metrics: []string{"Alloc"}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "audit post status: 400")
}

func TestHTTPSink_Notify_RetryThenSuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sink := NewHTTPSink(server.URL)
	err := sink.Notify(Event{TS: 1, Metrics: []string{"Alloc"}})
	require.NoError(t, err)
	require.GreaterOrEqual(t, attempts, 2)
}
