package xstruct

import (
	"encoding/json"
	"reflect"
)

// Copy 通过 JSON 序列化/反序列化实现结构体深拷贝
func Copy(src, dst any) error {
	if src == nil || dst == nil {
		return nil
	}

	data, err := json.Marshal(src)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dst)
}

// Clone 克隆一个结构体，返回新的副本
func Clone[T any](src T) (T, error) {
	var dst T
	err := Copy(src, &dst)
	return dst, err
}

// MustClone 克隆一个结构体，如果出错则 panic
func MustClone[T any](src T) T {
	dst, err := Clone(src)
	if err != nil {
		panic(err)
	}
	return dst
}

// GetField 获取结构体指定字段的值
func GetField(obj any, field string) (any, error) {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil, nil
	}
	f := rv.FieldByName(field)
	if !f.IsValid() {
		return nil, nil
	}
	return f.Interface(), nil
}

// SetField 设置结构体指定字段的值
func SetField(obj any, field string, value any) error {
	rv := reflect.ValueOf(obj)
	if rv.Kind() != reflect.Ptr {
		return nil
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return nil
	}
	f := rv.FieldByName(field)
	if !f.IsValid() || !f.CanSet() {
		return nil
	}
	fv := reflect.ValueOf(value)
	if fv.Type().ConvertibleTo(f.Type()) {
		f.Set(fv.Convert(f.Type()))
	}
	return nil
}

// GetFieldNames 获取结构体所有字段名称
func GetFieldNames(obj any) []string {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()
	names := make([]string, 0, rt.NumField())
	for i := 0; i < rt.NumField(); i++ {
		names = append(names, rt.Field(i).Name)
	}
	return names
}

// GetFieldValues 获取结构体所有字段值
func GetFieldValues(obj any) []any {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	values := make([]any, 0, rv.NumField())
	for i := 0; i < rv.NumField(); i++ {
		values = append(values, rv.Field(i).Interface())
	}
	return values
}

// GetFieldMap 将结构体转换为字段名到字段值的映射
func GetFieldMap(obj any) map[string]any {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()
	result := make(map[string]any)
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		result[field.Name] = rv.Field(i).Interface()
	}
	return result
}

// ToMap 将结构体转换为 map，支持通过 tag 指定键名
func ToMap(obj any, tagName ...string) map[string]any {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	tag := "json"
	if len(tagName) > 0 {
		tag = tagName[0]
	}

	rt := rv.Type()
	result := make(map[string]any)
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		tagValue := field.Tag.Get(tag)
		if tagValue == "-" {
			continue
		}
		key := field.Name
		if tagValue != "" {
			if idx := len(tagValue); idx > 0 {
				for j, c := range tagValue {
					if c == ',' {
						idx = j
						break
					}
				}
				if idx > 0 && tagValue[:idx] != "" {
					key = tagValue[:idx]
				}
			}
		}
		result[key] = rv.Field(i).Interface()
	}
	return result
}

// MapToStruct 将 map 转换为结构体
func MapToStruct(m map[string]any, obj any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, obj)
}

// HasField 检查结构体是否包含指定字段
func HasField(obj any, field string) bool {
	rv := reflect.ValueOf(obj)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return false
	}
	f := rv.FieldByName(field)
	return f.IsValid()
}

// IsStruct 检查值是否为结构体类型
func IsStruct(v any) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.Kind() == reflect.Struct
}

// IsSlice 检查值是否为切片类型
func IsSlice(v any) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.Kind() == reflect.Slice
}

// IsMap 检查值是否为 map 类型
func IsMap(v any) bool {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	return rv.Kind() == reflect.Map
}

// TypeName 获取类型的名称
func TypeName(v any) string {
	if v == nil {
		return ""
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Name()
}

// FullName 获取类型的完整名称（包含包路径）
func FullName(v any) string {
	if v == nil {
		return ""
	}
	t := reflect.TypeOf(v)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.PkgPath() + "." + t.Name()
}

// Merge 将 src 结构体的非零值字段合并到 dst 结构体
func Merge(dst, src any) error {
	if dst == nil || src == nil {
		return nil
	}

	dstRV := reflect.ValueOf(dst)
	srcRV := reflect.ValueOf(src)

	if dstRV.Kind() != reflect.Ptr || srcRV.Kind() != reflect.Ptr {
		return nil
	}

	dstRV = dstRV.Elem()
	srcRV = srcRV.Elem()

	if dstRV.Kind() != reflect.Struct || srcRV.Kind() != reflect.Struct {
		return nil
	}

	for i := 0; i < srcRV.NumField(); i++ {
		srcField := srcRV.Field(i)
		dstField := dstRV.Field(i)

		if !dstField.CanSet() {
			continue
		}

		if !srcField.IsZero() {
			dstField.Set(srcField)
		}
	}

	return nil
}

// PickFields 从结构体中选择指定字段，返回字段名到字段值的映射
func PickFields(src any, fields ...string) map[string]any {
	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	result := make(map[string]any)
	for _, field := range fields {
		f := rv.FieldByName(field)
		if f.IsValid() {
			result[field] = f.Interface()
		}
	}
	return result
}

// OmitFields 从结构体中排除指定字段，返回剩余字段名到字段值的映射
func OmitFields(src any, fields ...string) map[string]any {
	rv := reflect.ValueOf(src)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return nil
	}

	rt := rv.Type()
	omitSet := make(map[string]struct{})
	for _, f := range fields {
		omitSet[f] = struct{}{}
	}

	result := make(map[string]any)
	for i := 0; i < rv.NumField(); i++ {
		field := rt.Field(i)
		if _, ok := omitSet[field.Name]; !ok {
			result[field.Name] = rv.Field(i).Interface()
		}
	}
	return result
}
