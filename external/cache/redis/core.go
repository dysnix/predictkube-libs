package redis

import (
	"context"

	"github.com/go-redis/redis/v8"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	c "github.com/dysnix/predictkube-libs/external/cache"
	"github.com/dysnix/predictkube-libs/external/configs"
)

const (
	infoMsg = "💡 Redis cache connections success..."
)

var _ c.Cache = (*cache)(nil)

type cache struct {
	writeClient  *redis.ClusterClient
	readClient   *redis.ClusterClient
	singleClient redis.UniversalClient
	conf         configs.CacheGetter
	logger       *zap.SugaredLogger
}

func NewCache(options ...configs.CacheOption) (c c.Cache, err error) {
	cache := &cache{}

	for _, op := range options {
		err := op(cache)
		if err != nil {
			return nil, err
		}
	}

	if err = cache.connectCluster(context.Background()); err != nil {
		if err = cache.connectSingle(context.Background()); err != nil {
			return nil, err
		}
	}

	return cache, nil
}

func (c *cache) connectSingle(ctx context.Context) (err error) {
	if c.writeClient != nil {
		if err = c.writeClient.Close(); err != nil {
			c.logger.Debug(err)
		}

		c.writeClient = nil
	}

	if c.readClient != nil {
		if err = c.readClient.Close(); err != nil {
			c.logger.Debug(err)
		}

		c.readClient = nil
	}

	options := redis.Options{
		DB:                 int(c.conf.GetCache().Redis.DB),
		MaxRetries:         c.conf.GetCache().Redis.MaxRetries,
		MinRetryBackoff:    c.conf.GetCache().Redis.MinRetryBackoff,
		MaxRetryBackoff:    c.conf.GetCache().Redis.MaxRetryBackoff,
		DialTimeout:        c.conf.GetCache().Redis.DialTimeout,
		ReadTimeout:        c.conf.GetCache().Redis.ReadTimeout,
		WriteTimeout:       c.conf.GetCache().Redis.WriteTimeout,
		PoolSize:           c.conf.GetCache().Redis.Pool.PoolSize,
		PoolTimeout:        c.conf.GetCache().Redis.Pool.PoolTimeout,
		IdleTimeout:        c.conf.GetCache().Redis.Pool.IdleTimeout,
		IdleCheckFrequency: c.conf.GetCache().Redis.Pool.IdleCheckFrequency,
		MaxConnAge:         c.conf.GetCache().Redis.Pool.MaxConnAge,
		MinIdleConns:       c.conf.GetCache().Redis.Pool.MinIdleConns,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			c.logger.Debug("New redis pool read connection event")

			return ctx.Err()
		},
	}

	options.Addr = c.conf.GetCache().Redis.WriteAddrs[0]
	options.Password = c.conf.GetCache().Redis.Password
	options.Username = c.conf.GetCache().Redis.Username

	c.singleClient = redis.NewClient(&options)

	redis.SetLogger(&zapLogger{
		c.logger,
	})

	_, err = c.singleClient.Ping(ctx).Result()
	if err != nil {
		return err
	}

	c.logger.Info(ctx, infoMsg)

	return nil
}

func (c *cache) connectCluster(ctx context.Context) (err error) {
	readOptions := &redis.ClusterOptions{
		MaxRetries:         c.conf.GetCache().Redis.MaxRetries,
		MaxRedirects:       c.conf.GetCache().Redis.MaxRedirects,
		MinRetryBackoff:    c.conf.GetCache().Redis.MinRetryBackoff,
		MaxRetryBackoff:    c.conf.GetCache().Redis.MaxRetryBackoff,
		DialTimeout:        c.conf.GetCache().Redis.DialTimeout,
		ReadTimeout:        c.conf.GetCache().Redis.ReadTimeout,
		WriteTimeout:       c.conf.GetCache().Redis.WriteTimeout,
		PoolSize:           c.conf.GetCache().Redis.Pool.PoolSize,
		PoolTimeout:        c.conf.GetCache().Redis.Pool.PoolTimeout,
		IdleTimeout:        c.conf.GetCache().Redis.Pool.IdleTimeout,
		IdleCheckFrequency: c.conf.GetCache().Redis.Pool.IdleCheckFrequency,
		MaxConnAge:         c.conf.GetCache().Redis.Pool.MaxConnAge,
		MinIdleConns:       c.conf.GetCache().Redis.Pool.MinIdleConns,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			c.logger.Debug("New redis pool read connection event")

			return ctx.Err()
		},
		ReadOnly: true,
	}

	readOptions.Addrs = c.conf.GetCache().Redis.ReadAddrs
	readOptions.Password = c.conf.GetCache().Redis.Password
	readOptions.Username = c.conf.GetCache().Redis.Username

	c.readClient = redis.NewClusterClient(readOptions)

	writeOptions := &redis.ClusterOptions{
		MaxRetries:         c.conf.GetCache().Redis.MaxRetries,
		MaxRedirects:       c.conf.GetCache().Redis.MaxRedirects,
		MinRetryBackoff:    c.conf.GetCache().Redis.MinRetryBackoff,
		MaxRetryBackoff:    c.conf.GetCache().Redis.MaxRetryBackoff,
		DialTimeout:        c.conf.GetCache().Redis.DialTimeout,
		ReadTimeout:        c.conf.GetCache().Redis.ReadTimeout,
		WriteTimeout:       c.conf.GetCache().Redis.WriteTimeout,
		PoolSize:           c.conf.GetCache().Redis.Pool.PoolSize,
		PoolTimeout:        c.conf.GetCache().Redis.Pool.PoolTimeout,
		IdleTimeout:        c.conf.GetCache().Redis.Pool.IdleTimeout,
		IdleCheckFrequency: c.conf.GetCache().Redis.Pool.IdleCheckFrequency,
		MaxConnAge:         c.conf.GetCache().Redis.Pool.MaxConnAge,
		MinIdleConns:       c.conf.GetCache().Redis.Pool.MinIdleConns,
		OnConnect: func(ctx context.Context, cn *redis.Conn) error {
			c.logger.Debug("New redis pool write connection event")

			return ctx.Err()
		},
	}

	readOptions.Addrs = c.conf.GetCache().Redis.WriteAddrs
	writeOptions.Password = c.conf.GetCache().Redis.Password
	writeOptions.Username = c.conf.GetCache().Redis.Username

	c.writeClient = redis.NewClusterClient(writeOptions)

	redis.SetLogger(&zapLogger{
		c.logger,
	})

	if err = c.Ping(ctx); err != nil {
		return err
	}

	c.logger.Info(ctx, infoMsg)

	return nil
}

func (c *cache) SetCache(cache configs.CacheGetter) {
	c.conf = cache
}

func (c *cache) SetLogger(logger *zap.SugaredLogger) {
	c.logger = logger
}

func (c *cache) Stop() (err error) {
	defer func() {
		if c.writeClient != nil {
			var oldErr error
			if err != nil {
				oldErr = err
			}

			err = c.writeClient.Close()
			if oldErr != nil {
				err = errors.WithMessage(oldErr, err.Error())
			}
		}
	}()

	if c.readClient != nil {
		return c.readClient.Close()
	}

	return c.singleClient.Close()
}

func (c *cache) Ping(ctx context.Context) (err error) {
	if c.writeClient != nil && c.readClient != nil {
		eg, _ := errgroup.WithContext(ctx)
		eg.Go(func() error {
			return c.readClient.Ping(ctx).Err()
		})
		eg.Go(func() error {
			return c.writeClient.Ping(ctx).Err()
		})

		return eg.Wait()
	}

	return c.singleClient.Ping(ctx).Err()
}
