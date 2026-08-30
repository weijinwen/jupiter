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

package xmath

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAbs(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{"positive int", 5, 5},
		{"negative int", -5, 5},
		{"zero int", 0, 0},
		{"positive float", 3.14, 3.14},
		{"negative float", -3.14, 3.14},
		{"zero float", 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.input.(type) {
			case int:
				assert.Equal(t, tt.expected, Abs(v))
			case float64:
				assert.Equal(t, tt.expected, Abs(v))
			}
		})
	}
}

func TestMax(t *testing.T) {
	assert.Equal(t, 10, Max(5, 10))
	assert.Equal(t, 10, Max(10, 5))
	assert.Equal(t, 5, Max(5, 5))
	assert.Equal(t, 3.14, Max(2.71, 3.14))
	assert.Equal(t, 2.71, Max(2.71, 1.5))
}

func TestMin(t *testing.T) {
	assert.Equal(t, 5, Min(5, 10))
	assert.Equal(t, 5, Min(10, 5))
	assert.Equal(t, 5, Min(5, 5))
	assert.Equal(t, 2.71, Min(2.71, 3.14))
	assert.Equal(t, 1.5, Min(2.71, 1.5))
}

func TestClamp(t *testing.T) {
	assert.Equal(t, 5, Clamp(3, 5, 10))
	assert.Equal(t, 7, Clamp(7, 5, 10))
	assert.Equal(t, 10, Clamp(15, 5, 10))
	assert.Equal(t, 5.0, Clamp(3.0, 5.0, 10.0))
	assert.Equal(t, 7.5, Clamp(7.5, 5.0, 10.0))
	assert.Equal(t, 10.0, Clamp(15.0, 5.0, 10.0))
}

func TestSum(t *testing.T) {
	assert.Equal(t, 15, Sum([]int{1, 2, 3, 4, 5}...))
	assert.Equal(t, 0, Sum([]int{}...))
	assert.Equal(t, 10, Sum([]int{10}...))
	assert.Equal(t, 9.0, Sum([]float64{1.5, 2.5, 5.0}...))
}

func TestAverage(t *testing.T) {
	assert.Equal(t, 3.0, Average([]int{1, 2, 3, 4, 5}...))
	assert.Equal(t, 0.0, Average([]int{}...))
	assert.Equal(t, 2.5, Average([]int{1, 2, 3, 4}...))
}

func TestMedian(t *testing.T) {
	assert.Equal(t, 3.0, Median([]int{1, 2, 3, 4, 5}...))
	assert.Equal(t, 2.5, Median([]int{1, 2, 3, 4}...))
	assert.Equal(t, 0.0, Median([]int{}...))
}

func TestMode(t *testing.T) {
	result := Mode(1, 2, 2, 3, 3, 3)
	assert.Contains(t, result, 3)
	assert.Nil(t, Mode(1, 2, 3, 4, 5))
}

func TestPow(t *testing.T) {
	assert.Equal(t, 8.0, Pow(2, 3))
	assert.Equal(t, 1.0, Pow(5, 0))
	assert.Equal(t, 0.5, Pow(2, -1))
}

func TestSqrt(t *testing.T) {
	assert.Equal(t, 3.0, Sqrt(9))
	assert.Equal(t, 0.0, Sqrt(0))
	assert.InDelta(t, 1.414, Sqrt(2), 0.001)
}

func TestLog(t *testing.T) {
	assert.Equal(t, 0.0, Log(1))
	assert.Equal(t, 1.0, Log(2.718281828459045))
	assert.InDelta(t, 2.0, Log10(100), 0.0001)
	assert.Equal(t, 3.0, Log2(8))
}

func TestExp(t *testing.T) {
	assert.Equal(t, 1.0, Exp(0))
	assert.InDelta(t, 2.718, Exp(1), 0.001)
}

func TestCeil(t *testing.T) {
	assert.Equal(t, 3.0, Ceil(2.1))
	assert.Equal(t, 3.0, Ceil(2.9))
	assert.Equal(t, -2.0, Ceil(-2.1))
}

func TestFloor(t *testing.T) {
	assert.Equal(t, 2.0, Floor(2.1))
	assert.Equal(t, 2.0, Floor(2.9))
	assert.Equal(t, -3.0, Floor(-2.1))
}

func TestRound(t *testing.T) {
	assert.Equal(t, 3.0, Round(2.5))
	assert.Equal(t, 2.0, Round(2.4))
	assert.Equal(t, -2.0, Round(-2.4))
	assert.Equal(t, -3.0, Round(-2.5))
}

func TestSign(t *testing.T) {
	assert.Equal(t, -1, Sign(-5))
	assert.Equal(t, 0, Sign(0))
	assert.Equal(t, 1, Sign(5))
	assert.Equal(t, -1, Sign(-3.14))
	assert.Equal(t, 0, Sign(0.0))
	assert.Equal(t, 1, Sign(3.14))
}

func TestIsNaN(t *testing.T) {
	assert.True(t, IsNaN(math.NaN()))
	assert.False(t, IsNaN(1.0))
	assert.False(t, IsNaN(0.0))
}

func TestIsInf(t *testing.T) {
	assert.True(t, IsInf(math.Inf(1), 1))
	assert.True(t, IsInf(math.Inf(-1), -1))
	assert.False(t, IsInf(1.0, 0))
}

func TestIsFinite(t *testing.T) {
	assert.True(t, IsFinite(1.0))
	assert.True(t, IsFinite(0.0))
	assert.False(t, IsFinite(math.Inf(1)))
	assert.False(t, IsFinite(math.NaN()))
}

func TestLerp(t *testing.T) {
	assert.Equal(t, 2.5, Lerp(1.0, 4.0, 0.5))
	assert.Equal(t, 1.0, Lerp(1.0, 4.0, 0.0))
	assert.Equal(t, 4.0, Lerp(1.0, 4.0, 1.0))
}

func TestMap(t *testing.T) {
	assert.Equal(t, 50.0, Map(5, 0, 10, 0, 100))
	assert.Equal(t, 0.0, Map(0, 0, 10, 0, 100))
	assert.Equal(t, 100.0, Map(10, 0, 10, 0, 100))
}

func TestGCD(t *testing.T) {
	assert.Equal(t, 3, GCD(12, 9))
	assert.Equal(t, 1, GCD(7, 13))
	assert.Equal(t, 5, GCD(10, 15))
}

func TestLCM(t *testing.T) {
	assert.Equal(t, 12, LCM(4, 6))
	assert.Equal(t, 21, LCM(3, 7))
	assert.Equal(t, 0, LCM(0, 5))
}

func TestFactorial(t *testing.T) {
	assert.Equal(t, uint64(1), Factorial(0))
	assert.Equal(t, uint64(1), Factorial(1))
	assert.Equal(t, uint64(120), Factorial(5))
	assert.Equal(t, uint64(3628800), Factorial(10))
}

func TestFibonacci(t *testing.T) {
	assert.Equal(t, uint64(0), Fibonacci(0))
	assert.Equal(t, uint64(1), Fibonacci(1))
	assert.Equal(t, uint64(1), Fibonacci(2))
	assert.Equal(t, uint64(5), Fibonacci(5))
	assert.Equal(t, uint64(55), Fibonacci(10))
}

func TestIsPrime(t *testing.T) {
	assert.False(t, IsPrime(-1))
	assert.False(t, IsPrime(0))
	assert.False(t, IsPrime(1))
	assert.True(t, IsPrime(2))
	assert.True(t, IsPrime(3))
	assert.False(t, IsPrime(4))
	assert.True(t, IsPrime(5))
	assert.True(t, IsPrime(7))
	assert.False(t, IsPrime(9))
	assert.True(t, IsPrime(11))
}

func TestFactors(t *testing.T) {
	assert.Equal(t, []int{1, 12, 2, 6, 3, 4}, Factors(12))
	assert.Equal(t, []int{1, 7}, Factors(7))
	assert.Equal(t, []int{1}, Factors(1))
}

func TestIsEven(t *testing.T) {
	assert.True(t, IsEven(2))
	assert.True(t, IsEven(4))
	assert.False(t, IsEven(3))
	assert.False(t, IsEven(5))
}

func TestIsOdd(t *testing.T) {
	assert.False(t, IsOdd(2))
	assert.False(t, IsOdd(4))
	assert.True(t, IsOdd(3))
	assert.True(t, IsOdd(5))
}

func TestDigits(t *testing.T) {
	assert.Equal(t, 1, Digits(0))
	assert.Equal(t, 1, Digits(5))
	assert.Equal(t, 2, Digits(12))
	assert.Equal(t, 3, Digits(123))
	assert.Equal(t, 4, Digits(1234))
}

func TestReverseDigits(t *testing.T) {
	assert.Equal(t, 54321, ReverseDigits(12345))
	assert.Equal(t, 1, ReverseDigits(1))
	assert.Equal(t, 0, ReverseDigits(0))
	assert.Equal(t, -54321, ReverseDigits(-12345))
}

func TestIsPalindrome(t *testing.T) {
	assert.True(t, IsPalindrome(121))
	assert.True(t, IsPalindrome(1221))
	assert.False(t, IsPalindrome(123))
	assert.True(t, IsPalindrome(0))
	assert.False(t, IsPalindrome(-121))
}
