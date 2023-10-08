package cache

import (
	"encoding"
	"volo_meeting/lib/db/redis"
)

var instance = redis.Instance()

func Init() {
	instance = redis.Instance()
}

type MarshalAble interface {
	encoding.BinaryMarshaler
	encoding.BinaryUnmarshaler
}

type Z interface {
	ToZ() redis.Z
}

type MapAble interface {
	ToMap() map[string]any
}
