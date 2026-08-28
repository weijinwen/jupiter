package xstruct

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type TestPerson struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
	City string `json:"city"`
}

func TestCopy(t *testing.T) {
	src := TestPerson{Name: "John", Age: 30, City: "NYC"}
	var dst TestPerson
	err := Copy(src, &dst)
	assert.NoError(t, err)
	assert.Equal(t, src, dst)
}

func TestClone(t *testing.T) {
	src := TestPerson{Name: "John", Age: 30, City: "NYC"}
	dst, err := Clone(src)
	assert.NoError(t, err)
	assert.Equal(t, src, dst)

	dst.Name = "Jane"
	assert.NotEqual(t, src.Name, dst.Name)
}

func TestGetField(t *testing.T) {
	p := TestPerson{Name: "John", Age: 30, City: "NYC"}

	name, err := GetField(p, "Name")
	assert.NoError(t, err)
	assert.Equal(t, "John", name)

	age, err := GetField(p, "Age")
	assert.NoError(t, err)
	assert.Equal(t, 30, age)

	_, err = GetField(p, "Invalid")
	assert.NoError(t, err)
}

func TestSetField(t *testing.T) {
	p := &TestPerson{Name: "John", Age: 30}

	err := SetField(p, "Name", "Jane")
	assert.NoError(t, err)
	assert.Equal(t, "Jane", p.Name)

	err = SetField(p, "Age", 25)
	assert.NoError(t, err)
	assert.Equal(t, 25, p.Age)
}

func TestGetFieldNames(t *testing.T) {
	p := TestPerson{}
	names := GetFieldNames(p)
	assert.ElementsMatch(t, []string{"Name", "Age", "City"}, names)
}

func TestGetFieldValues(t *testing.T) {
	p := TestPerson{Name: "John", Age: 30, City: "NYC"}
	values := GetFieldValues(p)
	assert.Len(t, values, 3)
	assert.Contains(t, values, "John")
	assert.Contains(t, values, 30)
	assert.Contains(t, values, "NYC")
}

func TestGetFieldMap(t *testing.T) {
	p := TestPerson{Name: "John", Age: 30, City: "NYC"}
	m := GetFieldMap(p)
	assert.Equal(t, "John", m["Name"])
	assert.Equal(t, 30, m["Age"])
	assert.Equal(t, "NYC", m["City"])
}

func TestToMap(t *testing.T) {
	p := TestPerson{Name: "John", Age: 30, City: "NYC"}
	m := ToMap(p)
	assert.Equal(t, "John", m["name"])
	assert.Equal(t, 30, m["age"])
	assert.Equal(t, "NYC", m["city"])
}

func TestMapToStruct(t *testing.T) {
	m := map[string]any{
		"name": "John",
		"age":  30,
		"city": "NYC",
	}

	var p TestPerson
	err := MapToStruct(m, &p)
	assert.NoError(t, err)
	assert.Equal(t, "John", p.Name)
	assert.Equal(t, 30, p.Age)
	assert.Equal(t, "NYC", p.City)
}

func TestHasField(t *testing.T) {
	p := TestPerson{}
	assert.True(t, HasField(p, "Name"))
	assert.True(t, HasField(p, "Age"))
	assert.False(t, HasField(p, "Invalid"))
}

func TestIsStruct(t *testing.T) {
	p := TestPerson{}
	assert.True(t, IsStruct(p))
	assert.True(t, IsStruct(&p))
	assert.False(t, IsStruct(123))
	assert.False(t, IsStruct("abc"))
}

func TestIsSlice(t *testing.T) {
	assert.True(t, IsSlice([]int{1, 2, 3}))
	assert.False(t, IsSlice(123))
	assert.False(t, IsSlice(map[string]int{}))
}

func TestIsMap(t *testing.T) {
	assert.True(t, IsMap(map[string]int{}))
	assert.False(t, IsMap(123))
	assert.False(t, IsMap([]int{}))
}

func TestTypeName(t *testing.T) {
	p := TestPerson{}
	assert.Equal(t, "TestPerson", TypeName(p))
	assert.Equal(t, "TestPerson", TypeName(&p))
	assert.Equal(t, "int", TypeName(123))
}

func TestMerge(t *testing.T) {
	dst := &TestPerson{Name: "John", Age: 30}
	src := &TestPerson{Age: 25, City: "NYC"}

	err := Merge(dst, src)
	assert.NoError(t, err)
	assert.Equal(t, "John", dst.Name)
	assert.Equal(t, 25, dst.Age)
	assert.Equal(t, "NYC", dst.City)
}

func TestPickFields(t *testing.T) {
	p := TestPerson{Name: "John", Age: 30, City: "NYC"}
	result := PickFields(p, "Name", "Age")
	assert.Len(t, result, 2)
	assert.Equal(t, "John", result["Name"])
	assert.Equal(t, 30, result["Age"])
	assert.NotContains(t, result, "City")
}

func TestOmitFields(t *testing.T) {
	p := TestPerson{Name: "John", Age: 30, City: "NYC"}
	result := OmitFields(p, "City")
	assert.Len(t, result, 2)
	assert.Contains(t, result, "Name")
	assert.Contains(t, result, "Age")
	assert.NotContains(t, result, "City")
}

func TestMustClone(t *testing.T) {
	src := TestPerson{Name: "John", Age: 30}
	dst := MustClone(src)
	assert.Equal(t, src, dst)
}
