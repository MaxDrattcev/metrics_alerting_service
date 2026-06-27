package audit

import (
	"context"
	"errors"
	"net"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestIsRetriableAuditError(t *testing.T) {
	require.True(t, isRetriableAuditError(&net.OpError{Op: "dial", Err: syscall.ECONNREFUSED}))
	require.False(t, isRetriableAuditError(context.DeadlineExceeded)) // было True — неверно
	require.False(t, isRetriableAuditError(context.Canceled))
	require.False(t, isRetriableAuditError(errors.New("plain error")))
	require.True(t, isRetriableAuditError(syscall.ECONNREFUSED))
}

func TestRetryHTTPClient_Do_Success(t *testing.T) {
	client := newRetryHTTPClient(2 * time.Second)

	require.NotNil(t, client)
}
