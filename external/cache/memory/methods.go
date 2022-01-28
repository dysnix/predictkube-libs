package memory

import (
	"context"
	"encoding/json"
	"time"

	ch "github.com/dysnix/predictkube-libs/external/cache"
)

func (c *Cache) Set(_ context.Context, object interface{}, key string, ttl time.Duration) (err error) {
	defer func() {
		err = ch.ActionRecover(err)
	}()

	if object != nil {
		jsonStr, err := json.Marshal(object)
		if err != nil {
			return err
		}

		c.cache.Set(key, string(jsonStr), ttl)

		return nil
	}

	return ch.ErrEmptyObject
}

func (c *Cache) KeysCount(_ context.Context) (count int, err error) {
	defer func() {
		err = ch.ActionRecover(err)
	}()

	return len(c.cache.Items()), nil
}

func (c *Cache) Get(_ context.Context, key string, object interface{}) (err error) {
	defer func() {
		err = ch.ActionRecover(err)
	}()

	resp, ok := c.cache.Get(key)
	if !ok {
		return ch.ErrNil
	}

	return json.Unmarshal([]byte(resp.(string)), object)
}

func (c *Cache) Delete(_ context.Context, keys ...string) (err error) {
	defer func() {
		err = ch.ActionRecover(err)
	}()

	for _, key := range keys {
		c.cache.Delete(key)
	}

	return nil
}
