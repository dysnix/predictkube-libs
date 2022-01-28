package cache

import (
	"context"
	"time"

	"github.com/pkg/errors"

	"github.com/dysnix/predictkube-libs/external/configs"
)

var (
	ErrNil = errors.New("not found record")
)

type Value string

type Cache interface {
	configs.SignalStopperWithErr
	Ping(ctx context.Context) error
	Set(context.Context, interface{}, string, time.Duration) error
	Get(context.Context, string, interface{}) error
	Delete(context.Context, ...string) error
	KeysCount(context.Context) (int, error)
}
