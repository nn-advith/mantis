//go:build linux

package main_test

import (
	"path/filepath"
	"testing"

	"github.com/nn-advith/mantis/mantis"
)

func TestGetGlobalConfigPath(t *testing.T) {
	t.Setenv("HOME", "/home/testusr")
	expected := filepath.Join("/home/testusr", ".config", "mantis", "mantis.json")
	result := mantis.GetGlobalConfigPath()

	if expected != result {
		t.Errorf("Expected %s; Got %s", expected, result)
	}
}
