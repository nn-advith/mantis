//go:build windows

// Tests for mantis.

package main

import (
	"path/filepath"
	"testing"
)

func TestGetGlobalConfigPath(t *testing.T) {
	t.Setenv("APPDATA", "C:\\Users\\Test\\AppData")

	expected := filepath.Join("C:\\Users\\Test\\AppData", "mantis", "mantis.json")
	result := getGlobalConfigPath()

	if result != expected {
		t.Errorf("Expected %s; Got %s", expected, result)
	}
}
