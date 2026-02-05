package main

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestPointer(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "string value",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "int value",
			input:    42,
			expected: 42,
		},
		{
			name:     "bool value",
			input:    true,
			expected: true,
		},
		{
			name:     "zero string",
			input:    "",
			expected: "",
		},
		{
			name:     "zero int",
			input:    0,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.input.(type) {
			case string:
				result := Pointer(v)
				if result == nil {
					t.Errorf("Pointer() returned nil")
					return
				}
				if *result != tt.expected {
					t.Errorf("Pointer() = %v, want %v", *result, tt.expected)
				}
			case int:
				result := Pointer(v)
				if result == nil {
					t.Errorf("Pointer() returned nil")
					return
				}
				if *result != tt.expected {
					t.Errorf("Pointer() = %v, want %v", *result, tt.expected)
				}
			case bool:
				result := Pointer(v)
				if result == nil {
					t.Errorf("Pointer() returned nil")
					return
				}
				if *result != tt.expected {
					t.Errorf("Pointer() = %v, want %v", *result, tt.expected)
				}
			}
		})
	}
}

func TestValue(t *testing.T) {
	tests := []struct {
		name     string
		input    interface{}
		expected interface{}
	}{
		{
			name:     "string pointer",
			input:    Pointer("hello"),
			expected: "hello",
		},
		{
			name:     "int pointer",
			input:    Pointer(42),
			expected: 42,
		},
		{
			name:     "bool pointer",
			input:    Pointer(true),
			expected: true,
		},
		{
			name:     "nil string pointer",
			input:    (*string)(nil),
			expected: "",
		},
		{
			name:     "nil int pointer",
			input:    (*int)(nil),
			expected: 0,
		},
		{
			name:     "nil bool pointer",
			input:    (*bool)(nil),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch v := tt.input.(type) {
			case *string:
				result := Value(v)
				if result != tt.expected {
					t.Errorf("Value() = %v, want %v", result, tt.expected)
				}
			case *int:
				result := Value(v)
				if result != tt.expected {
					t.Errorf("Value() = %v, want %v", result, tt.expected)
				}
			case *bool:
				result := Value(v)
				if result != tt.expected {
					t.Errorf("Value() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestFilter(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		fn       func(string) bool
		expected []string
	}{
		{
			name:     "filter empty slice",
			input:    []string{},
			fn:       func(s string) bool { return len(s) > 3 },
			expected: []string{},
		},
		{
			name:     "filter strings by length",
			input:    []string{"a", "hello", "hi", "world"},
			fn:       func(s string) bool { return len(s) > 3 },
			expected: []string{"hello", "world"},
		},
		{
			name:     "filter strings by prefix",
			input:    []string{"apple", "banana", "apricot", "cherry"},
			fn:       func(s string) bool { return strings.HasPrefix(s, "ap") },
			expected: []string{"apple", "apricot"},
		},
		{
			name:     "filter all match",
			input:    []string{"test1", "test2", "test3"},
			fn:       func(s string) bool { return strings.HasPrefix(s, "test") },
			expected: []string{"test1", "test2", "test3"},
		},
		{
			name:     "filter none match",
			input:    []string{"apple", "banana", "cherry"},
			fn:       func(s string) bool { return strings.HasPrefix(s, "xyz") },
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.input, tt.fn)
			if len(result) != len(tt.expected) {
				t.Errorf("Filter() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Filter() = %v, want %v", result, tt.expected)
					break
				}
			}
		})
	}
}

func TestFilterInts(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) bool
		expected []int
	}{
		{
			name:     "filter even numbers",
			input:    []int{1, 2, 3, 4, 5, 6},
			fn:       func(n int) bool { return n%2 == 0 },
			expected: []int{2, 4, 6},
		},
		{
			name:     "filter numbers greater than 5",
			input:    []int{1, 3, 5, 7, 9, 11},
			fn:       func(n int) bool { return n > 5 },
			expected: []int{7, 9, 11},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Filter(tt.input, tt.fn)
			if len(result) != len(tt.expected) {
				t.Errorf("Filter() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Filter() = %v, want %v", result, tt.expected)
					break
				}
			}
		})
	}
}

func TestMap(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		fn       func(string) string
		expected []string
	}{
		{
			name:     "map empty slice",
			input:    []string{},
			fn:       strings.ToUpper,
			expected: []string{},
		},
		{
			name:     "map to uppercase",
			input:    []string{"hello", "world", "test"},
			fn:       strings.ToUpper,
			expected: []string{"HELLO", "WORLD", "TEST"},
		},
		{
			name:     "map add prefix",
			input:    []string{"apple", "banana", "cherry"},
			fn:       func(s string) string { return "fruit:" + s },
			expected: []string{"fruit:apple", "fruit:banana", "fruit:cherry"},
		},
		{
			name:     "map get length as string",
			input:    []string{"a", "hello", "world"},
			fn:       func(s string) string { return string(rune(len(s) + '0')) },
			expected: []string{"1", "5", "5"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.input, tt.fn)
			if len(result) != len(tt.expected) {
				t.Errorf("Map() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Map() = %v, want %v", result, tt.expected)
					break
				}
			}
		})
	}
}

func TestMapInts(t *testing.T) {
	tests := []struct {
		name     string
		input    []int
		fn       func(int) int
		expected []int
	}{
		{
			name:     "map multiply by 2",
			input:    []int{1, 2, 3, 4, 5},
			fn:       func(n int) int { return n * 2 },
			expected: []int{2, 4, 6, 8, 10},
		},
		{
			name:     "map square numbers",
			input:    []int{1, 2, 3, 4},
			fn:       func(n int) int { return n * n },
			expected: []int{1, 4, 9, 16},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Map(tt.input, tt.fn)
			if len(result) != len(tt.expected) {
				t.Errorf("Map() length = %v, want %v", len(result), len(tt.expected))
				return
			}
			for i, v := range result {
				if v != tt.expected[i] {
					t.Errorf("Map() = %v, want %v", result, tt.expected)
					break
				}
			}
		})
	}
}

func TestContains(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		value    string
		expected bool
	}{
		{
			name:     "contains in string slice",
			slice:    []string{"apple", "banana", "cherry"},
			value:    "banana",
			expected: true,
		},
		{
			name:     "does not contain in string slice",
			slice:    []string{"apple", "banana", "cherry"},
			value:    "grape",
			expected: false,
		},
		{
			name:     "empty slice",
			slice:    []string{},
			value:    "test",
			expected: false,
		},
		{
			name:     "contains empty string",
			slice:    []string{"", "test", "hello"},
			value:    "",
			expected: true,
		},
		{
			name:     "single element match",
			slice:    []string{"only"},
			value:    "only",
			expected: true,
		},
		{
			name:     "single element no match",
			slice:    []string{"only"},
			value:    "other",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.slice, tt.value)
			if result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestContainsInts(t *testing.T) {
	tests := []struct {
		name     string
		slice    []int
		value    int
		expected bool
	}{
		{
			name:     "contains in int slice",
			slice:    []int{1, 2, 3, 4, 5},
			value:    3,
			expected: true,
		},
		{
			name:     "does not contain in int slice",
			slice:    []int{1, 2, 3, 4, 5},
			value:    6,
			expected: false,
		},
		{
			name:     "contains zero",
			slice:    []int{0, 1, 2},
			value:    0,
			expected: true,
		},
		{
			name:     "negative numbers",
			slice:    []int{-3, -2, -1, 0, 1},
			value:    -2,
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Contains(tt.slice, tt.value)
			if result != tt.expected {
				t.Errorf("Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestMustGetEnv(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		envValue    string
		fallback    *string
		expected    string
		shouldPanic bool
	}{
		{
			name:        "existing environment variable",
			key:         "TEST_EXISTING",
			envValue:    "test_value",
			fallback:    nil,
			expected:    "test_value",
			shouldPanic: false,
		},
		{
			name:        "missing env with fallback",
			key:         "TEST_MISSING_WITH_FALLBACK",
			envValue:    "",
			fallback:    Pointer("fallback_value"),
			expected:    "fallback_value",
			shouldPanic: false,
		},
		{
			name:        "missing env without fallback should panic",
			key:         "TEST_MISSING_NO_FALLBACK",
			envValue:    "",
			fallback:    nil,
			expected:    "",
			shouldPanic: true,
		},
		{
			name:        "empty env with fallback",
			key:         "TEST_EMPTY_WITH_FALLBACK",
			envValue:    "",
			fallback:    Pointer("default_value"),
			expected:    "default_value",
			shouldPanic: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up environment
			if tt.envValue != "" {
				os.Setenv(tt.key, tt.envValue)
				defer os.Unsetenv(tt.key)
			} else {
				os.Unsetenv(tt.key)
			}

			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("MustGetEnv() should have panicked")
					} else {
						// Verify the panic message contains the expected key
						panicMsg := fmt.Sprintf("%v", r)
						if !strings.Contains(panicMsg, tt.key) {
							t.Errorf("Panic message should contain key %q, got: %v", tt.key, panicMsg)
						}
					}
				}()
				MustGetEnv(tt.key, tt.fallback)
			} else {
				result := MustGetEnv(tt.key, tt.fallback)
				if result != tt.expected {
					t.Errorf("MustGetEnv() = %v, want %v", result, tt.expected)
				}
			}
		})
	}
}

func TestCheck(t *testing.T) {
	tests := []struct {
		name        string
		err         error
		msg         string
		shouldPanic bool
	}{
		{
			name:        "no error",
			err:         nil,
			msg:         "test message",
			shouldPanic: false,
		},
		{
			name:        "with error should panic",
			err:         errors.New("test error"),
			msg:         "test message",
			shouldPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.shouldPanic {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("Check() should have panicked")
					} else {
						// Verify the panic message contains the expected message and error
						panicMsg := fmt.Sprintf("%v", r)
						if !strings.Contains(panicMsg, tt.msg) {
							t.Errorf("Panic message should contain %q, got: %v", tt.msg, panicMsg)
						}
						if !strings.Contains(panicMsg, tt.err.Error()) {
							t.Errorf("Panic message should contain error %q, got: %v", tt.err.Error(), panicMsg)
						}
					}
				}()
			}
			Check(tt.err, tt.msg)
		})
	}
}
