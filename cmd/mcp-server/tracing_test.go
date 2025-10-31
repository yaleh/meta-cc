package main

import (
	"testing"
)

// TestInitTracing tests the InitTracing function
func TestInitTracing(t *testing.T) {
	cleanup, err := InitTracing()
	if err != nil {
		t.Fatalf("InitTracing failed: %v", err)
	}
	defer cleanup()
}
