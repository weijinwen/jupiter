package xcast

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/spf13/cast"
)

// ToInt64Slice casts an interface to a []int64 type.
func ToInt64Slice(i any) []int64 {
	v, _ := ToInt64SliceE(i)
	return v
}

// ToInt64SliceE casts an empty interface to a []int64.
func ToInt64SliceE(i any) ([]int64, error) {
	if i == nil {
		return []int64{}, fmt.Errorf("Unable to Cast %#v to []int64", i)
	}

	switch v := i.(type) {
	case []int64:
		return v, nil
	}

	kind := reflect.TypeOf(i).Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(i)
		a := make([]int64, s.Len())
		for j := 0; j < s.Len(); j++ {
			val, err := cast.ToInt64E(s.Index(j).Interface())
			if err != nil {
				return []int64{}, fmt.Errorf("Unable to Cast %#v to []int64", i)
			}
			a[j] = val

		}
		return a, nil
	default:
		return []int64{}, fmt.Errorf("Unable to Cast %#v to []int64", i)
	}
}

// ToString casts an interface to a string type.
func ToString(i any) string {
	v, _ := ToStringE(i)
	return v
}

// ToStringE casts an interface to a string type.
func ToStringE(i any) (string, error) {
	return cast.ToStringE(i)
}

// ToInt casts an interface to an int type.
func ToInt(i any) int {
	v, _ := ToIntE(i)
	return v
}

// ToIntE casts an interface to an int type.
func ToIntE(i any) (int, error) {
	return cast.ToIntE(i)
}

// ToInt64 casts an interface to an int64 type.
func ToInt64(i any) int64 {
	v, _ := ToInt64E(i)
	return v
}

// ToInt64E casts an interface to an int64 type.
func ToInt64E(i any) (int64, error) {
	return cast.ToInt64E(i)
}

// ToFloat64 casts an interface to a float64 type.
func ToFloat64(i any) float64 {
	v, _ := ToFloat64E(i)
	return v
}

// ToFloat64E casts an interface to a float64 type.
func ToFloat64E(i any) (float64, error) {
	return cast.ToFloat64E(i)
}

// ToBool casts an interface to a bool type.
func ToBool(i any) bool {
	v, _ := ToBoolE(i)
	return v
}

// ToBoolE casts an interface to a bool type.
func ToBoolE(i any) (bool, error) {
	return cast.ToBoolE(i)
}

// ToStringSlice casts an interface to a []string type.
func ToStringSlice(i any) []string {
	v, _ := ToStringSliceE(i)
	return v
}

// ToStringSliceE casts an interface to a []string type.
func ToStringSliceE(i any) ([]string, error) {
	if i == nil {
		return []string{}, nil
	}

	switch v := i.(type) {
	case []string:
		return v, nil
	}

	kind := reflect.TypeOf(i).Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(i)
		a := make([]string, s.Len())
		for j := 0; j < s.Len(); j++ {
			val, err := cast.ToStringE(s.Index(j).Interface())
			if err != nil {
				return []string{}, fmt.Errorf("Unable to Cast %#v to []string", i)
			}
			a[j] = val
		}
		return a, nil
	default:
		return []string{}, fmt.Errorf("Unable to Cast %#v to []string", i)
	}
}

// ToIntSlice casts an interface to a []int type.
func ToIntSlice(i any) []int {
	v, _ := ToIntSliceE(i)
	return v
}

// ToIntSliceE casts an interface to a []int type.
func ToIntSliceE(i any) ([]int, error) {
	if i == nil {
		return []int{}, nil
	}

	switch v := i.(type) {
	case []int:
		return v, nil
	}

	kind := reflect.TypeOf(i).Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(i)
		a := make([]int, s.Len())
		for j := 0; j < s.Len(); j++ {
			val, err := cast.ToIntE(s.Index(j).Interface())
			if err != nil {
				return []int{}, fmt.Errorf("Unable to Cast %#v to []int", i)
			}
			a[j] = val
		}
		return a, nil
	default:
		return []int{}, fmt.Errorf("Unable to Cast %#v to []int", i)
	}
}

// ToFloat64Slice casts an interface to a []float64 type.
func ToFloat64Slice(i any) []float64 {
	v, _ := ToFloat64SliceE(i)
	return v
}

// ToFloat64SliceE casts an interface to a []float64 type.
func ToFloat64SliceE(i any) ([]float64, error) {
	if i == nil {
		return []float64{}, nil
	}

	switch v := i.(type) {
	case []float64:
		return v, nil
	}

	kind := reflect.TypeOf(i).Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(i)
		a := make([]float64, s.Len())
		for j := 0; j < s.Len(); j++ {
			val, err := cast.ToFloat64E(s.Index(j).Interface())
			if err != nil {
				return []float64{}, fmt.Errorf("Unable to Cast %#v to []float64", i)
			}
			a[j] = val
		}
		return a, nil
	default:
		return []float64{}, fmt.Errorf("Unable to Cast %#v to []float64", i)
	}
}

// ToStringMap casts an interface to a map[string]any type.
func ToStringMap(i any) map[string]any {
	v, _ := ToStringMapE(i)
	return v
}

// ToStringMapE casts an interface to a map[string]any type.
func ToStringMapE(i any) (map[string]any, error) {
	return cast.ToStringMapE(i)
}

// ToStringMapString casts an interface to a map[string]string type.
func ToStringMapString(i any) map[string]string {
	v, _ := ToStringMapStringE(i)
	return v
}

// ToStringMapStringE casts an interface to a map[string]string type.
func ToStringMapStringE(i any) (map[string]string, error) {
	return cast.ToStringMapStringE(i)
}

// ToMap casts an interface to a map[string]any type using JSON marshaling/unmarshaling.
func ToMap(i any) map[string]any {
	v, _ := ToMapE(i)
	return v
}

// ToMapE casts an interface to a map[string]any type using JSON marshaling/unmarshaling.
func ToMapE(i any) (map[string]any, error) {
	if i == nil {
		return nil, nil
	}

	switch v := i.(type) {
	case map[string]any:
		return v, nil
	case map[string]string:
		result := make(map[string]any, len(v))
		for k, val := range v {
			result[k] = val
		}
		return result, nil
	}

	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	var result map[string]any
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ToStringSliceMapString casts an interface to a []map[string]string type.
func ToStringSliceMapString(i any) []map[string]string {
	v, _ := ToStringSliceMapStringE(i)
	return v
}

// ToStringSliceMapStringE casts an interface to a []map[string]string type.
func ToStringSliceMapStringE(i any) ([]map[string]string, error) {
	if i == nil {
		return nil, nil
	}

	switch v := i.(type) {
	case []map[string]string:
		return v, nil
	}

	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	var result []map[string]string
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ToStringMapStringSlice casts an interface to a map[string][]string type.
func ToStringMapStringSlice(i any) map[string][]string {
	v, _ := ToStringMapStringSliceE(i)
	return v
}

// ToStringMapStringSliceE casts an interface to a map[string][]string type.
func ToStringMapStringSliceE(i any) (map[string][]string, error) {
	if i == nil {
		return nil, nil
	}

	switch v := i.(type) {
	case map[string][]string:
		return v, nil
	}

	data, err := json.Marshal(i)
	if err != nil {
		return nil, err
	}

	var result map[string][]string
	err = json.Unmarshal(data, &result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// ToSlice casts an interface to a []any type.
func ToSlice(i any) []any {
	v, _ := ToSliceE(i)
	return v
}

// ToSliceE casts an interface to a []any type.
func ToSliceE(i any) ([]any, error) {
	if i == nil {
		return nil, nil
	}

	switch v := i.(type) {
	case []any:
		return v, nil
	}

	kind := reflect.TypeOf(i).Kind()
	switch kind {
	case reflect.Slice, reflect.Array:
		s := reflect.ValueOf(i)
		a := make([]any, s.Len())
		for j := 0; j < s.Len(); j++ {
			a[j] = s.Index(j).Interface()
		}
		return a, nil
	default:
		return []any{i}, nil
	}
}

// ToUint casts an interface to a uint type.
func ToUint(i any) uint {
	v, _ := ToUintE(i)
	return v
}

// ToUintE casts an interface to a uint type.
func ToUintE(i any) (uint, error) {
	return cast.ToUintE(i)
}

// ToUint64 casts an interface to a uint64 type.
func ToUint64(i any) uint64 {
	v, _ := ToUint64E(i)
	return v
}

// ToUint64E casts an interface to a uint64 type.
func ToUint64E(i any) (uint64, error) {
	return cast.ToUint64E(i)
}

// ToTime casts an interface to a time.Time type.
func ToTime(i any) string {
	v, _ := ToStringE(i)
	return v
}

// ToTimeE casts an interface to a time.Time type.
func ToTimeE(i any) (string, error) {
	return cast.ToStringE(i)
}

// IsEmpty checks if the given value is empty.
func IsEmpty(i any) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	switch v.Kind() {
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Ptr, reflect.Interface:
		return v.IsNil()
	}
	return false
}

// IsNotEmpty checks if the given value is not empty.
func IsNotEmpty(i any) bool {
	return !IsEmpty(i)
}

// IsZero checks if the given value is the zero value for its type.
func IsZero(i any) bool {
	if i == nil {
		return true
	}
	v := reflect.ValueOf(i)
	return v.IsZero()
}

// IsNotZero checks if the given value is not the zero value for its type.
func IsNotZero(i any) bool {
	return !IsZero(i)
}

// ToPtr returns a pointer to the given value.
func ToPtr[T any](v T) *T {
	return &v
}

// ToVal dereferences a pointer to the given value, or returns the zero value if nil.
func ToVal[T any](ptr *T) T {
	if ptr == nil {
		var zero T
		return zero
	}
	return *ptr
}

// ToStringPtr returns a pointer to the given string value.
func ToStringPtr(s string) *string {
	return &s
}

// ToIntPtr returns a pointer to the given int value.
func ToIntPtr(i int) *int {
	return &i
}

// ToInt64Ptr returns a pointer to the given int64 value.
func ToInt64Ptr(i int64) *int64 {
	return &i
}

// ToFloat64Ptr returns a pointer to the given float64 value.
func ToFloat64Ptr(f float64) *float64 {
	return &f
}

// ToBoolPtr returns a pointer to the given bool value.
func ToBoolPtr(b bool) *bool {
	return &b
}

// IntToString converts an int to a string.
func IntToString(i int) string {
	return strconv.Itoa(i)
}

// Int64ToString converts an int64 to a string.
func Int64ToString(i int64) string {
	return strconv.FormatInt(i, 10)
}

// Float64ToString converts a float64 to a string with the specified precision.
func Float64ToString(f float64, precision int) string {
	return strconv.FormatFloat(f, 'f', precision, 64)
}

// StringToInt converts a string to an int.
func StringToInt(s string) (int, error) {
	return strconv.Atoi(s)
}

// StringToInt64 converts a string to an int64.
func StringToInt64(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

// StringToFloat64 converts a string to a float64.
func StringToFloat64(s string) (float64, error) {
	return strconv.ParseFloat(s, 64)
}

// StringToBool converts a string to a bool.
func StringToBool(s string) (bool, error) {
	return strconv.ParseBool(s)
}
