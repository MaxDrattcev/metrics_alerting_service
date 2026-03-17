package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgconn"
)

var dbRetryDelays = []time.Duration{1 * time.Second, 3 * time.Second, 5 * time.Second}

func doWithDBRetries(ctx context.Context, op func(context.Context) error) error {
	err := op(ctx)
	if err == nil {
		return nil
	}

	for _, d := range dbRetryDelays {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if !isRetriablePGTransportError(err) {
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

func isRetriablePGTransportError(err error) bool {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return strings.HasPrefix(pgErr.Code, "08")
	}

	return false
}
