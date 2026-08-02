package leakybucket

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// TestListBucketFullThenLazyDrain 验证 §5.2 的惰性补账：
// 容量 2 的桶连续入队 3 个，第 3 个被拒；等 60ms（应出掉 1 个，
// perRequest=50ms）后再次入队成功。
func TestListBucketFullThenLazyDrain(t *testing.T) {
	rdb := newTestRDB(t)

	b := NewListBucket(rdb, "lb:list", 20, 2) // perRequest=50ms, capacity=2
	ctx := context.Background()

	for i := 0; i < 2; i++ {
		ok, err := b.Enqueue(ctx, fmt.Sprintf("job-%d", i))
		if err != nil {
			t.Fatalf("Enqueue #%d: %v", i, err)
		}
		if !ok {
			t.Fatalf("Enqueue #%d should succeed: bucket has room", i)
		}
	}

	// 无流逝时间：桶仍满，第 3 个被拒绝
	ok, err := b.Enqueue(ctx, "job-2")
	if err != nil {
		t.Fatalf("Enqueue #3: %v", err)
	}
	if ok {
		t.Error("Enqueue #3 should be rejected: bucket full")
	}

	// 等 60ms：惰性补账应出掉 1 个（floor(60/50)=1），腾出容量
	time.Sleep(60 * time.Millisecond)
	ok, err = b.Enqueue(ctx, "job-3")
	if err != nil {
		t.Fatalf("Enqueue #4: %v", err)
	}
	if !ok {
		t.Error("Enqueue #4 should succeed after 60ms: 1 job lazily drained")
	}
}
