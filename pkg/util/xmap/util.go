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

package xmap

import (
	"fmt"
	"reflect"
	"sort"

	xcast "github.com/spf13/cast"
)

// MergeStringMap merge two map
func MergeStringMap(dest, src map[string]any) {
	for sk, sv := range src {
		tv, ok := dest[sk]
		if !ok {
			// val不存在时，直接赋值
			dest[sk] = sv
			continue
		}

		svType := reflect.TypeOf(sv)
		tvType := reflect.TypeOf(tv)
		if svType != tvType {
			fmt.Println("continue, type is different")
			continue
		}

		switch ttv := tv.(type) {
		case map[any]any:
			tsv := sv.(map[any]any)
			ssv := ToMapStringInterface(tsv)
			stv := ToMapStringInterface(ttv)
			MergeStringMap(stv, ssv)
			dest[sk] = stv
		case map[string]any:
			MergeStringMap(ttv, sv.(map[string]any))
			dest[sk] = ttv
		default:
			dest[sk] = sv
		}
	}
}

// ToMapStringInterface cast map[any]any to map[string]any
func ToMapStringInterface(src map[any]any) map[string]any {
	tgt := map[string]any{}
	for k, v := range src {
		tgt[fmt.Sprintf("%v", k)] = v
	}
	return tgt
}

// DeepSearchInMap deep search in map
func DeepSearchInMap(m map[string]any, paths ...string) map[string]any {
	//深度拷贝
	mtmp := make(map[string]any)
	for k, v := range m {
		mtmp[k] = v
	}
	for _, k := range paths {
		m2, ok := mtmp[k]
		if !ok {
			m3 := make(map[string]any)
			mtmp[k] = m3
			mtmp = m3
			continue
		}

		m3, err := xcast.ToStringMapE(m2)
		if err != nil {
			m3 = make(map[string]any)
			mtmp[k] = m3
		}
		// continue search
		mtmp = m3
	}
	return mtmp
}

// Keys returns all keys of the map.
func Keys[K comparable, V any](m map[K]V) []K {
	keys := make([]K, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// Values returns all values of the map.
func Values[K comparable, V any](m map[K]V) []V {
	values := make([]V, 0, len(m))
	for _, v := range m {
		values = append(values, v)
	}
	return values
}

// Entries returns all key-value pairs of the map as a slice of [K, V] pairs.
func Entries[K comparable, V any](m map[K]V) [][2]any {
	entries := make([][2]any, 0, len(m))
	for k, v := range m {
		entries = append(entries, [2]any{k, v})
	}
	return entries
}

// ContainsKey checks if the map contains the given key.
func ContainsKey[K comparable, V any](m map[K]V, key K) bool {
	_, ok := m[key]
	return ok
}

// ContainsValue checks if the map contains the given value.
func ContainsValue[K comparable, V comparable](m map[K]V, value V) bool {
	for _, v := range m {
		if v == value {
			return true
		}
	}
	return false
}

// GetOrDefault returns the value for the given key, or the default value if not found.
func GetOrDefault[K comparable, V any](m map[K]V, key K, defaultValue V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultValue
}

// GetOrElse returns the value for the given key, or the result of the default function if not found.
func GetOrElse[K comparable, V any](m map[K]V, key K, defaultFn func() V) V {
	if v, ok := m[key]; ok {
		return v
	}
	return defaultFn()
}

// Filter returns a new map containing only the key-value pairs that satisfy the predicate.
func Filter[K comparable, V any](m map[K]V, predicate func(K, V) bool) map[K]V {
	result := make(map[K]V)
	for k, v := range m {
		if predicate(k, v) {
			result[k] = v
		}
	}
	return result
}

// MapValues returns a new map with the same keys but transformed values.
func MapValues[K comparable, V any, R any](m map[K]V, transform func(K, V) R) map[K]R {
	result := make(map[K]R)
	for k, v := range m {
		result[k] = transform(k, v)
	}
	return result
}

// MapKeys returns a new map with transformed keys and the same values.
func MapKeys[K comparable, V any, R comparable](m map[K]V, transform func(K, V) R) map[R]V {
	result := make(map[R]V)
	for k, v := range m {
		result[transform(k, v)] = v
	}
	return result
}

// ForEach executes the given function for each key-value pair.
func ForEach[K comparable, V any](m map[K]V, fn func(K, V)) {
	for k, v := range m {
		fn(k, v)
	}
}

// Size returns the number of key-value pairs in the map.
func Size[K comparable, V any](m map[K]V) int {
	return len(m)
}

// IsEmpty checks if the map is empty.
func IsEmpty[K comparable, V any](m map[K]V) bool {
	return len(m) == 0
}

// IsNotEmpty checks if the map is not empty.
func IsNotEmpty[K comparable, V any](m map[K]V) bool {
	return len(m) > 0
}

// Clear removes all key-value pairs from the map.
func Clear[K comparable, V any](m map[K]V) {
	for k := range m {
		delete(m, k)
	}
}

// Copy returns a shallow copy of the map.
func Copy[K comparable, V any](m map[K]V) map[K]V {
	result := make(map[K]V, len(m))
	for k, v := range m {
		result[k] = v
	}
	return result
}

// Merge returns a new map containing all key-value pairs from both maps.
// If a key exists in both maps, the value from the second map is used.
func Merge[K comparable, V any](m1, m2 map[K]V) map[K]V {
	result := make(map[K]V, len(m1)+len(m2))
	for k, v := range m1 {
		result[k] = v
	}
	for k, v := range m2 {
		result[k] = v
	}
	return result
}

// MergeInPlace merges the second map into the first map.
func MergeInPlace[K comparable, V any](dest, src map[K]V) {
	for k, v := range src {
		dest[k] = v
	}
}

// KeysSorted returns all keys of the map sorted in ascending order.
func KeysSorted[K ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64 | ~string, V any](m map[K]V) []K {
	keys := Keys(m)
	sort.Slice(keys, func(i, j int) bool {
		return fmt.Sprintf("%v", keys[i]) < fmt.Sprintf("%v", keys[j])
	})
	return keys
}

// KeysSortedFunc returns all keys of the map sorted using the given comparison function.
func KeysSortedFunc[K comparable, V any](m map[K]V, less func(K, K) bool) []K {
	keys := Keys(m)
	sort.Slice(keys, func(i, j int) bool {
		return less(keys[i], keys[j])
	})
	return keys
}

// Equal checks if two maps are equal (have the same key-value pairs).
func Equal[K comparable, V comparable](m1, m2 map[K]V) bool {
	if len(m1) != len(m2) {
		return false
	}
	for k, v := range m1 {
		if v2, ok := m2[k]; !ok || v != v2 {
			return false
		}
	}
	return true
}

// Pick returns a new map containing only the specified keys.
func Pick[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	result := make(map[K]V, len(keys))
	for _, k := range keys {
		if v, ok := m[k]; ok {
			result[k] = v
		}
	}
	return result
}

// Omit returns a new map without the specified keys.
func Omit[K comparable, V any](m map[K]V, keys ...K) map[K]V {
	omitSet := make(map[K]struct{}, len(keys))
	for _, k := range keys {
		omitSet[k] = struct{}{}
	}
	result := make(map[K]V)
	for k, v := range m {
		if _, ok := omitSet[k]; !ok {
			result[k] = v
		}
	}
	return result
}

// GroupBy groups the values of the map by the key returned by the key function.
func GroupBy[K comparable, V any, R comparable](m map[K]V, keyFn func(K, V) R) map[R][]V {
	groups := make(map[R][]V)
	for k, v := range m {
		key := keyFn(k, v)
		groups[key] = append(groups[key], v)
	}
	return groups
}

// Reduce reduces the map to a single value using the given function.
func Reduce[K comparable, V any, R any](m map[K]V, initial R, reducer func(R, K, V) R) R {
	result := initial
	for k, v := range m {
		result = reducer(result, k, v)
	}
	return result
}

// Any checks if any key-value pair satisfies the predicate.
func Any[K comparable, V any](m map[K]V, predicate func(K, V) bool) bool {
	for k, v := range m {
		if predicate(k, v) {
			return true
		}
	}
	return false
}

// All checks if all key-value pairs satisfy the predicate.
func All[K comparable, V any](m map[K]V, predicate func(K, V) bool) bool {
	for k, v := range m {
		if !predicate(k, v) {
			return false
		}
	}
	return true
}

// Count counts the number of key-value pairs that satisfy the predicate.
func Count[K comparable, V any](m map[K]V, predicate func(K, V) bool) int {
	count := 0
	for k, v := range m {
		if predicate(k, v) {
			count++
		}
	}
	return count
}

// FromSlice creates a map from a slice using the key function.
func FromSlice[K comparable, V any](slice []V, keyFn func(V) K) map[K]V {
	result := make(map[K]V, len(slice))
	for _, v := range slice {
		result[keyFn(v)] = v
	}
	return result
}

// FromSliceWithValues creates a map from a slice with both key and value functions.
func FromSliceWithValues[K comparable, V any, R any](slice []V, keyFn func(V) K, valueFn func(V) R) map[K]R {
	result := make(map[K]R, len(slice))
	for _, v := range slice {
		result[keyFn(v)] = valueFn(v)
	}
	return result
}
