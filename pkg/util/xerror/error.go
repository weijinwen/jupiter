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
	"fmt"
	"runtime"
	"strings"
)

// Error represents an error with additional context.
type Error struct {
	Err     error  `json:"-"`
	Message string `json:"message"`
	Stack   string `json:"stack,omitempty"`
}

// New creates a new Error with a message.
func New(message string) *Error {
	return &Error{
		Message: message,
		Stack:   getStack(),
	}
}

// Newf creates a new Error with a formatted message.
func Newf(format string, args ...interface{}) *Error {
	return &Error{
		Message: fmt.Sprintf(format, args...),
		Stack:   getStack(),
	}
}

// Wrap wraps an error with a message.
func Wrap(err error, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Err:     err,
		Message: message,
		Stack:   getStack(),
	}
}

// Wrapf wraps an error with a formatted message.
func Wrapf(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Err:     err,
		Message: fmt.Sprintf(format, args...),
		Stack:   getStack(),
	}
}

// Error returns the error message.
func (e *Error) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

// Unwrap returns the wrapped error.
func (e *Error) Unwrap() error {
	return e.Err
}

// Cause returns the root cause error.
func (e *Error) Cause() error {
	if e.Err != nil {
		if cause, ok := e.Err.(*Error); ok {
			return cause.Cause()
		}
		return e.Err
	}
	return e
}

// Is checks if the target error is in the chain.
func (e *Error) Is(target error) bool {
	if e.Err != nil {
		if err, ok := e.Err.(*Error); ok {
			return err.Is(target)
		}
		return e.Err == target
	}
	return false
}

// As finds the first error in chain that matches target.
func (e *Error) As(target interface{}) bool {
	if e.Err != nil {
		if err, ok := e.Err.(*Error); ok {
			return err.As(target)
		}
		return e.Err == target
	}
	return false
}

// WithStack adds a stack trace to the error.
func WithStack(err error) *Error {
	if err == nil {
		return nil
	}
	if e, ok := err.(*Error); ok {
		return e
	}
	return &Error{
		Err:   err,
		Stack: getStack(),
	}
}

// WithMessage adds a message to the error.
func WithMessage(err error, message string) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Err:     err,
		Message: message,
		Stack:   getStack(),
	}
}

// WithMessagef adds a formatted message to the error.
func WithMessagef(err error, format string, args ...interface{}) *Error {
	if err == nil {
		return nil
	}
	return &Error{
		Err:     err,
		Message: fmt.Sprintf(format, args...),
		Stack:   getStack(),
	}
}

// getStack returns the stack trace as a string.
func getStack() string {
	var pcs [32]uintptr
	n := runtime.Callers(3, pcs[:])
	frames := runtime.CallersFrames(pcs[:n])

	var stack strings.Builder
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		stack.WriteString(fmt.Sprintf("%s\n\t%s:%d\n", frame.Function, frame.File, frame.Line))
	}
	return stack.String()
}

// Common errors
var (
	ErrNotFound     = New("not found")
	ErrInvalidInput = New("invalid input")
	ErrInternal     = New("internal error")
	ErrUnauthorized = New("unauthorized")
	ErrForbidden    = New("forbidden")
	ErrConflict     = New("conflict")
	ErrTimeout      = New("timeout")
	ErrClosed       = New("closed")
)

// NotFound creates a not found error.
func NotFound(resource string) *Error {
	return Newf("%s not found", resource)
}

// NotFoundf creates a not found error with formatted message.
func NotFoundf(format string, args ...interface{}) *Error {
	return Newf(format+" not found", args...)
}

// InvalidInput creates an invalid input error.
func InvalidInput(field, reason string) *Error {
	return Newf("invalid input: %s - %s", field, reason)
}

// InvalidInputf creates an invalid input error with formatted message.
func InvalidInputf(format string, args ...interface{}) *Error {
	return Newf("invalid input: "+format, args...)
}

// Internal creates an internal error.
func Internal(message string) *Error {
	return Newf("internal error: %s", message)
}

// Internalf creates an internal error with formatted message.
func Internalf(format string, args ...interface{}) *Error {
	return Newf("internal error: "+format, args...)
}

// Unauthorized creates an unauthorized error.
func Unauthorized(message string) *Error {
	return Newf("unauthorized: %s", message)
}

// Forbidden creates a forbidden error.
func Forbidden(message string) *Error {
	return Newf("forbidden: %s", message)
}

// Conflict creates a conflict error.
func Conflict(resource string) *Error {
	return Newf("%s conflict", resource)
}

// Timeout creates a timeout error.
func Timeout(operation string) *Error {
	return Newf("%s timeout", operation)
}

// IsNotFound checks if the error is a not found error.
func IsNotFound(err error) bool {
	return err == ErrNotFound || (err != nil && strings.Contains(err.Error(), "not found"))
}

// IsInvalidInput checks if the error is an invalid input error.
func IsInvalidInput(err error) bool {
	return err == ErrInvalidInput || (err != nil && strings.Contains(err.Error(), "invalid input"))
}

// IsInternal checks if the error is an internal error.
func IsInternal(err error) bool {
	return err == ErrInternal || (err != nil && strings.Contains(err.Error(), "internal error"))
}

// IsUnauthorized checks if the error is an unauthorized error.
func IsUnauthorized(err error) bool {
	return err == ErrUnauthorized || (err != nil && strings.Contains(err.Error(), "unauthorized"))
}

// IsForbidden checks if the error is a forbidden error.
func IsForbidden(err error) bool {
	return err == ErrForbidden || (err != nil && strings.Contains(err.Error(), "forbidden"))
}

// IsConflict checks if the error is a conflict error.
func IsConflict(err error) bool {
	return err == ErrConflict || (err != nil && strings.Contains(err.Error(), "conflict"))
}

// IsTimeout checks if the error is a timeout error.
func IsTimeout(err error) bool {
	return err == ErrTimeout || (err != nil && strings.Contains(err.Error(), "timeout"))
}

// IsClosed checks if the error is a closed error.
func IsClosed(err error) bool {
	return err == ErrClosed || (err != nil && strings.Contains(err.Error(), "closed"))
}
