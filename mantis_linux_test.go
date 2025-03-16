//go:build linux

package main

import (
	"path/filepath"
	"testing"
)

func TestGetGlobalConfigPath(t *testing.T) {
	t.Setenv("HOME", "/home/testusr")
	expected := filepath.Join("/home/testusr", ".config", "mantis", "mantis.json")
	result := getGlobalConfigPath()

	if expected != result {
		t.Errorf("Expected %s; Got %s", expected, result)
	}
}
