package pool

import (
	"context"
	"errors"
	"net/http"

	"godruid/internal/errorsx"
)

var (
	ErrPoolClosed   = errorsx.New(errorsx.CodePoolClosed, "pool is closed", http.StatusConflict)
	ErrWaitTimeout  = errorsx.New(errorsx.CodeWaitTimeout, "wait for connection timed out", 408)
	ErrCanceled     = errorsx.New(errorsx.CodeCanceled, "wait for connection canceled", 499)
	ErrInvalidPut   = errorsx.New(errorsx.CodeConflict, "connection is not borrowed from this pool", http.StatusConflict)
	ErrDoublePut    = errorsx.New(errorsx.CodeConflict, "connection already returned or not in use", http.StatusConflict)
	ErrCrossPool    = errorsx.New(errorsx.CodeConflict, "connection belongs to another pool", http.StatusConflict)
	ErrNilConn      = errorsx.New(errorsx.CodeInvalidArgument, "connection is nil", http.StatusBadRequest)
	ErrCapacity     = errorsx.New(errorsx.CodeConflict, "pool at max active capacity", http.StatusConflict)
)

func MapWaitError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return ErrWaitTimeout
	}
	if errors.Is(err, context.Canceled) {
		return ErrCanceled
	}
	return err
}
