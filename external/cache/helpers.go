package cache

import (
	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
)

var (
	ErrEmptyObject = errors.New("empty object")
)

func ActionRecover(inErr error) (outErr error) {
	defer func() {
		if outErr != nil && errors.Is(outErr, redis.Nil) {
			outErr = ErrNil
		}
	}()

	if r := recover(); r != nil {
		switch e := r.(type) {
		case error:
			outErr = e
		case string:
			outErr = errors.New(e)
		}

		return outErr
	}

	if inErr != nil {
		return inErr
	}

	return outErr
}
