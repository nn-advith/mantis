package main_test

import (
	"bytes"
	"os/exec"
	"strings"
	"testing"
)

func TestNoArgs(t *testing.T) {
	cmd := exec.Command("./mantis.exe")

	//hold output in buffer
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err != nil {
		t.Errorf("error executing binary: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Usage not shown; got %v; failing test", output)
	}
}
