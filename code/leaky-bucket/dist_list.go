package leakybucket

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// listBucketLua 对应笔记 §5.2：惰性补账——入队前先把「按匀速该出的队」
// 出掉，所有实例共享 last 出队时刻，避免多个出队者竞争加速。
const listBucketLua = `
-- KEYS[1] = 桶 list，KEYS[2] = last 出队时刻（毫秒）
-- ARGV[1] = perRequest（毫秒），ARGV[2] = capacity，ARGV[3] = job id
local t = redis.call('TIME')
local now = tonumber(t[1]) * 1000 + math.floor(tonumber(t[2]) / 1000)
local last = tonumber(redis.call('GET', KEYS[2]) or '0')
local perRequest = tonumber(ARGV[1])
local capacity = tonumber(ARGV[2])

if last > 0 then
    local drained = math.floor((now - last) / perRequest) -- 流逝时间应出掉的数量
    for i = 1, drained do
        redis.call('RPOP', KEYS[1])
    end
    redis.call('SET', KEYS[2], last + drained * perRequest)
else
    redis.call('SET', KEYS[2], now)
end

if redis.call('LLEN', KEYS[1]) >= capacity then
    return 0 -- 桶满：拒绝
end
redis.call('LPUSH', KEYS[1], ARGV[3])
return 1
`

// ListBucket 分布式队列式漏桶（笔记 §5.2）：Redis List 当桶，
// 惰性补账保证桶里积压的队形符合匀速。注意：出掉的 job 仍需
// 有人执行（调度器定时 RPOP + 派发，或直接上消息队列）。
type ListBucket struct {
	rdb        redis.Cmdable
	key        string // 桶 list
	lastKey    string // last 出队时刻
	perRequest time.Duration
	capacity   int64
}

// NewListBucket 新建 Redis List 桶式漏桶，rate 为每秒放行数。
func NewListBucket(rdb redis.Cmdable, key string, rate int, capacity int64) *ListBucket {
	return &ListBucket{
		rdb:        rdb,
		key:        key,
		lastKey:    key + ":last",
		perRequest: time.Second / time.Duration(rate),
		capacity:   capacity,
	}
}

// Enqueue 入队（先惰性出队）；桶满返回 false。
func (b *ListBucket) Enqueue(ctx context.Context, jobID string) (bool, error) {
	res, err := b.rdb.Eval(ctx, listBucketLua, []string{b.key, b.lastKey},
		b.perRequest.Milliseconds(), b.capacity, jobID).Result()
	if err != nil {
		return false, err
	}
	return res.(int64) == 1, nil
}
