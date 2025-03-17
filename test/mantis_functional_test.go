package main_test

import (
	"bytes"
	"os/exec"
	"regexp"
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

func TestVersion(t *testing.T) {

	cmd := exec.Command("./mantis.exe", "-v")

	//hold output in buffer
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	if err != nil {
		t.Errorf("error executing binary: %v", err)
	}

	pattern := `^(?m)mantis v[0-9]+\.[0-9]+\.[0-9]+$`
	re := regexp.MustCompile(pattern)

	if !re.MatchString(buf.String()) {
		t.Errorf("Doesn't match pattern %q; Got %q", pattern, buf.String())
	}

}

func TestHelpArgs(t *testing.T) {

	cmd := exec.Command("./mantis.exe", "-h")

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

func Test_NoLC_NoMod_SF(t *testing.T) {
	// test for a simple go file that prints a statement, assuming local config json is not present, and without modification
	// create temporary directory with file
	// cleanup after test
}
