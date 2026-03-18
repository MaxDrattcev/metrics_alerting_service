package agent

import (
	"context"
	"errors"
	"github.com/go-resty/resty/v2"
	"net"
	"net/http"
	"os"
	"syscall"
	"time"
)

var retryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

type RetryableClient struct {
	*resty.Client
}

func NewRetryableClient() *RetryableClient {
	c := resty.New()
	c.SetTimeout(5 * time.Second)
	return &RetryableClient{Client: c}
}

func (r *RetryableClient) PostWithRetry(ctx context.Context, url string, headers map[string]string, body []byte) (*resty.Response, error) {
	var lastResp *resty.Response
	var lastErr error

	for attempt := 0; ; attempt++ {
		req := r.R().SetContext(ctx).SetBody(body)
		for k, v := range headers {
			req.SetHeader(k, v)
		}
		resp, err := req.Post(url)

		lastResp, lastErr = resp, err

		needRetry := false
		if err != nil {
			needRetry = isRetriableTransportError(err)
		} else if resp != nil {
			sc := resp.StatusCode()
			needRetry = sc == http.StatusTooManyRequests || sc >= 500
		}

		if !needRetry {
			return lastResp, lastErr
		}

		if attempt >= len(retryDelays) {
			return lastResp, lastErr
		}

		if ctx.Err() != nil {
			return nil, ctx.Err()
		}

		t := time.NewTimer(retryDelays[attempt])
		select {
		case <-ctx.Done():
			t.Stop()
			return nil, ctx.Err()
		case <-t.C:
		}
	}
}

func isRetriableTransportError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() {
			return true
		}
	}

	var oe *net.OpError
	if errors.As(err, &oe) {
		return true
	}

	var ose *os.SyscallError
	if errors.As(err, &ose) {
		return isRetriableTransportError(ose.Err)
	}

	var se syscall.Errno
	if errors.As(err, &se) {
		switch se {
		case syscall.ECONNREFUSED, syscall.ECONNRESET, syscall.EPIPE, syscall.ETIMEDOUT, syscall.EHOSTUNREACH, syscall.ENETUNREACH:
			return true
		}
	}

	return false
}
