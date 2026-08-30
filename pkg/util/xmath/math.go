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
)

// Abs returns the absolute value of x.
func Abs[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

// Max returns the larger of x or y.
func Max[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](x, y T) T {
	if x > y {
		return x
	}
	return y
}

// Min returns the smaller of x or y.
func Min[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](x, y T) T {
	if x < y {
		return x
	}
	return y
}

// Clamp clamps the value to the specified range [min, max].
func Clamp[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](value, min, max T) T {
	if value < min {
		return min
	}
	if value > max {
		return max
	}
	return value
}

// Sum returns the sum of the given numbers.
func Sum[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](numbers ...T) T {
	var sum T
	for _, num := range numbers {
		sum += num
	}
	return sum
}

// Average returns the average of the given numbers.
func Average[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](numbers ...T) float64 {
	if len(numbers) == 0 {
		return 0
	}
	var sum float64
	for _, num := range numbers {
		sum += float64(num)
	}
	return sum / float64(len(numbers))
}

// Median returns the median of the given numbers.
func Median[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~float32 | ~float64](numbers ...T) float64 {
	if len(numbers) == 0 {
		return 0
	}

	// Create a copy to avoid modifying the original slice
	sorted := make([]float64, len(numbers))
	for i, num := range numbers {
		sorted[i] = float64(num)
	}

	// Simple bubble sort (for small slices)
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	mid := len(sorted) / 2
	if len(sorted)%2 == 0 {
		return (sorted[mid-1] + sorted[mid]) / 2
	}
	return sorted[mid]
}

// Mode returns the mode(s) of the given numbers.
func Mode[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64](numbers ...T) []T {
	if len(numbers) == 0 {
		return nil
	}

	freq := make(map[T]int)
	for _, num := range numbers {
		freq[num]++
	}

	maxFreq := 0
	for _, count := range freq {
		if count > maxFreq {
			maxFreq = count
		}
	}

	if maxFreq == 1 {
		return nil
	}

	var modes []T
	for num, count := range freq {
		if count == maxFreq {
			modes = append(modes, num)
		}
	}
	return modes
}

// Pow returns x raised to the power of y.
func Pow[T ~int | ~float32 | ~float64](x, y T) float64 {
	return math.Pow(float64(x), float64(y))
}

// Sqrt returns the square root of x.
func Sqrt[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Sqrt(float64(x))
}

// Cbrt returns the cube root of x.
func Cbrt[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Cbrt(float64(x))
}

// Log returns the natural logarithm of x.
func Log[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Log(float64(x))
}

// Log10 returns the base-10 logarithm of x.
func Log10[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Log10(float64(x))
}

// Log2 returns the base-2 logarithm of x.
func Log2[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Log2(float64(x))
}

// Exp returns e raised to the power of x.
func Exp[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Exp(float64(x))
}

// Ceil returns the least integer value greater than or equal to x.
func Ceil[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Ceil(float64(x))
}

// Floor returns the greatest integer value less than or equal to x.
func Floor[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Floor(float64(x))
}

// Round returns the nearest integer, rounding half away from zero.
func Round[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Round(float64(x))
}

// Trunc returns the integer value of x.
func Trunc[T ~int | ~float32 | ~float64](x T) float64 {
	return math.Trunc(float64(x))
}

// Sign returns:
//
//	-1 if x < 0
//	 0 if x == 0
//	+1 if x > 0
func Sign[T ~int | ~int8 | ~int16 | ~int32 | ~int64 | ~float32 | ~float64](x T) int {
	if x < 0 {
		return -1
	}
	if x > 0 {
		return 1
	}
	return 0
}

// IsNaN reports whether x is an IEEE 754 "not-a-number" value.
func IsNaN[T ~float32 | ~float64](x T) bool {
	return math.IsNaN(float64(x))
}

// IsInf reports whether x is positive or negative infinity.
func IsInf[T ~float32 | ~float64](x T, sign int) bool {
	return math.IsInf(float64(x), sign)
}

// IsFinite reports whether x is a finite number.
func IsFinite[T ~float32 | ~float64](x T) bool {
	return !math.IsInf(float64(x), 0) && !math.IsNaN(float64(x))
}

// Lerp returns the linear interpolation between x and y at the point t (0-1).
func Lerp[T ~int | ~float32 | ~float64](x, y T, t float64) float64 {
	return float64(x) + t*float64(y-x)
}

// Map maps a value from one range to another.
func Map[T ~int | ~float32 | ~float64](value, inMin, inMax, outMin, outMax T) float64 {
	return float64(value-inMin)*float64(outMax-outMin)/float64(inMax-inMin) + float64(outMin)
}

// GCD returns the greatest common divisor of a and b.
func GCD[T ~int | ~int8 | ~int16 | ~int32 | ~int64](a, b T) T {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// LCM returns the least common multiple of a and b.
func LCM[T ~int | ~int8 | ~int16 | ~int32 | ~int64](a, b T) T {
	if a == 0 || b == 0 {
		return 0
	}
	return Abs(a*b) / GCD(a, b)
}

// Factorial returns the factorial of n.
func Factorial[T ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) uint64 {
	if n < 0 {
		return 0
	}
	if n <= 1 {
		return 1
	}
	var result uint64 = 1
	for i := T(2); i <= n; i++ {
		result *= uint64(i)
	}
	return result
}

// Fibonacci returns the nth Fibonacci number.
func Fibonacci[T ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) uint64 {
	if n <= 0 {
		return 0
	}
	if n == 1 {
		return 1
	}
	var a, b uint64 = 0, 1
	for i := T(2); i <= n; i++ {
		a, b = b, a+b
	}
	return b
}

// IsPrime reports whether n is a prime number.
func IsPrime[T ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) bool {
	if n <= 1 {
		return false
	}
	if n <= 3 {
		return true
	}
	if n%2 == 0 || n%3 == 0 {
		return false
	}
	for i := T(5); i*i <= n; i += 6 {
		if n%i == 0 || n%(i+2) == 0 {
			return false
		}
	}
	return true
}

// Factors returns all factors of n.
func Factors[T ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) []T {
	var factors []T
	for i := T(1); i*i <= n; i++ {
		if n%i == 0 {
			factors = append(factors, i)
			if i != n/i {
				factors = append(factors, n/i)
			}
		}
	}
	return factors
}

// IsEven reports whether n is even.
func IsEven[T ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) bool {
	return n%2 == 0
}

// IsOdd reports whether n is odd.
func IsOdd[T ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) bool {
	return n%2 != 0
}

// Digits returns the number of digits in n.
func Digits[T ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) int {
	if n == 0 {
		return 1
	}
	count := 0
	n = Abs(n)
	for n > 0 {
		n /= 10
		count++
	}
	return count
}

// ReverseDigits reverses the digits of n.
func ReverseDigits[T ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) T {
	var result T
	for n != 0 {
		result = result*10 + n%10
		n /= 10
	}
	return result
}

// IsPalindrome reports whether n is a palindrome number.
func IsPalindrome[T ~int | ~int8 | ~int16 | ~int32 | ~int64](n T) bool {
	return n >= 0 && n == ReverseDigits(n)
}
