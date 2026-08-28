package xrand

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIntRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := IntRange(1, 10)
		assert.True(t, result >= 1 && result <= 10)
	}
}

func TestInt64Range(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := Int64Range(100, 1000)
		assert.True(t, result >= 100 && result <= 1000)
	}
}

func TestFloat64Range(t *testing.T) {
	for i := 0; i < 100; i++ {
		result := Float64Range(0.5, 1.5)
		assert.True(t, result >= 0.5 && result <= 1.5)
	}
}

func TestBool(t *testing.T) {
	trueCount := 0
	for i := 0; i < 100; i++ {
		if Bool() {
			trueCount++
		}
	}
	assert.True(t, trueCount > 20 && trueCount < 80)
}

func TestString(t *testing.T) {
	s := String(10)
	assert.Len(t, s, 10)
	for _, c := range s {
		assert.True(t, (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'))
	}
}

func TestStringWithChars(t *testing.T) {
	s := StringWithChars(5, "abc")
	assert.Len(t, s, 5)
	for _, c := range s {
		assert.True(t, c == 'a' || c == 'b' || c == 'c')
	}
}

func TestAlphanumericString(t *testing.T) {
	s := AlphanumericString(10)
	assert.Len(t, s, 10)
}

func TestNumericString(t *testing.T) {
	s := NumericString(10)
	assert.Len(t, s, 10)
	for _, c := range s {
		assert.True(t, c >= '0' && c <= '9')
	}
}

func TestUUID(t *testing.T) {
	uuid := UUID()
	assert.Len(t, uuid, 36)
	assert.Equal(t, byte('-'), uuid[8])
	assert.Equal(t, byte('-'), uuid[13])
	assert.Equal(t, byte('-'), uuid[18])
	assert.Equal(t, byte('-'), uuid[23])
}

func TestShortID(t *testing.T) {
	id := ShortID(8)
	assert.Len(t, id, 8)
}

func TestPick(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	for i := 0; i < 10; i++ {
		result := Pick(slice)
		assert.Contains(t, slice, result)
	}

	var emptySlice []int
	result := Pick(emptySlice)
	assert.Equal(t, 0, result)
}

func TestPickN(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := PickN(slice, 3)
	assert.Len(t, result, 3)

	result = PickN(slice, 10)
	assert.Len(t, result, 5)

	result = PickN(slice, 0)
	assert.Nil(t, result)
}

func TestShuffle(t *testing.T) {
	original := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}
	shuffled := Shuffle(original)

	assert.Equal(t, len(original), len(shuffled))
	assert.ElementsMatch(t, original, shuffled)

	assert.Equal(t, original, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10})
}

func TestSecureInt(t *testing.T) {
	n, err := SecureInt(100)
	assert.NoError(t, err)
	assert.True(t, n >= 0 && n < 100)
}

func TestSecureBytes(t *testing.T) {
	bytes, err := SecureBytes(16)
	assert.NoError(t, err)
	assert.Len(t, bytes, 16)
}

func TestWeightedPick(t *testing.T) {
	items := []string{"a", "b", "c"}
	weights := []int{1, 2, 3}

	counts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		result := WeightedPick(items, weights)
		counts[result]++
	}

	assert.True(t, counts["c"] > counts["b"])
	assert.True(t, counts["b"] > counts["a"])
}

func TestChance(t *testing.T) {
	trueCount := 0
	for i := 0; i < 1000; i++ {
		if Chance(0.3) {
			trueCount++
		}
	}
	assert.True(t, trueCount > 200 && trueCount < 400)

	assert.False(t, Chance(0))
	assert.False(t, Chance(-0.5))
	assert.True(t, Chance(1))
	assert.True(t, Chance(1.5))
}

func TestDuration(t *testing.T) {
	min := 100 * time.Millisecond
	max := 200 * time.Millisecond
	for i := 0; i < 100; i++ {
		result := Duration(min, max)
		assert.True(t, result >= min && result <= max)
	}
}

func TestTime(t *testing.T) {
	min := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	max := time.Date(2024, 12, 31, 23, 59, 59, 0, time.UTC)
	for i := 0; i < 100; i++ {
		result := Time(min, max)
		assert.True(t, (result.Equal(min) || result.After(min)) && (result.Equal(max) || result.Before(max)))
	}
}

func TestHex(t *testing.T) {
	h := Hex(16)
	assert.Len(t, h, 16)
	for _, c := range h {
		assert.True(t, (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f'))
	}
}

func TestPassword(t *testing.T) {
	pwd := Password(12)
	assert.Len(t, pwd, 12)
}
