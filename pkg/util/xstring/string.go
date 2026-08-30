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

package xstring

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"net"
	"strconv"
	"strings"
)

// Addr2Hex converts address string to hex string, only support ipv4.
func Addr2Hex(str string) (string, error) {
	ipStr, portStr, err := net.SplitHostPort(str)
	if err != nil {
		return "", err
	}

	ip := net.ParseIP(ipStr).To4()
	port, err := strconv.ParseUint(portStr, 10, 16)
	if err != nil {
		return "", nil
	}

	buf := make([]byte, 2)
	binary.BigEndian.PutUint16(buf, uint16(port))
	ip = append(ip, buf...)

	return hex.EncodeToString(ip), nil
}

// Hex2Addr converts hex string to address.
func Hex2Addr(str string) (string, error) {
	buf, err := hex.DecodeString(str)
	if err != nil {
		return "", err
	}
	if len(buf) < 4 {
		return "", fmt.Errorf("bad hex string length")
	}
	return fmt.Sprintf("%s:%d", net.IP(buf[:4]).String(), binary.BigEndian.Uint16(buf[4:])), nil
}

// Strings ...
type Strings []string

// KickEmpty kick empty elements from ss
func KickEmpty(ss []string) Strings {
	var ret = make([]string, 0)
	for _, str := range ss {
		if str != "" {
			ret = append(ret, str)
		}
	}
	return Strings(ret)
}

func Kick(ss []string, remove func(item string) bool) Strings {
	for i := 0; i < len(ss); i++ {
		if remove(ss[i]) {
			ss = append(ss[:i], ss[i+1:]...)
			i--
		}
	}
	return ss
}

// AnyBlank return true if ss has empty element
func AnyBlank(ss []string) bool {
	for _, str := range ss {
		if str == "" {
			return true
		}
	}

	return false
}

// Split splits the string s around all instances of sep.
func Split(s, sep string) []string {
	if s == "" {
		return []string{}
	}
	return strings.Split(s, sep)
}

// SplitN splits the string s around the first n instances of sep.
func SplitN(s, sep string, n int) []string {
	if s == "" {
		return []string{}
	}
	return strings.SplitN(s, sep, n)
}

// Join concatenates the elements of slice to create a single string.
func Join(slice []string, sep string) string {
	return strings.Join(slice, sep)
}

// Repeat returns a string consisting of count copies of the string s.
func Repeat(s string, count int) string {
	return strings.Repeat(s, count)
}

// Contains reports whether substr is within s.
func Contains(s, substr string) bool {
	return strings.Contains(s, substr)
}

// ContainsAny reports whether any Unicode code points in chars are within s.
func ContainsAny(s, chars string) bool {
	return strings.ContainsAny(s, chars)
}

// ContainsRune reports whether the Unicode code point r is within s.
func ContainsRune(s string, r rune) bool {
	return strings.ContainsRune(s, r)
}

// HasPrefix tests whether the string s begins with prefix.
func HasPrefix(s, prefix string) bool {
	return strings.HasPrefix(s, prefix)
}

// HasSuffix tests whether the string s ends with suffix.
func HasSuffix(s, suffix string) bool {
	return strings.HasSuffix(s, suffix)
}

// Index returns the index of the first instance of substr in s, or -1 if not present.
func Index(s, substr string) int {
	return strings.Index(s, substr)
}

// IndexAny returns the index of the first instance of any Unicode code point from chars in s, or -1 if not present.
func IndexAny(s, chars string) int {
	return strings.IndexAny(s, chars)
}

// IndexRune returns the index of the first instance of the Unicode code point r, or -1 if not present.
func IndexRune(s string, r rune) int {
	return strings.IndexRune(s, r)
}

// LastIndex returns the index of the last instance of substr in s, or -1 if not present.
func LastIndex(s, substr string) int {
	return strings.LastIndex(s, substr)
}

// LastIndexAny returns the index of the last instance of any Unicode code point from chars in s, or -1 if not present.
func LastIndexAny(s, chars string) int {
	return strings.LastIndexAny(s, chars)
}

// Count counts the number of non-overlapping instances of substr in s.
func Count(s, substr string) int {
	return strings.Count(s, substr)
}

// ToUpper returns a copy of the string s with all Unicode letters mapped to their upper case.
func ToUpper(s string) string {
	return strings.ToUpper(s)
}

// ToLower returns a copy of the string s with all Unicode letters mapped to their lower case.
func ToLower(s string) string {
	return strings.ToLower(s)
}

// ToTitle returns a copy of the string s with all Unicode letters mapped to their title case.
func ToTitle(s string) string {
	return strings.Title(s)
}

// Trim returns a copy of the string s with all leading and trailing Unicode code points contained in cutset removed.
func Trim(s string, cutset string) string {
	return strings.Trim(s, cutset)
}

// TrimLeft returns a copy of the string s with all leading Unicode code points contained in cutset removed.
func TrimLeft(s string, cutset string) string {
	return strings.TrimLeft(s, cutset)
}

// TrimRight returns a copy of the string s with all trailing Unicode code points contained in cutset removed.
func TrimRight(s string, cutset string) string {
	return strings.TrimRight(s, cutset)
}

// TrimSpace returns a copy of the string s with all leading and trailing white space removed.
func TrimSpace(s string) string {
	return strings.TrimSpace(s)
}

// Replace returns a copy of the string s with the first n instances of old replaced by new.
func Replace(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

// ReplaceAll returns a copy of the string s with all instances of old replaced by new.
func ReplaceAll(s, old, new string) string {
	return strings.ReplaceAll(s, old, new)
}

// EqualFold reports whether s and t, interpreted as UTF-8 strings, are equal under Unicode case-folding.
func EqualFold(s, t string) bool {
	return strings.EqualFold(s, t)
}

// IsBlank checks if a string is blank (empty or contains only whitespace).
func IsBlank(s string) bool {
	return strings.TrimSpace(s) == ""
}

// IsNotBlank checks if a string is not blank.
func IsNotBlank(s string) bool {
	return !IsBlank(s)
}

// DefaultIfBlank returns the default value if the string is blank.
func DefaultIfBlank(s, defaultValue string) string {
	if IsBlank(s) {
		return defaultValue
	}
	return s
}

// Truncate truncates a string to the specified length, adding suffix if truncated.
func Truncate(s string, length int, suffix string) string {
	if len(s) <= length {
		return s
	}
	if length <= len(suffix) {
		return suffix[:length]
	}
	return s[:length-len(suffix)] + suffix
}

// Wrap wraps a string with a prefix and suffix.
func Wrap(s, wrap string) string {
	return wrap + s + wrap
}

// WrapIfNotEmpty wraps a string with a prefix and suffix if it's not empty.
func WrapIfNotEmpty(s, wrap string) string {
	if IsBlank(s) {
		return s
	}
	return wrap + s + wrap
}

// Unwrap removes the first occurrence of the prefix and suffix from the string.
func Unwrap(s, wrap string) string {
	if HasPrefix(s, wrap) && HasSuffix(s, wrap) {
		return s[len(wrap) : len(s)-len(wrap)]
	}
	return s
}

// Reverse reverses the string.
func Reverse(s string) string {
	runes := []rune(s)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

// PadLeft pads a string on the left side to the specified length.
func PadLeft(s string, length int, padStr string) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(padStr, (length-len(s))/len(padStr)+1)
	return padding[:length-len(s)] + s
}

// PadRight pads a string on the right side to the specified length.
func PadRight(s string, length int, padStr string) string {
	if len(s) >= length {
		return s
	}
	padding := strings.Repeat(padStr, (length-len(s))/len(padStr)+1)
	return s + padding[:length-len(s)]
}

// Center centers a string within the specified length.
func Center(s string, length int, padStr string) string {
	if len(s) >= length {
		return s
	}
	leftPad := (length - len(s)) / 2
	rightPad := length - len(s) - leftPad
	return strings.Repeat(padStr, leftPad/len(padStr)+1)[:leftPad] +
		s +
		strings.Repeat(padStr, rightPad/len(padStr)+1)[:rightPad]
}
