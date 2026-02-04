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
