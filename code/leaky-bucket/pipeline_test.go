package leakybucket

import (
	"fmt"
	"testing"
	"time"
)

// TestPushPipelineNoLossAndConstantRate 验证笔记 §5.4 推送场景组合：
// 20 条消息瞬间灌入（远超瞬时发送能力），最终全部送达（不丢）、
// 发送间隔恒定 ≈ 50ms（rate=20/s）、积压被匀速清空。
// 单 worker + sendDelay=0：放行间隔 = 发送间隔。
func TestPushPipelineNoLossAndConstantRate(t *testing.T) {
	const (
		rate   = 20 // perRequest = 50ms
		total  = 20
		queueC = 100
	)
	p := NewPushPipeline(rate, queueC, 1, 0)

	// 瞬间灌入：消息一条不落（阻塞提交，不拒绝）
	for i := 0; i < total; i++ {
		p.Submit(fmt.Sprintf("msg-%d", i))
	}

	// 匀速送达：全部收到（不丢）
	times := make([]time.Time, 0, total)
	for i := 0; i < total; i++ {
		select {
		case <-p.Sent():
			times = append(times, time.Now())
		case <-time.After(3 * time.Second):
			t.Fatalf("timeout: got %d/%d messages", i, total)
		}
	}
	p.Stop()

	// 发送间隔 ≈ 50ms（rate=20/s）
	for i := 1; i < len(times); i++ {
		d := times[i].Sub(times[i-1])
		if d < 25*time.Millisecond || d > 65*time.Millisecond {
			t.Errorf("interval[%d] = %v, want ≈ 50ms", i, d)
		}
	}
	// 总时长 ≈ (total-1) * 50ms：积压被匀速清空
	want := time.Duration(total-1) * 50 * time.Millisecond
	span := times[len(times)-1].Sub(times[0])
	if span < want-100*time.Millisecond || span > want+100*time.Millisecond {
		t.Errorf("total span = %v, want ≈ %v", span, want)
	}
}

// TestPushPipelineHighRateWithWorkers 验证厂商硬场景：QPS 上限 500、
// 单条发送耗时 50ms。放行速率 500/s（每 2ms 一条），但发送必须在
// worker pool 里并发做：workers ≈ 500 × 0.05 = 25（Little's Law）。
// 500 条瞬间灌入 → 全部送达（不丢），约 1 秒清空（吞吐 ≈ 500/s）。
func TestPushPipelineHighRateWithWorkers(t *testing.T) {
	const (
		rate      = 500
		total     = 500
		queueCap  = 2000
		workers   = 25
		sendDelay = 50 * time.Millisecond
	)
	p := NewPushPipeline(rate, queueCap, workers, sendDelay)

	for i := 0; i < total; i++ {
		p.Submit(fmt.Sprintf("msg-%d", i))
	}

	times := make([]time.Time, 0, total)
	for i := 0; i < total; i++ {
		select {
		case <-p.Sent():
			times = append(times, time.Now())
		case <-time.After(5 * time.Second):
			t.Fatalf("timeout: got %d/%d messages", i, total)
		}
	}
	p.Stop()

	span := times[len(times)-1].Sub(times[0])
	// 期望 ≈1s；容差放宽到 2.5s 以内（吞吐 > 200/s，远高于单 worker
	// 同步发送的 20/s），验证并发 worker 使高 QPS × 长耗时场景成立
	if span > 2500*time.Millisecond {
		t.Errorf("500 msgs in %v, want ≈ 1s (25 workers × 50ms should sustain 500/s)", span)
	}
}

// TestPushPipelineWorkersStarved 对照实验：worker 不足时吞吐被
// workers ÷ sendDelay 锁死（5 个 worker × 50ms = 100/s），积压增长，
// 漏桶放行速率形同虚设——验证 Little's Law 的约束方向。
func TestPushPipelineWorkersStarved(t *testing.T) {
	const (
		rate      = 500
		total     = 200
		queueCap  = 2000
		workers   = 5
		sendDelay = 50 * time.Millisecond
	)
	p := NewPushPipeline(rate, queueCap, workers, sendDelay)

	for i := 0; i < total; i++ {
		p.Submit(fmt.Sprintf("msg-%d", i))
	}

	times := make([]time.Time, 0, total)
	for i := 0; i < total; i++ {
		select {
		case <-p.Sent():
			times = append(times, time.Now())
		case <-time.After(10 * time.Second):
			t.Fatalf("timeout: got %d/%d messages", i, total)
		}
	}
	p.Stop()

	span := times[len(times)-1].Sub(times[0])
	// 5 workers × 20/s = 100/s → 200 条 ≈ 2s；断言显著慢于 25 worker 场景
	if span < 1200*time.Millisecond {
		t.Errorf("200 msgs in %v, want ≈ 2s (starved: only 5 workers × 20/s)", span)
	}
}
