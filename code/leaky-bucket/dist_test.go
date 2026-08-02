package leakybucket

import (
	"context"
	"sort"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

// newTestRDB 起一个 miniredis（纯 Go 内存 Redis，支持 EVAL/Lua），
// 分布式测试不需要外部 Redis 服务；关闭由 t.Cleanup 统一处理。
func newTestRDB(t *testing.T) *redis.Client {
	t.Helper()
	s := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: s.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	return rdb
}

// TestDistributedSerialPass 验证 §5.1：3 个实例（3 个 goroutine）
// 并发请求共享同一 key，1 个立即放行、其余按 perRequest 间隔
// 依次放行（第一个不排队，后两个各等 ~20ms）。
func TestDistributedSerialPass(t *testing.T) {
	rdb := newTestRDB(t)

	b := NewDistributedBucket(rdb, "lb:serial", 50, 10) // perRequest=20ms
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const n = 3
	done := make(chan time.Time, n)
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := b.Take(ctx); err != nil {
				t.Errorf("Take: %v", err)
				return
			}
			done <- time.Now()
		}()
	}
	wg.Wait()
	close(done)

	times := make([]time.Time, 0, n)
	for tt := range done {
		times = append(times, tt)
	}
	sort.Slice(times, func(i, j int) bool { return times[i].Before(times[j]) })

	for i := 1; i < len(times); i++ {
		d := times[i].Sub(times[i-1])
		if d < 10*time.Millisecond || d > 35*time.Millisecond {
			t.Errorf("interval[%d] = %v, want ≈ 20ms", i, d)
		}
	}
}

// TestDistributedQueueCap 验证 §5.1 的 capacity 语义：capacity=2 时
// 1 个立即放行 + 2 个排队放行 = 3 个成功，其余并发请求被拒绝
// （正好补上单机版「等待无上限」的短板）。
func TestDistributedQueueCap(t *testing.T) {
	rdb := newTestRDB(t)

	b := NewDistributedBucket(rdb, "lb:cap", 50, 2) // perRequest=20ms, capacity=2
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	const n = 20
	var (
		ok   int32
		wait sync.WaitGroup
	)
	for i := 0; i < n; i++ {
		wait.Add(1)
		go func() {
			defer wait.Done()
			if b.Take(ctx) == nil {
				atomic.AddInt32(&ok, 1)
			}
		}()
	}
	wait.Wait()

	if ok != 3 {
		t.Errorf("allowed = %d, want 3 (1 immediate + 2 queued, rest rejected)", ok)
	}
}

// TestDistributedCtxCancel 验证等待期间 ctx 取消返回 ctx.Err()，
// 调用方取消不再傻等。
func TestDistributedCtxCancel(t *testing.T) {
	rdb := newTestRDB(t)

	b := NewDistributedBucket(rdb, "lb:ctx", 50, 10)
	if err := b.Take(context.Background()); err != nil {
		t.Fatalf("first Take: %v", err)
	}

	// 第二个请求需等 ~20ms，用 1ms 超时的 ctx 应拿到 DeadlineExceeded
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()
	if err := b.Take(ctx); err != context.DeadlineExceeded {
		t.Fatalf("Take with cancelled ctx: got %v, want %v", err, context.DeadlineExceeded)
	}
}
