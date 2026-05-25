package audit

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

var auditRetryDelays = []time.Duration{
	time.Second,
	3 * time.Second,
	5 * time.Second,
}

// httpDoer — минимальный контракт для HTTPSink (обычный или с ретраями).
type httpDoer interface {
	Do(req *http.Request) (*http.Response, error)
}

type retryHTTPClient struct {
	inner *http.Client
}

func newRetryHTTPClient(timeout time.Duration) httpDoer {
	return &retryHTTPClient{
		inner: &http.Client{Timeout: timeout},
	}
}

func (c *retryHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var body []byte
	if req.Body != nil {
		var err error
		body, err = io.ReadAll(req.Body)
		_ = req.Body.Close()
		if err != nil {
			return nil, err
		}
	}

	ctx := req.Context()
	if ctx == nil {
		ctx = context.Background()
	}

	var lastResp *http.Response
	var lastErr error

	for attempt := 0; ; attempt++ {
		reqAttempt := req.Clone(ctx)
		if reqAttempt == nil {
			return nil, errors.New("audit: failed to clone request")
		}
		if body != nil {
			reqAttempt.Body = io.NopCloser(bytes.NewReader(body))
			reqAttempt.ContentLength = int64(len(body))
		}

		resp, err := c.inner.Do(reqAttempt)
		lastResp, lastErr = resp, err

		needRetry := false
		if err != nil {
			needRetry = isRetriableAuditError(err)
		} else if resp != nil {
			sc := resp.StatusCode
			needRetry = sc == http.StatusTooManyRequests || sc >= http.StatusInternalServerError
		}

		if !needRetry {
			return lastResp, lastErr
		}

		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		if attempt >= len(auditRetryDelays) {
			return lastResp, lastErr
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		t := time.NewTimer(auditRetryDelays[attempt])
		select {
		case <-ctx.Done():
			t.Stop()
			return nil, ctx.Err()
		case <-t.C:
		}
	}
}

func isRetriableAuditError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var ne net.Error
	if errors.As(err, &ne) && ne.Timeout() {
		return true
	}

	var oe *net.OpError
	if errors.As(err, &oe) {
		return true
	}

	var ose *os.SyscallError
	if errors.As(err, &ose) {
		return isRetriableAuditError(ose.Err)
	}

	var se syscall.Errno
	if errors.As(err, &se) {
		switch se {
		case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.EPIPE,
			syscall.ETIMEDOUT, syscall.EHOSTUNREACH, syscall.ENETUNREACH:
			return true
		}
	}

	return false
}
