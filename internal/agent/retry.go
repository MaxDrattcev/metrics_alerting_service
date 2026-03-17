package agent

import (
	"context"
	"errors"
	"net"
	"os"
	"syscall"
	"time"
)

var agentRetryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func doWithAgentRetries(ctx context.Context, op func(context.Context) error) error {
	err := op(ctx)
	if err == nil {
		return nil
	}

	for _, d := range agentRetryDelays {
		if ctx.Err() != nil {
			return ctx.Err()
		}

		if !isRetriableTransportError(err) {
			return err
		}

		t := time.NewTimer(d)
		select {
		case <-ctx.Done():
			t.Stop()
			return ctx.Err()
		case <-t.C:
		}

		err = op(ctx)
		if err == nil {
			return nil
		}
	}

	return err
}

func isRetriableTransportError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var ne net.Error
	if errors.As(err, &ne) {
		if ne.Timeout() || ne.Temporary() {
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
