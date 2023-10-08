package cache

import (
	"context"
	"reflect"
	"time"
	"volo_meeting/lib/db/redis"
)

const structTag = "redis"

// ToMap : only convert struct field with 'redis' tag to map
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

func ZRevRange(ctx context.Context, key string, start, stop int64) ([]string, error) {
	return instance.ZRevRange(ctx, key, start, stop).Result()
}

func ZAdd(ctx context.Context, key string, members []redis.Z) error {
	return instance.ZAdd(ctx, key, members...).Err()
}

func ZRem(ctx context.Context, key string, members []string) error {
	return instance.ZRem(ctx, key, members).Err()
}

func SetNX(ctx context.Context, key string, value string, expiration time.Duration) (bool, error) {
	return instance.SetNX(ctx, key, value, expiration).Result()
}

func Get(ctx context.Context, key string) (string, error) {
	return instance.Get(ctx, key).Result()
}

func Del(ctx context.Context, keys ...string) error {
	return instance.Del(ctx, keys...).Err()
}

func HSet(ctx context.Context, key string, value map[string]any) error {
	return instance.HSet(ctx, key, value).Err()
}

// HGet : data must be a pointer
func HGet(ctx context.Context, key string, data any) error {
	return instance.HGetAll(ctx, key).Scan(data)
}
