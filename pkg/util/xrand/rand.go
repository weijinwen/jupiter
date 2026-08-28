package xrand

import (
	"crypto/rand"
	"math/big"
	mrand "math/rand"
	"sync"
	"time"
)

var (
	source = mrand.New(mrand.NewSource(time.Now().UnixNano()))
	mu     sync.Mutex
)

const (
	LowercaseLetters = "abcdefghijklmnopqrstuvwxyz"
	UppercaseLetters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Letters          = LowercaseLetters + UppercaseLetters
	Digits           = "0123456789"
	Alphanumeric     = Letters + Digits
	SpecialChars     = "!@#$%^&*()_+-=[]{}|;':\",./<>?"
	AllChars         = Alphanumeric + SpecialChars
)

// Int 返回一个非负的随机整数
func Int() int {
	mu.Lock()
	defer mu.Unlock()
	return source.Int()
}

// Intn 返回 [0, n) 范围内的随机整数
func Intn(n int) int {
	if n <= 0 {
		return 0
	}
	mu.Lock()
	defer mu.Unlock()
	return source.Intn(n)
}

// IntRange 返回 [min, max] 范围内的随机整数
func IntRange(min, max int) int {
	if min >= max {
		return min
	}
	return min + Intn(max-min+1)
}

// Int64 返回一个非负的 64 位随机整数
func Int64() int64 {
	mu.Lock()
	defer mu.Unlock()
	return source.Int63()
}

// Int64n 返回 [0, n) 范围内的随机 64 位整数
func Int64n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	mu.Lock()
	defer mu.Unlock()
	return source.Int63n(n)
}

// Int64Range 返回 [min, max] 范围内的随机 64 位整数
func Int64Range(min, max int64) int64 {
	if min >= max {
		return min
	}
	return min + Int64n(max-min+1)
}

// Float64 返回 [0.0, 1.0) 范围内的随机浮点数
func Float64() float64 {
	mu.Lock()
	defer mu.Unlock()
	return source.Float64()
}

// Float64Range 返回 [min, max) 范围内的随机浮点数
func Float64Range(min, max float64) float64 {
	if min >= max {
		return min
	}
	return min + Float64()*(max-min)
}

// Bool 返回一个随机布尔值
func Bool() bool {
	return Intn(2) == 1
}

// String 生成指定长度的随机字母字符串
func String(length int) string {
	return StringWithChars(length, Letters)
}

// StringWithChars 使用指定字符集生成随机字符串
func StringWithChars(length int, chars string) string {
	if length <= 0 {
		return ""
	}
	if chars == "" {
		chars = Letters
	}

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = chars[Intn(len(chars))]
	}
	return string(result)
}

// AlphanumericString 生成指定长度的随机字母数字字符串
func AlphanumericString(length int) string {
	return StringWithChars(length, Alphanumeric)
}

// NumericString 生成指定长度的随机数字字符串
func NumericString(length int) string {
	return StringWithChars(length, Digits)
}

// Password 生成指定长度的随机密码字符串（包含特殊字符）
func Password(length int) string {
	return StringWithChars(length, AllChars)
}

// Hex 生成指定长度的随机十六进制字符串
func Hex(length int) string {
	return StringWithChars(length, "0123456789abcdef")
}

// UUID 生成一个 UUID 字符串
func UUID() string {
	return UUIDv4()
}

// UUIDv4 生成一个 UUID v4 字符串
func UUIDv4() string {
	bytes := make([]byte, 16)
	for i := 0; i < 16; i++ {
		bytes[i] = byte(Intn(256))
	}
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	hex := make([]byte, 36)
	hex[8] = '-'
	hex[13] = '-'
	hex[18] = '-'
	hex[23] = '-'

	hexEncode := func(b byte) byte {
		if b < 10 {
			return b + '0'
		}
		return b - 10 + 'a'
	}

	hex[0] = hexEncode(bytes[0] >> 4)
	hex[1] = hexEncode(bytes[0] & 0x0f)
	hex[2] = hexEncode(bytes[1] >> 4)
	hex[3] = hexEncode(bytes[1] & 0x0f)
	hex[4] = hexEncode(bytes[2] >> 4)
	hex[5] = hexEncode(bytes[2] & 0x0f)
	hex[6] = hexEncode(bytes[3] >> 4)
	hex[7] = hexEncode(bytes[3] & 0x0f)

	hex[9] = hexEncode(bytes[4] >> 4)
	hex[10] = hexEncode(bytes[4] & 0x0f)
	hex[11] = hexEncode(bytes[5] >> 4)
	hex[12] = hexEncode(bytes[5] & 0x0f)

	hex[14] = hexEncode(bytes[6] >> 4)
	hex[15] = hexEncode(bytes[6] & 0x0f)
	hex[16] = hexEncode(bytes[7] >> 4)
	hex[17] = hexEncode(bytes[7] & 0x0f)

	hex[19] = hexEncode(bytes[8] >> 4)
	hex[20] = hexEncode(bytes[8] & 0x0f)
	hex[21] = hexEncode(bytes[9] >> 4)
	hex[22] = hexEncode(bytes[9] & 0x0f)

	hex[24] = hexEncode(bytes[10] >> 4)
	hex[25] = hexEncode(bytes[10] & 0x0f)
	hex[26] = hexEncode(bytes[11] >> 4)
	hex[27] = hexEncode(bytes[11] & 0x0f)
	hex[28] = hexEncode(bytes[12] >> 4)
	hex[29] = hexEncode(bytes[12] & 0x0f)
	hex[30] = hexEncode(bytes[13] >> 4)
	hex[31] = hexEncode(bytes[13] & 0x0f)
	hex[32] = hexEncode(bytes[14] >> 4)
	hex[33] = hexEncode(bytes[14] & 0x0f)
	hex[34] = hexEncode(bytes[15] >> 4)
	hex[35] = hexEncode(bytes[15] & 0x0f)

	return string(hex)
}

// ShortID 生成指定长度的短 ID 字符串
func ShortID(length int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	return StringWithChars(length, chars)
}

// Pick 从切片中随机选择一个元素
func Pick[T any](slice []T) T {
	var zero T
	if len(slice) == 0 {
		return zero
	}
	return slice[Intn(len(slice))]
}

// PickN 从切片中随机选择 n 个不重复的元素
func PickN[T any](slice []T, n int) []T {
	if len(slice) == 0 || n <= 0 {
		return nil
	}
	if n >= len(slice) {
		return slice
	}

	result := make([]T, n)
	indices := make([]int, len(slice))
	for i := range indices {
		indices[i] = i
	}

	for i := 0; i < n; i++ {
		j := Intn(len(indices))
		result[i] = slice[indices[j]]
		indices = append(indices[:j], indices[j+1:]...)
	}
	return result
}

// Shuffle 随机打乱切片中元素的顺序
func Shuffle[T any](slice []T) []T {
	if len(slice) <= 1 {
		return slice
	}

	result := make([]T, len(slice))
	copy(result, slice)

	mu.Lock()
	defer mu.Unlock()
	source.Shuffle(len(result), func(i, j int) {
		result[i], result[j] = result[j], result[i]
	})

	return result
}

// SecureInt 使用加密安全的随机数生成器生成 [0, max) 范围内的随机整数
func SecureInt(max int64) (int64, error) {
	if max <= 0 {
		return 0, nil
	}
	n, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		return 0, err
	}
	return n.Int64(), nil
}

// SecureBytes 使用加密安全的随机数生成器生成指定长度的随机字节切片
func SecureBytes(length int) ([]byte, error) {
	if length <= 0 {
		return nil, nil
	}
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		return nil, err
	}
	return bytes, nil
}

// SecureString 使用加密安全的随机数生成器生成指定长度的随机字符串
func SecureString(length int) (string, error) {
	_, err := SecureBytes(length)
	if err != nil {
		return "", err
	}
	return StringWithChars(length, Alphanumeric), nil
}

// WeightedPick 根据权重从切片中随机选择一个元素
func WeightedPick[T any](items []T, weights []int) T {
	var zero T
	if len(items) == 0 || len(weights) == 0 || len(items) != len(weights) {
		return zero
	}

	total := 0
	for _, w := range weights {
		total += w
	}

	if total == 0 {
		return zero
	}

	r := Intn(total)
	for i, w := range weights {
		r -= w
		if r < 0 {
			return items[i]
		}
	}

	return items[len(items)-1]
}

// Chance 根据概率返回 true 或 false，probability 范围为 [0, 1]
func Chance(probability float64) bool {
	if probability <= 0 {
		return false
	}
	if probability >= 1 {
		return true
	}
	return Float64() < probability
}

// Duration 返回指定范围内的随机时间间隔
func Duration(min, max time.Duration) time.Duration {
	if min >= max {
		return min
	}
	return min + time.Duration(Int64n(int64(max-min)))
}

// Time 返回指定时间范围内的随机时间
func Time(min, max time.Time) time.Time {
	if min.After(max) || min.Equal(max) {
		return min
	}
	delta := max.Unix() - min.Unix()
	return min.Add(time.Duration(Int64n(delta)) * time.Second)
}
