package main

import (
	"log"
	"os"
)

func Check(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %v", msg, err)
	}
}

func MustGetEnv(key string, fallback *string) string {
	value := os.Getenv(key)
	if value == "" {
		if fallback != nil {
			return *fallback
		}
		log.Fatalf("environment variable %q is required", key)
	}
	return value
}

func Pointer[T any](v T) *T {
	return &v
}

func Value[T any](p *T) T {
	var v T
	if p == nil {
		return v
	}
	return *p
}

func Filter[T any](input []T, fn func(T) bool) []T {
	if len(input) == 0 {
		return []T{}
	}

	result := make([]T, 0)
	for _, value := range input {
		if fn(value) {
			result = append(result, value)
		}
	}
	return result
}

func Map[T any](input []T, fn func(T) T) []T {
	if len(input) == 0 {
		return []T{}
	}

	result := make([]T, len(input))
	for i, value := range input {
		result[i] = fn(value)
	}
	return result
}

func Contains[T comparable](slice []T, value T) bool {
	for _, v := range slice {
		if v == value {
			return true
		}
	}
	return false
}
