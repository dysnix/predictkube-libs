package memory

import (
	"context"

	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"

	c "github.com/dysnix/predictkube-libs/external/cache"
	"github.com/dysnix/predictkube-libs/external/configs"
)

var (
	infoMsg = "💡 Memory cache connections success..."

	_ c.Cache = (*Cache)(nil)
)

type Cache struct {
	conf   configs.CacheGetter
	cache  *cache.Cache
	logger *zap.SugaredLogger
}

func NewCache(options ...configs.CacheOption) (result *Cache, err error) {
	result = &Cache{}

	for _, op := range options {
		err := op(result)
		if err != nil {
			return nil, err
		}
	}

	result.cache = cache.New(result.conf.GetCache().GlobalTTL.TTL, result.conf.GetCache().Memory.CleanupInterval)

	result.logger.Info(infoMsg)

	return result, nil
}

func (c *Cache) SetCache(ch configs.CacheGetter) {
	c.conf = ch
}

func (c *Cache) SetLogger(logger *zap.SugaredLogger) {
	c.logger = logger
}

func (c *Cache) Ping(_ context.Context) error {
	return nil
}

func (c *Cache) Stop() error {
	c.cache.Flush()
	return nil
}
