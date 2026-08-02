package leakybucket

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrLimited 表示请求被漏桶拒绝（HTTP 层应返回 429 + Retry-After）。
var ErrLimited = errors.New("leaky bucket: rate limited")

// leakyBucketLua 对应笔记 §5.1：所有实例共享同一个 key，Lua 保证
// 「读 last → 计算 → 写 last」原子性；时间戳取 Redis 服务器时间，
// 避免各实例时钟漂移让空闲判定失真。
const leakyBucketLua = `
-- KEYS[1] = 限流 key（值为上次放行时刻 last，毫秒）
-- ARGV[1] = perRequest（放行间隔毫秒，= 1000 / rate）
-- ARGV[2] = capacity（允许排队的最大请求数）
-- 返回 {1, waitMs} 放行（waitMs 为需等待时长）；{0, waitMs} 拒绝（waitMs 可作为 Retry-After）
local t = redis.call('TIME')
local now = tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000)
local last = tonumber(redis.call('GET', KEYS[1]) or '0')
local perRequest = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])

if last == 0 or now >= last + perRequest then
    redis.call('SET', KEYS[1], now)
    return {1, 0}
end

local wait = last + perRequest - now
if wait > capacity * perRequest then
    return {0, wait}
end

redis.call('SET', KEYS[1], last + perRequest)
return {1, wait}
`

// DistributedBucket 分布式虚拟队列漏桶（笔记 §5.1）：状态在 Redis，
// 多实例共享同一个 key；capacity 参数即虚拟队列的排队上限，天然
// 限住排队数（补上单机版「等待无上限」的短板）。
type DistributedBucket struct {
	rdb        redis.Cmdable
	key        string
	perRequest time.Duration
	capacity   int64
}

// NewDistributedBucket 新建分布式虚拟队列漏桶，rate 为每秒放行数，
// capacity 为允许排队的最大请求数（超限立即拒绝）。
func NewDistributedBucket(rdb redis.Cmdable, key string, rate int, capacity int64) *DistributedBucket {
	return &DistributedBucket{
		rdb:        rdb,
		key:        key,
		perRequest: time.Second / time.Duration(rate),
		capacity:   capacity,
	}
}

// Take 原子入队判定：放行但需等待时阻塞到放行时刻；拒绝返回 ErrLimited；
// 等待期间 ctx 取消则返回 ctx.Err()。
func (b *DistributedBucket) Take(ctx context.Context) error {
	res, err := b.rdb.Eval(ctx, leakyBucketLua, []string{b.key},
		b.perRequest.Milliseconds(), b.capacity).Result()
	if err != nil {
		return err
	}
	vals := res.([]interface{})
	if vals[0].(int64) != 1 {
		return ErrLimited // 被拒绝时 vals[1] 是建议等待时长，HTTP 层作 Retry-After
	}
	if wait := time.Duration(vals[1].(int64)) * time.Millisecond; wait > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err() // 调用方取消，不再等
		case <-time.After(wait):
		}
	}
	return nil
}
