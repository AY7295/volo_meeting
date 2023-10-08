package cache

import (
	"context"
	"fmt"
	jsoniter "github.com/json-iterator/go"
	"github.com/redis/go-redis/v9"
	"testing"
	"time"
	"volo_meeting/lib/id"
)

type Device struct {
	Id       string `json:"id" redis:"id"`
	Nickname string `json:"nickname" redis:"nickname"`
	Platform string `json:"platform" redis:"platform"`
	JoinTime uint64 `json:"join_time" redis:"-"`
}

func (d *Device) UnmarshalBinary(data []byte) error {
	return jsoniter.Unmarshal(data, d)
}

func (d *Device) MarshalBinary() (data []byte, err error) {
	return jsoniter.Marshal(d)
}

func (d *Device) ToMap() map[string]any {
	return map[string]any{
		"id":        d.Id,
		"nickname":  d.Nickname,
		"platform":  d.Platform,
		"join_time": d.JoinTime,
	}
}

func (d *Device) ToZ() redis.Z {
	return redis.Z{
		Score:  float64(d.JoinTime),
		Member: d.Id,
	}
}

var devices = []*Device{
	{
		Id:       "dC6yLnXeLLY1FYKWU6sZ9",
		Nickname: "n0",
		Platform: "p0",
		JoinTime: 0,
	},
	{
		Id:       "dC6yLnXeLLY1FYKWU6sZ9",
		Nickname: "n1",
		Platform: "p1",
		JoinTime: 1,
	},
	{
		Id:       id.Must(),
		Nickname: "n2",
		Platform: "p2",
		JoinTime: 2,
	},
	{
		Id:       id.Must(),
		Nickname: "n3",
		Platform: "p3",
		JoinTime: 3,
	},
	{
		Id:       id.Must(),
		Nickname: "n4",
		Platform: "p4",
		JoinTime: 4,
	},
	{
		Id:       id.Must(),
		Nickname: "n5",
		Platform: "p5",
		JoinTime: 5,
	},
}

var meetingId = "IVzGgV2MMxCsim2yR5q7J"
var friendlyId = "75710399"

func TestSetNX(t *testing.T) {
	type args struct {
		ctx    context.Context
		key    string
		value  string
		expire time.Duration
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "success1", args: args{ctx: context.TODO(), key: "k1", value: "v1", expire: time.Second * 10000}},
		{name: "fail", args: args{ctx: context.TODO(), key: "k1", value: "v3", expire: time.Second * 100}},
		{name: "success2", args: args{ctx: context.TODO(), key: friendlyId, value: meetingId, expire: time.Hour * 24 * 30}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SetNX(tt.args.ctx, tt.args.key, tt.args.value, tt.args.expire)
			if err != nil {
				panic(err)
			}
			fmt.Println(got)
		})
	}
}

func TestGet(t *testing.T) {
	type args[T MarshalAble] struct {
		ctx context.Context
		key string
	}
	type testCase[T MarshalAble] struct {
		name    string
		args    args[T]
		wantErr bool
	}
	device := &Device{}
	tests := []testCase[*Device]{
		{name: "success", args: args[*Device]{ctx: context.TODO(), key: "k12"}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := Get(tt.args.ctx, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("Get() error = %v, wantErr %v", err, tt.wantErr)
			}
			err = device.UnmarshalBinary([]byte(res))
			if err != nil {
				panic(err)
			}
			fmt.Println(device)
		})
	}
}

func getDeviceZs() []redis.Z {
	zs := make([]redis.Z, len(devices))
	for i, device := range devices {
		zs[i] = device.ToZ()
	}

	return zs
}

func TestZAdd(t *testing.T) {
	type args[T Z] struct {
		ctx     context.Context
		key     string
		members []redis.Z
	}
	type testCase[T Z] struct {
		name    string
		args    args[T]
		wantErr bool
	}
	tests := []testCase[*Device]{
		{name: "success", args: args[*Device]{ctx: context.TODO(), key: meetingId, members: getDeviceZs()}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ZAdd(tt.args.ctx, tt.args.key, tt.args.members); (err != nil) != tt.wantErr {
				t.Errorf("ZAdd() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestZRem(t *testing.T) {
	type args struct {
		ctx     context.Context
		key     string
		members []string
	}
	type testCase struct {
		name    string
		args    args
		wantErr bool
	}
	tests := []testCase{
		{name: "success", args: args{ctx: context.TODO(), key: meetingId, members: []string{devices[0].Id, devices[3].Id}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ZRem(tt.args.ctx, tt.args.key, tt.args.members); (err != nil) != tt.wantErr {
				t.Errorf("ZRem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestZRevRange(t *testing.T) {
	type args struct {
		ctx   context.Context
		key   string
		start int64
		stop  int64
	}
	type testCase struct {
		name    string
		args    args
		wantErr bool
	}
	tests := []testCase{
		{name: "success", args: args{ctx: context.TODO(), key: meetingId, start: 0, stop: -1}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			all, err := ZRevRange(tt.args.ctx, tt.args.key, tt.args.start, tt.args.stop)
			if (err != nil) != tt.wantErr {
				t.Errorf("ZRevRange() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Println(all)
		})
	}
}

func TestDel(t *testing.T) {
	type args struct {
		ctx  context.Context
		keys []string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "success", args: args{ctx: context.TODO(), keys: []string{meetingId}}, wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Del(tt.args.ctx, tt.args.keys...); (err != nil) != tt.wantErr {
				t.Errorf("Del() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHSet(t *testing.T) {
	type args struct {
		ctx   context.Context
		key   string
		value map[string]any
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "t1", args: args{ctx: context.TODO(), key: devices[0].Id, value: devices[0].ToMap()}},
		{name: "t2", args: args{ctx: context.TODO(), key: devices[1].Id, value: devices[1].ToMap()}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HSet(tt.args.ctx, tt.args.key, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("HSet() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestHGet(t *testing.T) {
	type args struct {
		ctx  context.Context
		key  string
		data any
	}
	device := &Device{}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{name: "t1", args: args{ctx: context.TODO(), key: devices[0].Id, data: device}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := HGet(tt.args.ctx, tt.args.key, tt.args.data); (err != nil) != tt.wantErr {
				t.Errorf("HGet() error = %v, wantErr %v", err, tt.wantErr)
			}
			fmt.Println(device)
		})
	}
}

func TestToMap(t *testing.T) {
	type args struct {
		item any
	}
	tests := []struct {
		name string
		args args
	}{
		{name: "t1", args: args{item: devices[0]}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ToMap(tt.args.item)
			fmt.Println(got)
		})
	}
}
