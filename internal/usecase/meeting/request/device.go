package request

import (
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"volo_meeting/internal/cache"
	error2 "volo_meeting/lib/error"
)

var (
	_ cache.Z       = (*Device)(nil)
	_ cache.MapAble = (*Device)(nil)
)

type Device struct {
	Id       string `json:"id"`
	Nickname string `json:"nickname"`
	Platform string `json:"platform"`
	JoinTime uint64 `json:"join_time"`
}

func (d *Device) ToMap() map[string]any {
	return map[string]any{
		"id":        d.Id,
		"nickname":  d.Nickname,
		"platform":  d.Platform,
		"join_time": d.JoinTime,
	}
}

func (d *Device) FromMap(m map[string]any) error {
	ok := m["id"] != nil && m["nickname"] != nil && m["platform"] != nil && m["join_time"] != nil
	if !ok {
		return error2.InvalidTypeAssert
	}

	d.Id, ok = m["id"].(string)
	d.Nickname, ok = m["nickname"].(string)
	d.Platform, ok = m["platform"].(string)
	d.JoinTime, ok = m["join_time"].(uint64)

	if !ok {
		return error2.InvalidTypeAssert
	}
	return nil
}

func (d *Device) ToZ() redis.Z {
	return redis.Z{
		Score:  float64(d.JoinTime),
		Member: d.Id,
	}
}

func (d *Device) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, d)
}

func (d *Device) MarshalBinary() (data []byte, err error) {
	return jsoniter.Marshal(d)
}
