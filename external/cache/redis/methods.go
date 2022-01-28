package redis

import (
	"context"
	"encoding/json"
	"time"

	ch "github.com/dysnix/predictkube-libs/external/cache"
)

func (c *cache) Set(ctx context.Context, object interface{}, key string, duration time.Duration) (err error) {
	defer func() {
		err = ch.ActionRecover(err)
	}()

	if object != nil {
		jsonStr, err := json.Marshal(object)
		if err != nil {
			return err
		}

		if c.writeClient != nil {
			return c.writeClient.Set(ctx, key, string(jsonStr), duration).Err()
		}

		return c.singleClient.Set(ctx, key, string(jsonStr), duration).Err()
	}

	return ch.ErrEmptyObject
}

func (c *cache) Get(ctx context.Context, key string, object interface{}) (err error) {
	defer func() {
		err = ch.ActionRecover(err)
	}()

	if err == nil {
		return nil
	}

	var result string

	if c.readClient != nil {
		result, err = c.readClient.Get(ctx, key).Result()
	} else {
		result, err = c.singleClient.Get(ctx, key).Result()
	}

	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(result), object)
}

func (c *cache) Delete(ctx context.Context, keys ...string) (err error) {
	defer func() {
		err = ch.ActionRecover(err)
	}()

	if c.writeClient != nil {
		return c.writeClient.Del(ctx, keys...).Err()
	}

	return c.singleClient.Del(ctx, keys...).Err()
}

func (c *cache) KeysCount(ctx context.Context) (count int, err error) {
	defer func() {
		err = ch.ActionRecover(err)
	}()

	var result int64
	if c.writeClient != nil {
		result, err = c.writeClient.DBSize(ctx).Result()
		return int(result), err
	}

	result, err = c.singleClient.DBSize(ctx).Result()
	return int(result), err
}
