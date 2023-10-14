package cache

import (
	"context"
	"reflect"
	"time"
	"volo_meeting/consts"
	"volo_meeting/lib/db/redis"
	error2 "volo_meeting/lib/error"
)

const structTag = "redis"

// ToMap : only convert struct field with 'redis' tag to map[string]any
func ToMap(item any) map[string]any {
	val := reflect.ValueOf(item).Elem()
	typ := val.Type()
	if typ.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]any)
	for i := 0; i < val.NumField(); i++ {
		tag := typ.Field(i).Tag.Get(structTag)
		if tag != "" && tag != "-" {
			result[tag] = val.Field(i).Interface()
		}
	}

	return result
}

func wrap(err error) error {
	return error2.New(consts.WSError, err)
}

func ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	data, err := instance.ZRevRange(ctx, key, start, stop).Result()
	return data, wrap(err)
}

func ZAdd(ctx context.Context, key string, members []redis.Z) error {

	return wrap(instance.ZAdd(ctx, key, members...).Err())
}

func ZRem(ctx context.Context, key string, members []string) error {
	return wrap(instance.ZRem(ctx, key, members).Err())
}

func SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	data, err := instance.SetNX(ctx, key, value, expiration).Result()
	return data, wrap(err)
}

func Get(ctx context.Context, key string) (string, error) {
	data, err := instance.Get(ctx, key).Result()
	return data, wrap(err)
}

func Del(ctx context.Context, keys ...string) error {
	return wrap(instance.Del(ctx, keys...).Err())
}

func HSet(ctx context.Context, key string, value map[string]any) error {
	return wrap(instance.HSet(ctx, key, value).Err())
}

// HGet : data must be a pointer
func HGet(ctx context.Context, key string, data any) error {
	return wrap(instance.HGetAll(ctx, key).Scan(data))
}
