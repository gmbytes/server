package utils

import (
	"sort"
	"testing"
)

// -----------------------------------------------------------------------
// Keys
// -----------------------------------------------------------------------

func TestKeys_WhenNonEmpty_ExpectAllKeys(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	keys := Keys(m)
	if len(keys) != 3 {
		t.Errorf("Keys 期望 3 个键，得到 %d", len(keys))
	}
	sort.Strings(keys)
	for i, want := range []string{"a", "b", "c"} {
		if keys[i] != want {
			t.Errorf("Keys[%d] 期望 %q，得到 %q", i, want, keys[i])
		}
	}
}

func TestKeys_WhenEmpty_ExpectEmptySlice(t *testing.T) {
	m := map[int]string{}
	keys := Keys(m)
	if len(keys) != 0 {
		t.Errorf("空 map 的 Keys 期望空切片，得到 %d 个元素", len(keys))
	}
}

func TestKeys_WhenNil_ExpectEmptySlice(t *testing.T) {
	var m map[int]string
	keys := Keys(m)
	if len(keys) != 0 {
		t.Errorf("nil map 的 Keys 期望空切片，得到 %d 个元素", len(keys))
	}
}

func TestKeys_WhenSingleEntry_ExpectSingleKey(t *testing.T) {
	m := map[int]int{42: 100}
	keys := Keys(m)
	if len(keys) != 1 || keys[0] != 42 {
		t.Errorf("单条目 map 的 Keys 期望 [42]，得到 %v", keys)
	}
}

// -----------------------------------------------------------------------
// Values
// -----------------------------------------------------------------------

func TestValues_WhenNonEmpty_ExpectAllValues(t *testing.T) {
	m := map[string]int{"a": 1, "b": 2, "c": 3}
	values := Values(m)
	if len(values) != 3 {
		t.Errorf("Values 期望 3 个值，得到 %d", len(values))
	}
	sort.Ints(values)
	for i, want := range []int{1, 2, 3} {
		if values[i] != want {
			t.Errorf("Values[%d] 期望 %d，得到 %d", i, want, values[i])
		}
	}
}

func TestValues_WhenEmpty_ExpectEmptySlice(t *testing.T) {
	m := map[string]int{}
	values := Values(m)
	if len(values) != 0 {
		t.Errorf("空 map 的 Values 期望空切片，得到 %d 个元素", len(values))
	}
}

func TestValues_WhenNil_ExpectEmptySlice(t *testing.T) {
	var m map[string]int
	values := Values(m)
	if len(values) != 0 {
		t.Errorf("nil map 的 Values 期望空切片，得到 %d 个元素", len(values))
	}
}

func TestValues_WhenSingleEntry_ExpectSingleValue(t *testing.T) {
	m := map[string]float64{"pi": 3.14}
	values := Values(m)
	if len(values) != 1 || values[0] != 3.14 {
		t.Errorf("单条目 map 的 Values 期望 [3.14]，得到 %v", values)
	}
}

// -----------------------------------------------------------------------
// Keys 和 Values 长度一致性
// -----------------------------------------------------------------------

func TestKeysAndValues_WhenSameMap_ExpectSameLength(t *testing.T) {
	m := map[int]string{1: "one", 2: "two", 3: "three", 4: "four"}
	keys := Keys(m)
	values := Values(m)
	if len(keys) != len(values) {
		t.Errorf("Keys 和 Values 长度应相等：len(keys)=%d, len(values)=%d", len(keys), len(values))
	}
	if len(keys) != 4 {
		t.Errorf("期望 4 个元素，得到 %d", len(keys))
	}
}
