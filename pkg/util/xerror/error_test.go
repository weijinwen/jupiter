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

package xerror

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	err := New("test error")
	assert.Equal(t, "test error", err.Error())
	assert.Contains(t, err.Stack, "xerror.TestNew")
}

func TestNewf(t *testing.T) {
	err := Newf("error %s %d", "code", 42)
	assert.Equal(t, "error code 42", err.Error())
}

func TestWrap(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrap(original, "wrapped")
	assert.Equal(t, "wrapped: original error", wrapped.Error())
	assert.Equal(t, original, wrapped.Unwrap())
}

func TestWrapNil(t *testing.T) {
	wrapped := Wrap(nil, "wrapped")
	assert.Nil(t, wrapped)
}

func TestWrapf(t *testing.T) {
	original := errors.New("original error")
	wrapped := Wrapf(original, "wrapped %s", "message")
	assert.Equal(t, "wrapped message: original error", wrapped.Error())
}

func TestCause(t *testing.T) {
	inner := errors.New("inner")
	middle := Wrap(inner, "middle")
	outer := Wrap(middle, "outer")

	assert.Equal(t, inner, outer.Cause())
}

func TestWithStack(t *testing.T) {
	original := errors.New("original")
	withStack := WithStack(original)
	assert.Equal(t, original, withStack.Err)
	assert.Contains(t, withStack.Stack, "xerror.TestWithStack")
}

func TestWithMessage(t *testing.T) {
	original := errors.New("original")
	withMessage := WithMessage(original, "additional context")
	assert.Contains(t, withMessage.Error(), "additional context")
	assert.Contains(t, withMessage.Error(), "original")
}

func TestWithMessagef(t *testing.T) {
	original := errors.New("original")
	withMessage := WithMessagef(original, "code: %d", 500)
	assert.Contains(t, withMessage.Error(), "code: 500")
}

func TestNotFound(t *testing.T) {
	err := NotFound("user")
	assert.True(t, IsNotFound(err))
	assert.Contains(t, err.Error(), "not found")
}

func TestInvalidInput(t *testing.T) {
	err := InvalidInput("email", "invalid format")
	assert.True(t, IsInvalidInput(err))
	assert.Contains(t, err.Error(), "invalid input")
}

func TestInternal(t *testing.T) {
	err := Internal("database connection failed")
	assert.True(t, IsInternal(err))
	assert.Contains(t, err.Error(), "internal error")
}

func TestUnauthorized(t *testing.T) {
	err := Unauthorized("missing token")
	assert.True(t, IsUnauthorized(err))
	assert.Contains(t, err.Error(), "unauthorized")
}

func TestForbidden(t *testing.T) {
	err := Forbidden("insufficient permissions")
	assert.True(t, IsForbidden(err))
	assert.Contains(t, err.Error(), "forbidden")
}

func TestConflict(t *testing.T) {
	err := Conflict("resource")
	assert.True(t, IsConflict(err))
	assert.Contains(t, err.Error(), "conflict")
}

func TestTimeout(t *testing.T) {
	err := Timeout("operation")
	assert.True(t, IsTimeout(err))
	assert.Contains(t, err.Error(), "timeout")
}

func TestIsClosed(t *testing.T) {
	err := errors.New("connection closed")
	assert.True(t, IsClosed(err))
}

func TestErrorChain(t *testing.T) {
	err1 := errors.New("first")
	err2 := Wrap(err1, "second")
	err3 := Wrap(err2, "third")

	assert.Contains(t, err3.Error(), "third")
	assert.Contains(t, err3.Error(), "second")
	assert.Contains(t, err3.Error(), "first")
	assert.Equal(t, err1, err3.Cause())
}

func TestIsChecks(t *testing.T) {
	// Test IsNotFound
	err := NotFound("test")
	assert.True(t, IsNotFound(err))
	assert.False(t, IsInvalidInput(err))
	assert.False(t, IsInternal(err))
	assert.False(t, IsUnauthorized(err))
	assert.False(t, IsForbidden(err))
	assert.False(t, IsConflict(err))
	assert.False(t, IsTimeout(err))

	// Test plain error messages
	assert.True(t, IsNotFound(errors.New("not found error")))
	assert.False(t, IsNotFound(errors.New("other error")))
}

func TestIsMethod(t *testing.T) {
	err1 := errors.New("test")
	err2 := Wrap(err1, "wrapped")
	assert.True(t, err2.Is(err1))
}

func TestAsMethod(t *testing.T) {
	err1 := errors.New("test")
	err2 := Wrap(err1, "wrapped")
	// Test that we can extract the error
	assert.Equal(t, err1, err2.Unwrap())
	assert.Equal(t, "wrapped", err2.Message)
}
