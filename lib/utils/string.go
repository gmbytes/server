package utils

import (
	"strconv"
	"strings"
	"unicode"
)

func ValidateName(str string) (int, bool) {
	rs := []rune(str)
	length := 0
	for idx, r := range rs {
		if r < 255 && (r == '_' || r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || idx != 0 && r >= '0' && r <= '9') {
			length++
		} else if unicode.Is(unicode.Han, r) {
			length += 2
		} else {
			return 0, false
		}
	}
	return length, true
}

func StrLen(name string) int {
	var l int
	for _, r := range []rune(name) {
		i := int(r)
		if i > 256 {
			l += 2
		} else {
			l++
		}
	}
	return l
}

func Atoi(i string) int {
	v, err := strconv.Atoi(strings.TrimSpace(i))
	if err != nil {
		panic(err)
	}
	return v
}

func Atoi16(i string) int16 {
	return int16(Atoi(i))
}

func Atoi32(i string) int32 {
	return int32(Atoi(i))
}

func Atoi64(i string) int64 {
	v, err := strconv.ParseInt(strings.TrimSpace(i), 10, 64)
	if err != nil {
		panic(err)
	}
	return v
}

func Atou16(i string) uint16 {
	return uint16(Atoi(i))
}

func Atou32(i string) uint32 {
	return uint32(Atoi(i))
}

func Atou64(i string) uint64 {
	return uint64(Atoi64(i))
}

func Itoa(i int64) string {
	return strconv.Itoa(int(i))
}

func ItoaInt(i int) string {
	return strconv.Itoa(i)
}

func MaxInt64(a, b int64) int64 {
	if a > b {
		return a
	}
	return b
}

func MinInt64(a, b int64) int64 {
	if a < b {
		return a
	}
	return b
}

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func AbsInt64(a int64) int64 {
	if a < 0 {
		return -a
	}
	return a
}
