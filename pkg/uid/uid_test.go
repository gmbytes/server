package uid

import (
	"strconv"
	"sync"
	"testing"
)

// -----------------------------------------------------------------------
// Uid 类型方法
// -----------------------------------------------------------------------

func TestUid_ToString_WhenPositive_ExpectDecimalString(t *testing.T) {
	u := Uid(123456)
	got := u.ToString()
	if got != "123456" {
		t.Errorf("期望 \"123456\"，得到 %q", got)
	}
}

func TestUid_ToString_WhenZero_ExpectZeroString(t *testing.T) {
	u := Uid(0)
	if got := u.ToString(); got != "0" {
		t.Errorf("期望 \"0\"，得到 %q", got)
	}
}

func TestUid_ToInt64_WhenPositive_ExpectSameValue(t *testing.T) {
	u := Uid(987654321)
	if got := u.ToInt64(); got != 987654321 {
		t.Errorf("期望 987654321，得到 %d", got)
	}
}

func TestUid_ToInt64_WhenZero_ExpectZero(t *testing.T) {
	u := Uid(0)
	if got := u.ToInt64(); got != 0 {
		t.Errorf("期望 0，得到 %d", got)
	}
}

func TestUid_IsValid_WhenNonZero_ExpectTrue(t *testing.T) {
	u := Uid(1)
	if !u.IsValid() {
		t.Error("期望 IsValid()=true，对非零 Uid")
	}
}

func TestUid_IsValid_WhenZero_ExpectFalse(t *testing.T) {
	u := Uid(0)
	if u.IsValid() {
		t.Error("期望 IsValid()=false，对零值 Uid")
	}
}

// -----------------------------------------------------------------------
// FromString
// -----------------------------------------------------------------------

func TestFromString_WhenValidDecimal_ExpectCorrectUid(t *testing.T) {
	cases := []struct {
		input string
		want  Uid
	}{
		{"1", Uid(1)},
		{"999", Uid(999)},
		{"9223372036854775807", Uid(9223372036854775807)}, // math.MaxInt64
	}

	for _, c := range cases {
		got := FromString(c.input)
		if got != c.want {
			t.Errorf("FromString(%q)=%d，期望 %d", c.input, got, c.want)
		}
	}
}

func TestFromString_WhenEmpty_ExpectZero(t *testing.T) {
	got := FromString("")
	if got != 0 {
		t.Errorf("FromString(\"\") 期望 0，得到 %d", got)
	}
}

// -----------------------------------------------------------------------
// Uid 与 ToString/FromString 互转
// -----------------------------------------------------------------------

func TestUid_RoundTrip_ToString_FromString(t *testing.T) {
	original := Uid(1234567890)
	str := original.ToString()
	got := FromString(str)
	if got != original {
		t.Errorf("往返转换失败：期望 %d，得到 %d", original, got)
	}
}

// -----------------------------------------------------------------------
// Init / Gen
// -----------------------------------------------------------------------

func TestInit_WhenValidNodeID_ExpectNoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Init 不应 panic，但 panic 了：%v", r)
		}
	}()
	Init(0)
	Init(1)
	Init(maxNodeID)
}

func TestInit_WhenNodeIDOverflow_ExpectPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Init(maxNodeID+1) 期望 panic，但没有")
		}
	}()
	Init(maxNodeID + 1)
}

func TestInit_WhenNegativeNodeID_ExpectPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Init(-1) 期望 panic，但没有")
		}
	}()
	Init(-1)
}

func TestGen_WhenCalled_ExpectPositiveUID(t *testing.T) {
	Init(1)
	u := Gen()
	if u <= 0 {
		t.Errorf("Gen() 期望正数，得到 %d", u)
	}
}

func TestGen_WhenCalledMultipleTimes_ExpectUniqueIDs(t *testing.T) {
	Init(1)
	const n = 100
	seen := make(map[Uid]bool, n)
	for i := 0; i < n; i++ {
		u := Gen()
		if seen[u] {
			t.Errorf("Gen() 生成了重复的 ID: %d", u)
		}
		seen[u] = true
	}
}

func TestGen_WhenCalledConcurrently_ExpectUniqueIDs(t *testing.T) {
	Init(2)
	const goroutines = 10
	const perGoroutine = 50
	results := make(chan Uid, goroutines*perGoroutine)

	var wg sync.WaitGroup
	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < perGoroutine; j++ {
				results <- Gen()
			}
		}()
	}
	wg.Wait()
	close(results)

	seen := make(map[Uid]bool, goroutines*perGoroutine)
	for u := range results {
		if seen[u] {
			t.Errorf("并发 Gen() 生成了重复的 ID: %d", u)
		}
		seen[u] = true
	}
}

func TestGen_WhenCalled_ExpectIDFitsInt63(t *testing.T) {
	Init(1)
	u := Gen()
	// Gen 生成的 ID 应在 int63 范围内（最高位为符号位，不应为负）
	if u.ToInt64() < 0 {
		t.Errorf("Gen() 不应生成负数 ID，得到 %d", u)
	}
}

// -----------------------------------------------------------------------
// 零值常量
// -----------------------------------------------------------------------

func TestZero_ExpectInvalid(t *testing.T) {
	if Zero.IsValid() {
		t.Error("Zero 应为无效 Uid")
	}
	if Zero.ToInt64() != 0 {
		t.Error("Zero.ToInt64() 期望 0")
	}
	if Zero.ToString() != "0" {
		t.Error("Zero.ToString() 期望 \"0\"")
	}
}

// -----------------------------------------------------------------------
// 辅助：确认 maxIndex 常量值正确
// -----------------------------------------------------------------------

func TestConst_MaxIndex_ExpectCorrectBitMask(t *testing.T) {
	expected := int64(1<<20 - 1)
	if maxIndex != expected {
		t.Errorf("maxIndex 期望 %d，得到 %d", expected, maxIndex)
	}
}

// -----------------------------------------------------------------------
// VAR_ZONE_EntityId 初始值
// -----------------------------------------------------------------------

func TestVAR_ZONE_EntityId_ExpectOne(t *testing.T) {
	if VAR_ZONE_EntityId != Uid(1) {
		t.Errorf("VAR_ZONE_EntityId 期望 1，得到 %s", strconv.FormatInt(VAR_ZONE_EntityId.ToInt64(), 10))
	}
}
