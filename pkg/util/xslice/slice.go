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

// Contains checks if an element is present in a slice.
func Contains[T comparable](slice []T, target T) bool {
	for _, item := range slice {
		if item == target {
			return true
		}
	}
	return false
}

// ContainsFunc checks if any element satisfies the predicate.
func ContainsFunc[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}
	return false
}

// Filter returns elements that satisfy the predicate.
func Filter[T any](slice []T, predicate func(T) bool) []T {
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if predicate(item) {
			result = append(result, item)
		}
	}
	return result
}

// Map transforms each element using the provided function.
func Map[T any, R any](slice []T, transform func(T) R) []R {
	result := make([]R, len(slice))
	for i, item := range slice {
		result[i] = transform(item)
	}
	return result
}

// FlatMap transforms each element to a slice and flattens the result.
func FlatMap[T any, R any](slice []T, transform func(T) []R) []R {
	result := make([]R, 0, len(slice))
	for _, item := range slice {
		result = append(result, transform(item)...)
	}
	return result
}

// Reduce reduces the slice to a single value using the provided function.
func Reduce[T any, R any](slice []T, initial R, reducer func(R, T) R) R {
	result := initial
	for _, item := range slice {
		result = reducer(result, item)
	}
	return result
}

// ForEach executes the provided function for each element.
func ForEach[T any](slice []T, fn func(T)) {
	for _, item := range slice {
		fn(item)
	}
}

// ForEachIndex executes the provided function for each element with index.
func ForEachIndex[T any](slice []T, fn func(int, T)) {
	for i, item := range slice {
		fn(i, item)
	}
}

// Unique removes duplicate elements from the slice.
func Unique[T comparable](slice []T) []T {
	seen := make(map[T]struct{}, len(slice))
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		if _, exists := seen[item]; !exists {
			seen[item] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// UniqueFunc removes duplicate elements using a key function.
func UniqueFunc[T any, K comparable](slice []T, keyFn func(T) K) []T {
	seen := make(map[K]struct{}, len(slice))
	result := make([]T, 0, len(slice))
	for _, item := range slice {
		key := keyFn(item)
		if _, exists := seen[key]; !exists {
			seen[key] = struct{}{}
			result = append(result, item)
		}
	}
	return result
}

// GroupBy groups elements by the key returned by the key function.
func GroupBy[T any, K comparable](slice []T, keyFn func(T) K) map[K][]T {
	groups := make(map[K][]T)
	for _, item := range slice {
		key := keyFn(item)
		groups[key] = append(groups[key], item)
	}
	return groups
}

// Partition splits the slice into two groups based on the predicate.
func Partition[T any](slice []T, predicate func(T) bool) ([]T, []T) {
	matching := make([]T, 0, len(slice))
	nonMatching := make([]T, 0, len(slice))
	for _, item := range slice {
		if predicate(item) {
			matching = append(matching, item)
		} else {
			nonMatching = append(nonMatching, item)
		}
	}
	return matching, nonMatching
}

// First returns the first element that satisfies the predicate, or zero value if not found.
func First[T any](slice []T, predicate func(T) bool) (T, bool) {
	for _, item := range slice {
		if predicate(item) {
			return item, true
		}
	}
	var zero T
	return zero, false
}

// Last returns the last element that satisfies the predicate, or zero value if not found.
func Last[T any](slice []T, predicate func(T) bool) (T, bool) {
	var zero T
	for i := len(slice) - 1; i >= 0; i-- {
		if predicate(slice[i]) {
			return slice[i], true
		}
	}
	return zero, false
}

// Index returns the index of the first occurrence of the target, or -1 if not found.
func Index[T comparable](slice []T, target T) int {
	for i, item := range slice {
		if item == target {
			return i
		}
	}
	return -1
}

// IndexFunc returns the index of the first element that satisfies the predicate, or -1 if not found.
func IndexFunc[T any](slice []T, predicate func(T) bool) int {
	for i, item := range slice {
		if predicate(item) {
			return i
		}
	}
	return -1
}

// Count returns the number of elements that satisfy the predicate.
func Count[T any](slice []T, predicate func(T) bool) int {
	count := 0
	for _, item := range slice {
		if predicate(item) {
			count++
		}
	}
	return count
}

// All returns true if all elements satisfy the predicate.
func All[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if !predicate(item) {
			return false
		}
	}
	return true
}

// Any returns true if any element satisfies the predicate.
func Any[T any](slice []T, predicate func(T) bool) bool {
	for _, item := range slice {
		if predicate(item) {
			return true
		}
	}
	return false
}

// Sum returns the sum of all elements.
func Sum[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](slice []T) T {
	var sum T
	for _, item := range slice {
		sum += item
	}
	return sum
}

// Min returns the minimum element in the slice.
func Min[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	min := slice[0]
	for _, item := range slice[1:] {
		if item < min {
			min = item
		}
	}
	return min, true
}

// Max returns the maximum element in the slice.
func Max[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](slice []T) (T, bool) {
	if len(slice) == 0 {
		var zero T
		return zero, false
	}
	max := slice[0]
	for _, item := range slice[1:] {
		if item > max {
			max = item
		}
	}
	return max, true
}

// Reverse reverses the slice in place.
func Reverse[T any](slice []T) {
	for i, j := 0, len(slice)-1; i < j; i, j = i+1, j-1 {
		slice[i], slice[j] = slice[j], slice[i]
	}
}

// Copy creates a copy of the slice.
func Copy[T any](slice []T) []T {
	result := make([]T, len(slice))
	copy(result, slice)
	return result
}

// Chunk splits the slice into chunks of the specified size.
func Chunk[T any](slice []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	chunks := make([][]T, 0, (len(slice)+size-1)/size)
	for i := 0; i < len(slice); i += size {
		end := i + size
		if end > len(slice) {
			end = len(slice)
		}
		chunks = append(chunks, slice[i:end])
	}
	return chunks
}

// Flatten flattens a slice of slices into a single slice.
func Flatten[T any](slice [][]T) []T {
	result := make([]T, 0)
	for _, sub := range slice {
		result = append(result, sub...)
	}
	return result
}

// Compact removes zero values from the slice.
func Compact[T comparable](slice []T) []T {
	var zero T
	return Filter(slice, func(item T) bool {
		return item != zero
	})
}
