package container

import (
	"testing"
)

func TestLMap_Basic(t *testing.T) {
	m := NewLMap[string, int]()

	// Test Set and Get
	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	if val, ok := m.Get("a"); !ok || val != 1 {
		t.Errorf("Expected a=1, got %v", val)
	}

	if val, ok := m.Get("b"); !ok || val != 2 {
		t.Errorf("Expected b=2, got %v", val)
	}

	// Test Len
	if m.Len() != 3 {
		t.Errorf("Expected len=3, got %d", m.Len())
	}

	// Test Has
	if !m.Has("a") {
		t.Error("Expected Has(a)=true")
	}

	if m.Has("d") {
		t.Error("Expected Has(d)=false")
	}
}

func TestLMap_Order(t *testing.T) {
	m := NewLMap[int, string]()

	// 按顺序插入
	m.Set(3, "three")
	m.Set(1, "one")
	m.Set(4, "four")
	m.Set(2, "two")

	// 验证遍历顺序（使用 Keys）
	keys := m.Keys()
	expected := []int{3, 1, 4, 2}

	if len(keys) != len(expected) {
		t.Errorf("Expected %d keys, got %d", len(expected), len(keys))
	}

	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("Expected key=%d at index %d, got %d", expected[i], i, key)
		}
	}
}

func TestLMap_Delete(t *testing.T) {
	m := NewLMap[string, int]()

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)
	m.Set("d", 4)

	// 删除中间元素
	if !m.Delete("b") {
		t.Error("Expected Delete(b)=true")
	}

	if m.Len() != 3 {
		t.Errorf("Expected len=3 after delete, got %d", m.Len())
	}

	// 验证删除后剩余的键（注意：交换删除法会改变最后元素位置）
	// 删除 b 后，d 会移到 b 的位置，所以顺序变为 a, d, c
	if !m.Has("a") || !m.Has("c") || !m.Has("d") {
		t.Error("Expected map to have keys a, c, d")
	}

	if m.Has("b") {
		t.Error("Expected b to be deleted")
	}

	// 删除头节点
	m.Delete("a")
	if m.Len() != 2 {
		t.Errorf("Expected len=2, got %d", m.Len())
	}

	// 删除尾节点
	m.Delete("d")
	if m.Len() != 1 {
		t.Errorf("Expected len=1, got %d", m.Len())
	}

	// 验证剩余元素
	if val, ok := m.Get("c"); !ok || val != 3 {
		t.Errorf("Expected c=3, got %v", val)
	}
}

func TestLMap_Update(t *testing.T) {
	m := NewLMap[string, int]()

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	// 更新已存在的键
	m.Set("b", 20)

	if val, ok := m.Get("b"); !ok || val != 20 {
		t.Errorf("Expected b=20, got %v", val)
	}

	// 验证长度不变
	if m.Len() != 3 {
		t.Errorf("Expected len=3, got %d", m.Len())
	}

	// 验证顺序不变（使用 Keys）
	keys := m.Keys()
	expected := []string{"a", "b", "c"}

	if len(keys) != len(expected) {
		t.Errorf("Expected %d keys, got %d", len(expected), len(keys))
	}

	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("Expected key=%s at index %d, got %s", expected[i], i, key)
		}
	}
}

func TestLMap_ForEachBreakable(t *testing.T) {
	m := NewLMap[int, string]()

	m.Set(1, "one")
	m.Set(2, "two")
	m.Set(3, "three")
	m.Set(4, "four")

	count := 0
	m.ForEachBreakable(func(value string) bool {
		count++
		// 遍历前3个元素后中断
		return count < 3
	})

	if count != 3 {
		t.Errorf("Expected 3 iterations, got %d", count)
	}
}

func TestLMap_ForEachReverse(t *testing.T) {
	m := NewLMap[int, string]()

	m.Set(1, "one")
	m.Set(2, "two")
	m.Set(3, "three")

	expected := []string{"three", "two", "one"}
	index := 0

	m.ForEachReverse(func(value string) {
		if value != expected[index] {
			t.Errorf("Expected value=%s at index %d, got %s", expected[index], index, value)
		}
		index++
	})
}

func TestLMap_Keys(t *testing.T) {
	m := NewLMap[string, int]()

	m.Set("c", 3)
	m.Set("a", 1)
	m.Set("b", 2)

	keys := m.Keys()
	expected := []string{"c", "a", "b"}

	if len(keys) != len(expected) {
		t.Errorf("Expected %d keys, got %d", len(expected), len(keys))
	}

	for i, key := range keys {
		if key != expected[i] {
			t.Errorf("Expected key=%s at index %d, got %s", expected[i], i, key)
		}
	}
}

func TestLMap_Values(t *testing.T) {
	m := NewLMap[string, int]()

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	values := m.Values()
	expected := []int{1, 2, 3}

	if len(values) != len(expected) {
		t.Errorf("Expected %d values, got %d", len(expected), len(values))
	}

	for i, val := range values {
		if val != expected[i] {
			t.Errorf("Expected value=%d at index %d, got %d", expected[i], i, val)
		}
	}
}

func TestLMap_Clear(t *testing.T) {
	m := NewLMap[string, int]()

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	m.Clear()

	if m.Len() != 0 {
		t.Errorf("Expected len=0 after clear, got %d", m.Len())
	}

	if _, ok := m.Get("a"); ok {
		t.Error("Expected Get(a) to fail after clear")
	}
}

func TestLMap_Filter(t *testing.T) {
	m := NewLMap[int, string]()

	m.Set(1, "one")
	m.Set(2, "two")
	m.Set(3, "three")
	m.Set(4, "four")

	// 过滤出包含 "o" 的值
	filtered := m.Filter(func(value string) bool {
		return len(value) == 3 // "two" 和 "one" 长度为3，但 "one" 的key是1
	})

	if filtered.Len() != 2 {
		t.Errorf("Expected filtered len=2, got %d", filtered.Len())
	}

	// 验证过滤结果包含长度为3的值
	if !filtered.Has(1) || !filtered.Has(2) {
		t.Error("Expected filtered map to have keys 1 and 2")
	}
}

func TestLMap_Clone(t *testing.T) {
	m := NewLMap[string, int]()

	m.Set("a", 1)
	m.Set("b", 2)
	m.Set("c", 3)

	cloned := m.Clone()

	if cloned.Len() != m.Len() {
		t.Errorf("Expected cloned len=%d, got %d", m.Len(), cloned.Len())
	}

	// 修改原Map不应影响克隆
	m.Set("d", 4)

	if cloned.Len() != 3 {
		t.Errorf("Expected cloned len=3 after modifying original, got %d", cloned.Len())
	}
}
