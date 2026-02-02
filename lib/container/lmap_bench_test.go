package container

import (
	"sync"
	"testing"
)

// Benchmark: LMap vs 普通 map

func BenchmarkLMap_Set(b *testing.B) {
	m := NewLMap[int, int]()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
}

func BenchmarkMap_Set(b *testing.B) {
	m := make(map[int]int)
	var mu sync.RWMutex
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.Lock()
		m[i] = i
		mu.Unlock()
	}
}

func BenchmarkLMap_Get(b *testing.B) {
	m := NewLMap[int, int]()
	for i := 0; i < 10000; i++ {
		m.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(i % 10000)
	}
}

func BenchmarkMap_Get(b *testing.B) {
	m := make(map[int]int)
	for i := 0; i < 10000; i++ {
		m[i] = i
	}
	var mu sync.RWMutex
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		_ = m[i%10000]
		mu.RUnlock()
	}
}

func BenchmarkLMap_Delete(b *testing.B) {
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		m := NewLMap[int, int]()
		for j := 0; j < 1000; j++ {
			m.Set(j, j)
		}
		b.StartTimer()
		m.Delete(500)
		b.StopTimer()
	}
}

func BenchmarkMap_Delete(b *testing.B) {
	var mu sync.RWMutex
	b.StopTimer()
	for i := 0; i < b.N; i++ {
		m := make(map[int]int)
		for j := 0; j < 1000; j++ {
			m[j] = j
		}
		b.StartTimer()
		mu.Lock()
		delete(m, 500)
		mu.Unlock()
		b.StopTimer()
	}
}

func BenchmarkLMap_ForEach(b *testing.B) {
	m := NewLMap[int, int]()
	for i := 0; i < 1000; i++ {
		m.Set(i, i)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.ForEach(func(v int) {
			_ = v
		})
	}
}

func BenchmarkMap_ForEach(b *testing.B) {
	m := make(map[int]int)
	for i := 0; i < 1000; i++ {
		m[i] = i
	}
	var mu sync.RWMutex
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		mu.RLock()
		for k, v := range m {
			_ = k + v
		}
		mu.RUnlock()
	}
}

// 模拟战斗系统的实际使用场景
func BenchmarkLMap_CombatScenario(b *testing.B) {
	m := NewLMap[int, *EffectData]()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 添加效果
		for j := 0; j < 10; j++ {
			m.Set(i*10+j, &EffectData{
				Id:       i*10 + j,
				Damage:   100,
				Duration: 5000,
			})
		}

		// 更新效果
		m.ForEach(func(effect *EffectData) {
			effect.Duration -= 50
		})

		// 删除过期效果
		// 注意：需要先收集所有过期效果的 key
		toDelete := make([]int, 0)
		for _, entry := range m.Entries() {
			if entry.Value.Duration <= 0 {
				toDelete = append(toDelete, entry.Key)
			}
		}
		for _, id := range toDelete {
			m.Delete(id)
		}
	}
}

func BenchmarkMap_CombatScenario(b *testing.B) {
	m := make(map[int]*EffectData)
	var mu sync.RWMutex

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// 添加效果
		mu.Lock()
		for j := 0; j < 10; j++ {
			m[i*10+j] = &EffectData{
				Id:       i*10 + j,
				Damage:   100,
				Duration: 5000,
			}
		}
		mu.Unlock()

		// 更新效果
		mu.Lock()
		for _, effect := range m {
			effect.Duration -= 50
		}
		mu.Unlock()

		// 删除过期效果
		mu.Lock()
		for id, effect := range m {
			if effect.Duration <= 0 {
				delete(m, id)
			}
		}
		mu.Unlock()
	}
}

type EffectData struct {
	Id       int
	Damage   int
	Duration int
}

// 内存分配测试
func BenchmarkLMap_Allocations(b *testing.B) {
	b.ReportAllocs()
	m := NewLMap[int, int]()
	for i := 0; i < b.N; i++ {
		m.Set(i, i)
	}
}

func BenchmarkMap_Allocations(b *testing.B) {
	b.ReportAllocs()
	m := make(map[int]int)
	for i := 0; i < b.N; i++ {
		m[i] = i
	}
}
