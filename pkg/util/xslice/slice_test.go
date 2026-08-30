// Copyright 2020 Douyu
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package xslice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		target   int
		expected bool
	}{
		{"contains element", []int{1, 2, 3, 4, 5}, 3, true},
		{"does not contain element", []int{1, 2, 3, 4, 5}, 6, false},
		{"empty slice", []int{}, 1, false},
		{"nil slice", nil, 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, Contains(tt.slice, tt.target))
		})
	}
}

func TestFilter(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	result := Filter(slice, func(i int) bool {
		return i%2 == 0
	})
	assert.Equal(t, []int{2, 4, 6}, result)
}

func TestMap(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := Map(slice, func(i int) string {
		return string(rune('A' + i - 1))
	})
	assert.Equal(t, []string{"A", "B", "C", "D", "E"}, result)
}

func TestFlatMap(t *testing.T) {
	slice := []int{1, 2, 3}
	result := FlatMap(slice, func(i int) []int {
		return []int{i, i * 2}
	})
	assert.Equal(t, []int{1, 2, 2, 4, 3, 6}, result)
}

func TestReduce(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result := Reduce(slice, 0, func(acc, i int) int {
		return acc + i
	})
	assert.Equal(t, 15, result)
}

func TestUnique(t *testing.T) {
	slice := []int{1, 2, 2, 3, 3, 3, 4}
	result := Unique(slice)
	assert.Equal(t, []int{1, 2, 3, 4}, result)
}

func TestGroupBy(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	result := GroupBy(slice, func(i int) string {
		if i%2 == 0 {
			return "even"
		}
		return "odd"
	})
	assert.Equal(t, map[string][]int{
		"odd":  {1, 3, 5},
		"even": {2, 4, 6},
	}, result)
}

func TestPartition(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	matching, nonMatching := Partition(slice, func(i int) bool {
		return i%2 == 0
	})
	assert.Equal(t, []int{2, 4, 6}, matching)
	assert.Equal(t, []int{1, 3, 5}, nonMatching)
}

func TestFirst(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result, found := First(slice, func(i int) bool {
		return i > 3
	})
	assert.True(t, found)
	assert.Equal(t, 4, result)

	result, found = First(slice, func(i int) bool {
		return i > 10
	})
	assert.False(t, found)
}

func TestLast(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	result, found := Last(slice, func(i int) bool {
		return i > 3
	})
	assert.True(t, found)
	assert.Equal(t, 5, result)
}

func TestIndex(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	assert.Equal(t, 2, Index(slice, 3))
	assert.Equal(t, -1, Index(slice, 6))
}

func TestCount(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6}
	count := Count(slice, func(i int) bool {
		return i%2 == 0
	})
	assert.Equal(t, 3, count)
}

func TestAll(t *testing.T) {
	slice := []int{2, 4, 6, 8}
	assert.True(t, All(slice, func(i int) bool {
		return i%2 == 0
	}))

	slice2 := []int{2, 3, 6, 8}
	assert.False(t, All(slice2, func(i int) bool {
		return i%2 == 0
	}))
}

func TestAny(t *testing.T) {
	slice := []int{1, 3, 5, 7}
	assert.False(t, Any(slice, func(i int) bool {
		return i%2 == 0
	}))

	slice2 := []int{1, 2, 5, 7}
	assert.True(t, Any(slice2, func(i int) bool {
		return i%2 == 0
	}))
}

func TestSum(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	assert.Equal(t, 15, Sum(slice))

	floatSlice := []float64{1.5, 2.5, 3.0}
	assert.Equal(t, 7.0, Sum(floatSlice))
}

func TestMin(t *testing.T) {
	slice := []int{5, 3, 1, 4, 2}
	min, ok := Min(slice)
	assert.True(t, ok)
	assert.Equal(t, 1, min)

	_, ok = Min([]int{})
	assert.False(t, ok)
}

func TestMax(t *testing.T) {
	slice := []int{5, 3, 1, 4, 2}
	max, ok := Max(slice)
	assert.True(t, ok)
	assert.Equal(t, 5, max)

	_, ok = Max([]int{})
	assert.False(t, ok)
}

func TestReverse(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	Reverse(slice)
	assert.Equal(t, []int{5, 4, 3, 2, 1}, slice)
}

func TestCopy(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5}
	copy := Copy(slice)
	copy[0] = 100

	assert.Equal(t, []int{1, 2, 3, 4, 5}, slice)
	assert.Equal(t, []int{100, 2, 3, 4, 5}, copy)
}

func TestChunk(t *testing.T) {
	slice := []int{1, 2, 3, 4, 5, 6, 7}
	chunks := Chunk(slice, 3)
	assert.Equal(t, [][]int{{1, 2, 3}, {4, 5, 6}, {7}}, chunks)

	chunks = Chunk(slice, 10)
	assert.Equal(t, [][]int{{1, 2, 3, 4, 5, 6, 7}}, chunks)
}

func TestFlatten(t *testing.T) {
	slice := [][]int{{1, 2}, {3, 4}, {5}}
	result := Flatten(slice)
	assert.Equal(t, []int{1, 2, 3, 4, 5}, result)
}

func TestCompact(t *testing.T) {
	slice := []int{1, 0, 3, 0, 5}
	result := Compact(slice)
	assert.Equal(t, []int{1, 3, 5}, result)
}

func TestForEach(t *testing.T) {
	slice := []int{1, 2, 3}
	var result []int
	ForEach(slice, func(i int) {
		result = append(result, i*2)
	})
	assert.Equal(t, []int{2, 4, 6}, result)
}

func TestForEachIndex(t *testing.T) {
	slice := []string{"a", "b", "c"}
	var result []string
	ForEachIndex(slice, func(i int, s string) {
		result = append(result, s)
	})
	assert.Equal(t, []string{"a", "b", "c"}, result)
}
