package utils

import (
	"testing"
)

// -----------------------------------------------------------------------
// ValidateName
// -----------------------------------------------------------------------

func TestValidateName_WhenASCIILetters_ExpectValidAndCorrectLength(t *testing.T) {
	length, ok := ValidateName("abc")
	if !ok {
		t.Error("期望 ValidateName(\"abc\") 返回 true")
	}
	if length != 3 {
		t.Errorf("期望长度 3，得到 %d", length)
	}
}

func TestValidateName_WhenMixedCaseLetters_ExpectValid(t *testing.T) {
	length, ok := ValidateName("AbcDEF")
	if !ok {
		t.Error("期望大小写混合字母通过验证")
	}
	if length != 6 {
		t.Errorf("期望长度 6，得到 %d", length)
	}
}

func TestValidateName_WhenUnderscorePrefix_ExpectValid(t *testing.T) {
	length, ok := ValidateName("_abc")
	if !ok {
		t.Error("期望下划线开头的名字通过验证")
	}
	if length != 4 {
		t.Errorf("期望长度 4，得到 %d", length)
	}
}

func TestValidateName_WhenLettersAndDigits_ExpectValid(t *testing.T) {
	// 数字不能出现在第一个字符
	length, ok := ValidateName("abc123")
	if !ok {
		t.Error("期望字母后跟数字通过验证")
	}
	if length != 6 {
		t.Errorf("期望长度 6，得到 %d", length)
	}
}

func TestValidateName_WhenDigitAsFirstChar_ExpectInvalid(t *testing.T) {
	_, ok := ValidateName("1abc")
	if ok {
		t.Error("期望数字开头的名字验证失败")
	}
}

func TestValidateName_WhenChineseChars_ExpectValidAndDoubleLength(t *testing.T) {
	// 单个汉字占 2 个长度单位
	length, ok := ValidateName("你好")
	if !ok {
		t.Error("期望汉字通过验证")
	}
	if length != 4 {
		t.Errorf("两个汉字期望长度 4，得到 %d", length)
	}
}

func TestValidateName_WhenMixedChineseAndASCII_ExpectValidMixedLength(t *testing.T) {
	// "a你" = 1 + 2 = 3
	length, ok := ValidateName("a你")
	if !ok {
		t.Error("期望中英混合名字通过验证")
	}
	if length != 3 {
		t.Errorf("\"a你\" 期望长度 3，得到 %d", length)
	}
}

func TestValidateName_WhenSpecialChars_ExpectInvalid(t *testing.T) {
	cases := []string{"a@b", "a b", "a-b", "a.b"}
	for _, c := range cases {
		_, ok := ValidateName(c)
		if ok {
			t.Errorf("期望 %q 验证失败，但通过了", c)
		}
	}
}

func TestValidateName_WhenEmpty_ExpectValidWithZeroLength(t *testing.T) {
	length, ok := ValidateName("")
	if !ok {
		t.Error("空字符串应通过验证（空名字）")
	}
	if length != 0 {
		t.Errorf("空字符串期望长度 0，得到 %d", length)
	}
}

// -----------------------------------------------------------------------
// StrLen
// -----------------------------------------------------------------------

func TestStrLen_WhenASCII_ExpectOnePerChar(t *testing.T) {
	if got := StrLen("hello"); got != 5 {
		t.Errorf("StrLen(\"hello\") 期望 5，得到 %d", got)
	}
}

func TestStrLen_WhenChinese_ExpectTwoPerChar(t *testing.T) {
	// 汉字 Unicode > 256，每个占 2
	if got := StrLen("你好"); got != 4 {
		t.Errorf("StrLen(\"你好\") 期望 4，得到 %d", got)
	}
}

func TestStrLen_WhenEmpty_ExpectZero(t *testing.T) {
	if got := StrLen(""); got != 0 {
		t.Errorf("StrLen(\"\") 期望 0，得到 %d", got)
	}
}

func TestStrLen_WhenMixed_ExpectCorrectSum(t *testing.T) {
	// "a你b" = 1 + 2 + 1 = 4
	if got := StrLen("a你b"); got != 4 {
		t.Errorf("StrLen(\"a你b\") 期望 4，得到 %d", got)
	}
}

// -----------------------------------------------------------------------
// Atoi / Atoi16 / Atoi32 / Atoi64
// -----------------------------------------------------------------------

func TestAtoi_WhenValidInt_ExpectCorrectValue(t *testing.T) {
	if got := Atoi("42"); got != 42 {
		t.Errorf("Atoi(\"42\") 期望 42，得到 %d", got)
	}
}

func TestAtoi_WhenNegativeInt_ExpectCorrectValue(t *testing.T) {
	if got := Atoi("-100"); got != -100 {
		t.Errorf("Atoi(\"-100\") 期望 -100，得到 %d", got)
	}
}

func TestAtoi_WhenTrimSpace_ExpectCorrectValue(t *testing.T) {
	if got := Atoi("  7  "); got != 7 {
		t.Errorf("Atoi(\"  7  \") 期望 7，得到 %d", got)
	}
}

func TestAtoi_WhenInvalid_ExpectPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("期望无效字符串触发 panic")
		}
	}()
	Atoi("not_a_number")
}

func TestAtoi16_WhenValid_ExpectCorrectValue(t *testing.T) {
	if got := Atoi16("100"); got != 100 {
		t.Errorf("Atoi16(\"100\") 期望 100，得到 %d", got)
	}
}

func TestAtoi32_WhenValid_ExpectCorrectValue(t *testing.T) {
	if got := Atoi32("200"); got != 200 {
		t.Errorf("Atoi32(\"200\") 期望 200，得到 %d", got)
	}
}

func TestAtoi64_WhenLargeNumber_ExpectCorrectValue(t *testing.T) {
	if got := Atoi64("9223372036854775807"); got != 9223372036854775807 {
		t.Errorf("Atoi64 大数转换失败，得到 %d", got)
	}
}

func TestAtoi64_WhenNegative_ExpectCorrectValue(t *testing.T) {
	if got := Atoi64("-1"); got != -1 {
		t.Errorf("Atoi64(\"-1\") 期望 -1，得到 %d", got)
	}
}

func TestAtoi64_WhenInvalid_ExpectPanic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("期望无效字符串触发 panic")
		}
	}()
	Atoi64("abc")
}

// -----------------------------------------------------------------------
// Atou16 / Atou32 / Atou64
// -----------------------------------------------------------------------

func TestAtou16_WhenValid_ExpectCorrectValue(t *testing.T) {
	if got := Atou16("65535"); got != 65535 {
		t.Errorf("Atou16(\"65535\") 期望 65535，得到 %d", got)
	}
}

func TestAtou32_WhenValid_ExpectCorrectValue(t *testing.T) {
	if got := Atou32("4294967295"); got != 4294967295 {
		t.Errorf("Atou32(\"4294967295\") 期望 4294967295，得到 %d", got)
	}
}

func TestAtou64_WhenValid_ExpectCorrectValue(t *testing.T) {
	if got := Atou64("123456789"); got != 123456789 {
		t.Errorf("Atou64(\"123456789\") 期望 123456789，得到 %d", got)
	}
}

// -----------------------------------------------------------------------
// Itoa / ItoaInt
// -----------------------------------------------------------------------

func TestItoa_WhenPositive_ExpectDecimalString(t *testing.T) {
	if got := Itoa(42); got != "42" {
		t.Errorf("Itoa(42) 期望 \"42\"，得到 %q", got)
	}
}

func TestItoa_WhenNegative_ExpectNegativeString(t *testing.T) {
	if got := Itoa(-99); got != "-99" {
		t.Errorf("Itoa(-99) 期望 \"-99\"，得到 %q", got)
	}
}

func TestItoaInt_WhenZero_ExpectZeroString(t *testing.T) {
	if got := ItoaInt(0); got != "0" {
		t.Errorf("ItoaInt(0) 期望 \"0\"，得到 %q", got)
	}
}

// -----------------------------------------------------------------------
// MaxInt64 / MinInt64
// -----------------------------------------------------------------------

func TestMaxInt64_WhenFirstGreater_ExpectFirst(t *testing.T) {
	if got := MaxInt64(10, 5); got != 10 {
		t.Errorf("MaxInt64(10, 5) 期望 10，得到 %d", got)
	}
}

func TestMaxInt64_WhenSecondGreater_ExpectSecond(t *testing.T) {
	if got := MaxInt64(3, 8); got != 8 {
		t.Errorf("MaxInt64(3, 8) 期望 8，得到 %d", got)
	}
}

func TestMaxInt64_WhenEqual_ExpectSameValue(t *testing.T) {
	if got := MaxInt64(5, 5); got != 5 {
		t.Errorf("MaxInt64(5, 5) 期望 5，得到 %d", got)
	}
}

func TestMaxInt64_WhenNegatives_ExpectLargerNegative(t *testing.T) {
	if got := MaxInt64(-3, -7); got != -3 {
		t.Errorf("MaxInt64(-3, -7) 期望 -3，得到 %d", got)
	}
}

func TestMinInt64_WhenFirstSmaller_ExpectFirst(t *testing.T) {
	if got := MinInt64(2, 9); got != 2 {
		t.Errorf("MinInt64(2, 9) 期望 2，得到 %d", got)
	}
}

func TestMinInt64_WhenSecondSmaller_ExpectSecond(t *testing.T) {
	if got := MinInt64(10, 4); got != 4 {
		t.Errorf("MinInt64(10, 4) 期望 4，得到 %d", got)
	}
}

func TestMinInt64_WhenEqual_ExpectSameValue(t *testing.T) {
	if got := MinInt64(7, 7); got != 7 {
		t.Errorf("MinInt64(7, 7) 期望 7，得到 %d", got)
	}
}

// -----------------------------------------------------------------------
// MaxInt / MinInt
// -----------------------------------------------------------------------

func TestMaxInt_WhenFirstGreater_ExpectFirst(t *testing.T) {
	if got := MaxInt(100, 50); got != 100 {
		t.Errorf("MaxInt(100, 50) 期望 100，得到 %d", got)
	}
}

func TestMaxInt_WhenSecondGreater_ExpectSecond(t *testing.T) {
	if got := MaxInt(30, 80); got != 80 {
		t.Errorf("MaxInt(30, 80) 期望 80，得到 %d", got)
	}
}

func TestMinInt_WhenFirstSmaller_ExpectFirst(t *testing.T) {
	if got := MinInt(1, 99); got != 1 {
		t.Errorf("MinInt(1, 99) 期望 1，得到 %d", got)
	}
}

func TestMinInt_WhenSecondSmaller_ExpectSecond(t *testing.T) {
	if got := MinInt(55, 11); got != 11 {
		t.Errorf("MinInt(55, 11) 期望 11，得到 %d", got)
	}
}

// -----------------------------------------------------------------------
// AbsInt64
// -----------------------------------------------------------------------

func TestAbsInt64_WhenPositive_ExpectSameValue(t *testing.T) {
	if got := AbsInt64(42); got != 42 {
		t.Errorf("AbsInt64(42) 期望 42，得到 %d", got)
	}
}

func TestAbsInt64_WhenNegative_ExpectPositive(t *testing.T) {
	if got := AbsInt64(-42); got != 42 {
		t.Errorf("AbsInt64(-42) 期望 42，得到 %d", got)
	}
}

func TestAbsInt64_WhenZero_ExpectZero(t *testing.T) {
	if got := AbsInt64(0); got != 0 {
		t.Errorf("AbsInt64(0) 期望 0，得到 %d", got)
	}
}
