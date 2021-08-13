package stringx

import (
	"errors"
	"fmt"
	"strconv"

	"gitlab.deepwisdomai.com/infra/go-zero/core/lang"
)

var (
	// ErrInvalidStartPosition is an error that indicates the start position is invalid.
	ErrInvalidStartPosition = errors.New("start position is invalid")
	// ErrInvalidStopPosition is an error that indicates the stop position is invalid.
	ErrInvalidStopPosition = errors.New("stop position is invalid")
)

// Contains checks if str is in list.
func Contains(list []string, str string) bool {
	for _, each := range list {
		if each == str {
			return true
		}
	}

	return false
}

// Filter filters chars from s with given filter function.
func Filter(s string, filter func(r rune) bool) string {
	var n int
	chars := []rune(s)
	for i, x := range chars {
		if n < i {
			chars[n] = x
		}
		if !filter(x) {
			n++
		}
	}

	return string(chars[:n])
}

// HasEmpty checks if there are empty strings in args.
func HasEmpty(args ...string) bool {
	for _, arg := range args {
		if len(arg) == 0 {
			return true
		}
	}

	return false
}

// NotEmpty checks if all strings are not empty in args.
func NotEmpty(args ...string) bool {
	return !HasEmpty(args...)
}

// Remove removes given strs from strings.
func Remove(strings []string, strs ...string) []string {
	out := append([]string(nil), strings...)

	for _, str := range strs {
		var n int
		for _, v := range out {
			if v != str {
				out[n] = v
				n++
			}
		}
		out = out[:n]
	}

	return out
}

// Reverse reverses s.
func Reverse(s string) string {
	runes := []rune(s)

	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}

	return string(runes)
}

// Substr returns runes between start and stop [start, stop) regardless of the chars are ascii or utf8.
func Substr(str string, start, stop int) (string, error) {
	rs := []rune(str)
	length := len(rs)

	if start < 0 || start > length {
		return "", ErrInvalidStartPosition
	}

	if stop < 0 || stop > length {
		return "", ErrInvalidStopPosition
	}

	return string(rs[start:stop]), nil
}

// TakeOne returns valid string if not empty or later one.
func TakeOne(valid, or string) string {
	if len(valid) > 0 {
		return valid
	}

	return or
}

// TakeWithPriority returns the first not empty result from fns.
func TakeWithPriority(fns ...func() string) string {
	for _, fn := range fns {
		val := fn()
		if len(val) > 0 {
			return val
		}
	}

	return ""
}

// Union merges the strings in first and second.
func Union(first, second []string) []string {
	set := make(map[string]lang.PlaceholderType)

	for _, each := range first {
		set[each] = lang.Placeholder
	}
	for _, each := range second {
		set[each] = lang.Placeholder
	}

	merged := make([]string, 0, len(set))
	for k := range set {
		merged = append(merged, k)
	}

	return merged
}

// byte转换成C_string（以'\0'作为字符串结束的标志）
func ByteToCString(p []byte) string {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return string(p[0:i])
		}
	}
	return string(p)
}

// 计算C_string风格字符串的长度（以'\0'作为字符串结束的标志）
func CStrLen(p []byte) int {
	for i := 0; i < len(p); i++ {
		if p[i] == 0 {
			return i
		}
	}
	return len(p)
}

// 二进制转为可读的字符串，eg. []byte{123,77} => "7B4D"
func Bin2Str(p []byte) string {
	if len(p) == 0 {
		return ""
	}

	s := ""
	for i := 0; i < len(p); i++ {
		s += fmt.Sprintf("%02x", p[i])
	}
	return s
}

// 可读字符串转为二进制，eg. "7B4D" => []byte{123,77}
func Str2Bin(s string) []byte {
	buf := []byte{}

	for i := 0; i < len(s); i += 2 {
		sNum := ""
		if i+1 >= len(s) {
			sNum = string(s[i]) + "0"
		} else {
			sNum = s[i : i+2]
		}
		if v, err := strconv.ParseUint(sNum, 16, 8); err == nil {
			buf = append(buf, uint8(v))
		} else {
			buf = append(buf, uint8(0))
		}
	}

	return buf
}

// 是否是16进制字符串
// 判断标准是：（1）只能有 0-9, A-F, a-f；（2）字母要么全大写，要么全小写
func IsHexStr(src string) bool {
	if len(src) <= 0 || len(src)%2 != 0 {
		return false
	}

	hasUp := false
	hasLower := false
	for _, c := range src {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}

		if !hasUp {
			hasUp = (c >= 'A' && c <= 'F')
		}
		if !hasLower {
			hasLower = (c >= 'a' && c <= 'f')
		}

		if hasUp && hasLower {
			return false
		}
	}
	return true
}

// 16进制字符串=>10进制数（eg. "1F"  => 1*16+15 = 31）
// 16进制字符串中不含负号
func Hex2Int(src string) (uint64, error) {
	return strconv.ParseUint(src, 16, 64)
}

func HasString(dest_slice []string, s string) bool {
	for _, str := range dest_slice {
		if s == str {
			return true
		}
	}
	return false
}

//MaskStringTail will replace last N characters with "*". example: input "abcde" will return "ab***"
func MaskStringTail(src string, maskLen int) (dst string) {
	strRune := []rune(src)
	strLen := len(strRune)
	if strLen < maskLen {
		maskLen = strLen
	}

	dstRune := strRune[0 : strLen-maskLen]
	dst = string(dstRune) + "***"
	return dst
}

//DotsStringTail will remove all characters after headLen, and append "..." to tail. example: input "abcde" will return "ab..."
func DotsStringTail(src string, headLen int) (dst string) {
	strRune := []rune(src)
	strLen := len(strRune)
	if strLen <= headLen+3 {
		return src
	}

	dstRune := strRune[0:headLen]
	dst = string(dstRune) + "..."
	return dst
}

//MaskStringMiddle will replace middle N characters with "*", first & last character would reserved. example: input "abcde" will return "a***e"
func MaskStringMiddle(src string) (dst string) {
	strRune := []rune(src)
	strLen := len(strRune)
	if strLen < 3 {
		return "**"
	}

	headRune := strRune[0:1]
	tailRune := strRune[strLen-1 : strLen]
	dst = string(headRune) + "**" + string(tailRune)
	return dst
}

func ReverseString(s string) string {
	runes := []rune(s)
	for from, to := 0, len(runes)-1; from < to; from, to = from+1, to-1 {
		runes[from], runes[to] = runes[to], runes[from]
	}
	return string(runes)
}
