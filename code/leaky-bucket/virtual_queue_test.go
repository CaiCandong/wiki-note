package leakybucket

import (
	"fmt"
	"sort"
	"testing"
	"time"
)

// mockMQ 用内存 channel 模拟持久化队列（生产换 RabbitMQ）：
// 多个 pipeline 共享同一个 channel 时，多接收者竞争消费 = MQ 语义。
func mockMQ(cap int) (chan string, func() (string, bool)) {
	mq := make(chan string, cap)
	fetch := func() (string, bool) {
		select {
		case msg := <-mq:
			return msg, true
		default:
			return "", false
		}
	}
	return mq, fetch
}

// collectSents 从全部 pipeline 的发送出口收满 total 条消息并记录时间戳。
func collectSents(t *testing.T, pipes []*VirtualQueuePipeline, total int) []time.Time {
	t.Helper()
	times := make([]time.Time, 0, total)
	deadline := time.After(10 * time.Second)
	for len(times) < total {
		for _, p := range pipes {
			select {
			case <-p.Sent():
				times = append(times, time.Now())
			default:
			}
		}
		select {
		case <-deadline:
			t.Fatalf("timeout: got %d/%d messages", len(times), total)
		default:
		}
	}
	return times
}

// TestVirtualQueueConstantRate 虚拟队列版：worker 在时间轴上排队，
// 无内存队列；20 条消息在 mock MQ 里，单 worker 按放行间隔
// （50ms）逐条拉取发送——全部送达、间隔恒定、积压匀速清空。
func TestVirtualQueueConstantRate(t *testing.T) {
	const (
		rate  = 20 // perRequest = 50ms
		total = 20
	)
	mq, fetch := mockMQ(100)
	for i := 0; i < total; i++ {
		mq <- fmt.Sprintf("msg-%d", i)
	}

	p := NewVirtualQueuePipeline(blockingAdapter{NewBlockingBucket(rate)}, 1, 0, fetch)
	times := collectSents(t, []*VirtualQueuePipeline{p}, total)
	p.Stop()

	for i := 1; i < len(times); i++ {
		d := times[i].Sub(times[i-1])
		if d < 25*time.Millisecond || d > 65*time.Millisecond {
			t.Errorf("interval[%d] = %v, want ≈ 50ms", i, d)
		}
	}
}

// TestVirtualQueueHighRateWithWorkers 虚拟队列版高 QPS × 长耗时：
// rate=500/s、单条 50ms → workers ≈ 25（Little's Law），
// 500 条 ≈ 1s 全部送达。
func TestVirtualQueueHighRateWithWorkers(t *testing.T) {
	const (
		rate      = 500
		total     = 500
		workers   = 25
		sendDelay = 50 * time.Millisecond
	)
	mq, fetch := mockMQ(2000)
	for i := 0; i < total; i++ {
		mq <- fmt.Sprintf("msg-%d", i)
	}

	p := NewVirtualQueuePipeline(blockingAdapter{NewBlockingBucket(rate)}, workers, sendDelay, fetch)
	times := collectSents(t, []*VirtualQueuePipeline{p}, total)
	p.Stop()

	span := times[len(times)-1].Sub(times[0])
	if span > 2500*time.Millisecond {
		t.Errorf("500 msgs in %v, want ≈ 1s (25 workers × 50ms should sustain 500/s)", span)
	}
}

// TestVirtualQueueMultiPodSharedBucket 多 pod 生产形态（用户场景）：
// 单 pod worker 上限不够 → 多 pod 扩容，漏桶状态必须 Redis 共享——
// 3 个 pod 的 15 个 worker 全部调同一个 Redis key 的 Take（时间轴
// 排队），共享 mock MQ 竞争消费，全局总放行速率 = 300/s，
// 600 条 ≈ 2s 全部送达。
func TestVirtualQueueMultiPodSharedBucket(t *testing.T) {
	const (
		rate      = 300
		total     = 600
		pods      = 3
		workers   = 5 // 每 pod 5 worker，总在途 15 ≫ 300×10ms=3
		sendDelay = 10 * time.Millisecond
	)
	rdb := newTestRDB(t)

	shared := NewDistributedBucket(rdb, "lb:vq-multipod", rate, 1000)
	mq, fetch := mockMQ(2000)
	for i := 0; i < total; i++ {
		mq <- fmt.Sprintf("msg-%d", i)
	}

	pipelines := make([]*VirtualQueuePipeline, pods)
	for i := range pipelines {
		pipelines[i] = NewVirtualQueuePipeline(distAdapter{shared}, workers, sendDelay, fetch)
		defer pipelines[i].Stop()
	}

	times := collectSents(t, pipelines, total)
	for _, p := range pipelines {
		p.Stop()
	}

	span := times[len(times)-1].Sub(times[0])
	want := time.Duration(total) * time.Second / time.Duration(rate) // 600/300 = 2s
	if span < want*80/100 {
		t.Errorf("shared bucket: 600 msgs in %v, want ≈ %v (global rate must stay %d/s)", span, want, rate)
	}

	// 平稳性：放行间隔应围绕 perRequest（≈3.3ms）分布——无大空窗
	//（间隔翻倍堆积）、无突发堆叠（多个请求同一时刻放行）
	perReq := time.Second / time.Duration(rate)
	intervals := make([]time.Duration, 0, len(times)-1)
	for i := 1; i < len(times); i++ {
		intervals = append(intervals, times[i].Sub(times[i-1]))
	}
	sort.Slice(intervals, func(i, j int) bool { return intervals[i] < intervals[j] })
	median := intervals[len(intervals)/2]
	maxI := intervals[len(intervals)-1]
	smooth := 0
	for _, d := range intervals {
		if d >= perReq/2 && d <= perReq*2 {
			smooth++
		}
	}

	// 窗口速率：100ms 窗口内放行数应稳定在 rate/10 ≈ 30 附近
	const win = 100 * time.Millisecond
	windowCounts := make([]int, 0)
	start, cur := times[0], 0
	for _, tt := range times {
		if tt.Sub(start) >= win {
			windowCounts = append(windowCounts, cur)
			start, cur = tt, 0
		}
		cur++
	}
	windowCounts = append(windowCounts, cur)
	wmin, wmax := windowCounts[0], windowCounts[0]
	for _, c := range windowCounts[1:] {
		if c < wmin {
			wmin = c
		}
		if c > wmax {
			wmax = c
		}
	}
	t.Logf("smoothness: median=%v max=%v smooth=%.0f%% | windows(100ms): min=%d max=%d",
		median, maxI, float64(smooth)/float64(len(intervals))*100, wmin, wmax)

	if median < perReq*50/100 || median > perReq*180/100 {
		t.Errorf("median interval = %v, want ≈ %v (±50%%..+80%%)", median, perReq)
	}
	if maxI > perReq*4 {
		t.Errorf("max interval = %v, want < %v (no long gaps)", maxI, perReq*4)
	}
	if float64(smooth)/float64(len(intervals)) < 0.75 {
		t.Errorf("smooth ratio = %.0f%%, want ≥ 75%% (intervals within [perReq/2, 2×perReq])",
			float64(smooth)/float64(len(intervals))*100)
	}
	wantPerWin := int(win) * rate / int(time.Second) // 30
	if wmin < wantPerWin*60/100 || wmax > wantPerWin*150/100 {
		t.Errorf("window counts [%d, %d], want ≈ %d ±50%% (rate %d/s)", wmin, wmax, wantPerWin, rate)
	}
}

// TestVirtualQueueMultiPodLocalBuckets 对照组（错误做法）：每 pod
// 各用单机漏桶——总放行速率 = 3 × rate = 900/s，600 条 ≈ 0.67s，
// 实测证明「各 pod 独立限流必然超限」。
func TestVirtualQueueMultiPodLocalBuckets(t *testing.T) {
	const (
		rate      = 300
		total     = 600
		pods      = 3
		workers   = 5
		sendDelay = 10 * time.Millisecond
	)
	mq, fetch := mockMQ(2000)
	for i := 0; i < total; i++ {
		mq <- fmt.Sprintf("msg-%d", i)
	}

	pipelines := make([]*VirtualQueuePipeline, pods)
	for i := range pipelines {
		pipelines[i] = NewVirtualQueuePipeline(blockingAdapter{NewBlockingBucket(rate)}, workers, sendDelay, fetch)
		defer pipelines[i].Stop()
	}

	times := collectSents(t, pipelines, total)
	for _, p := range pipelines {
		p.Stop()
	}

	span := times[len(times)-1].Sub(times[0])
	want := time.Duration(total) * time.Second / time.Duration(rate) // 若共享应为 2s
	if span > want*80/100 {
		t.Errorf("per-pod buckets: 600 msgs in %v, want ≪ %v (3×rate over-limit = 900/s)", span, want)
	}
}
