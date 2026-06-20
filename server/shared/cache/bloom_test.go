package cache

import (
	"fmt"
	"sync"
	"testing"
)

// 测试基础正确性
func TestBloomFilter_BasicOperations(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		addItems  []string // 添加入布隆过滤器的物品
		checkHit  []string // 检查应当判定为存在的物品
		checkMiss []string // 检查应当被判定为不存在或者误判为存在的物品
	}{
		{
			name:      "单个元素添加",
			addItems:  []string{"apple"},
			checkHit:  []string{"apple"},
			checkMiss: []string{"banana", "grape", "watermelon"},
		},
		{
			name:      "多个元素添加",
			addItems:  []string{"apple", "banana", "grape"},
			checkHit:  []string{"apple", "banana", "grape"},
			checkMiss: []string{"melon", "cherry", "pear"},
		},
		{
			name:      "空字符串",
			addItems:  []string{},
			checkHit:  []string{},
			checkMiss: []string{"notempty"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			bf := NewBloomFilter(1000, 0.01)

			for _, item := range tt.addItems {
				bf.Add(item)
			}

			for _, item := range tt.checkHit {
				if !bf.MayContain(item) {
					t.Errorf("MayContain(%q) = false, want true", item)
				}
			}

			for _, item := range tt.checkMiss {
				if bf.MayContain(item) {
					t.Logf("MayContain(%q) = true, want false", item)
				}
			}
		})
	}
}

// 测试错误参数
func TestBloomFilter_Params(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		n    int
		p    float64
	}{
		{"n=0 应使用默认值", 0, 0.01},
		{"n=-1 应使用默认值", -1, 0.01},
		{"p=0 应使用默认值", 1000, 0},
		{"p=1 应使用默认值", 1000, 1},
		{"p=-0.5 应使用默认值", 1000, -0.5},
		{"p=1.5 应使用默认值", 1000, 1.5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			bf := NewBloomFilter(tt.n, tt.p)
			bf.Add("test")
			if !bf.MayContain("test") {
				t.Error("使用错误参数后功能失效")
			}
		})
	}
}

// 误判率验证
func TestBloomFilter_FalsePositiveRate(t *testing.T) {
	t.Parallel()

	n := 100_000
	p := 0.001
	tolerance := 0.2

	bf := NewBloomFilter(n, p)

	for i := 0; i < n; i++ {
		bf.Add(fmt.Sprintf("item-%d", i))
	}

	testCount := 100_000
	falsePositives := 0
	for i := 0; i < testCount; i++ {
		if bf.MayContain(fmt.Sprintf("check-%d", i)) {
			falsePositives++
		}
	}

	actualRate := float64(falsePositives) / float64(testCount)
	upperBound := p * (1 + tolerance)

	t.Logf("理论误判率： %.4f，实际误判率： %.4f (%d/%d)",
		p, actualRate, falsePositives, testCount)

	if actualRate > upperBound {
		t.Errorf("误判率 %.4f 超过上界 %.4f (理论值 %.4f + 20%%)",
			actualRate, upperBound, p)
	}
}

// 竞态测试
func TestBloomFilter_ConcurrentSafety(t *testing.T) {
	t.Parallel()

	bf := NewBloomFilter(100_000, 0.01)
	goroutines := 100
	itemsPerGoRoutine := 1000

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for g := 0; g < goroutines; g++ {
		go func() {
			defer wg.Done()
			for i := 0; i < itemsPerGoRoutine; i++ {
				bf.Add(fmt.Sprintf("g%d-item%d", g, i))
			}
		}()
	}
	wg.Wait()

	for g := 0; g < goroutines; g++ {
		for i := 0; i < itemsPerGoRoutine; i++ {
			key := fmt.Sprintf("g%d-item%d", g, i)
			if !bf.MayContain(key) {
				t.Errorf("Concurrent Test MayContain(%q) = false, want true", key)
			}
		}
	}
}

// 测试 RWMutex 不会死锁或 Panic
func TestBloomFilter_ConcurrentReadWrite(t *testing.T) {
	t.Parallel()

	bf := NewBloomFilter(100_000, 0.01)
	var wg sync.WaitGroup

	for w := 0; w < 50; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				bf.Add(fmt.Sprintf("w%d-item%d", w, i))
			}
		}()
	}

	for r := 0; r < 50; r++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 200; i++ {
				bf.MayContain(fmt.Sprintf("r%d-item%d", r, i))
			}
		}()
	}

	wg.Wait()
}
