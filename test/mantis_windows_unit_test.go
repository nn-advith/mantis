//go:build unit && windows

// Tests for mantis.

package main_test

import (
	"path/filepath"
	"testing"

	"github.com/nn-advith/mantis/mantis"
)

func TestGetGlobalConfigPath(t *testing.T) {
	t.Setenv("APPDATA", "C:\\Users\\Test\\AppData")

	expected := filepath.Join("C:\\Users\\Test\\AppData", "mantis", "mantis.json")
	result := mantis.GetGlobalConfigPath()

	if result != expected {
		t.Errorf("Expected %s; Got %s", expected, result)
	}
}
