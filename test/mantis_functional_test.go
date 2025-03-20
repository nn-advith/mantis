//go:build functional

package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"testing"
	"time"
)

func createLocalConfig(t *testing.T, data map[string]string) {
	t.Helper()
	jsondata, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		t.Fatalf("failed to marshal data %v", err)
	}
	err = os.WriteFile("./testdata/mantis.json", jsondata, 0644)
	if err != nil {
		t.Fatalf("failed to write config file %v", err)
	}

}

func cleanupLocalConfig(t *testing.T) {
	t.Helper()
	file := "./testdata/mantis.json"
	err := os.Remove(file)
	if err != nil {
		t.Fatalf("failed to remove temp config file %v", err)
	}
}

func runCommand(t *testing.T, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("./mantis", args...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	err := cmd.Run()
	return buf.String(), err
}

func runInteractiveCommand(t *testing.T, sec int, args ...string) (string, error) {
	t.Helper()
	cmd := exec.Command("./mantis", args...)

	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("failed to get stdin pipe: %v", err)
	}

	if err := cmd.Start(); err != nil {
		t.Fatalf("failed to start command: %v", err)
	}

	time.Sleep(time.Duration(sec) * time.Second)
	stdin.Write([]byte("q\n")) // Send stop command
	err = cmd.Wait()
	if err != nil {
		t.Fatalf("command stop failed: %v", err)
	}
	return buf.String(), err
}

func TestNoArgs(t *testing.T) {
	output, err := runCommand(t)
	if err != nil {
		t.Errorf("error executing binary: %v", err)
	}
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Usage not shown; got %v; failing test", output)
	}
}

func TestVersion(t *testing.T) {

	output, err := runCommand(t, "-v")
	if err != nil {
		t.Errorf("error executing binary: %v", err)
	}

	pattern := `^(?m)mantis v[0-9]+\.[0-9]+\.[0-9]+$`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(output) {
		t.Errorf("Doesn't match pattern %q; Got %q", pattern, output)
	}

}

func TestHelpArgs(t *testing.T) {

	output, err := runCommand(t, "-h")
	if err != nil {
		t.Errorf("error executing binary: %v", err)
	}
	if !strings.Contains(output, "Usage:") {
		t.Errorf("Usage not shown; got %v; failing test", output)
	}
}

func Test_NoLC_NoMod_SF(t *testing.T) {

	output, err := runInteractiveCommand(t, 1, "-f", "./testdata/sample.go")
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	expected1 := "using global mantis config"
	expected2 := `run ./testdata/sample.go`
	pattern := `(?m) started new process ([0-9]+)`

	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(output)
	if len(match) > 1 {
		pattern = fmt.Sprintf("(?m)%s> TEMP", match[1])
		re = regexp.MustCompile(pattern)
		if !re.MatchString(output) {
			t.Errorf("Wrong output from file; expected \"%s> TEMP\" in output", match[1])
		}
	} else {
		t.Errorf("Expected process start in format %v", pattern)
	}

	if !strings.Contains(output, expected1) {
		t.Errorf("not using global config; expected in output: %v", expected1)
	}
	if !strings.Contains(output, expected2) {
		t.Errorf("not running the correct file")
	}

}

func Test_NC_Delay(t *testing.T) {
	output, err := runInteractiveCommand(t, 2, "-f", "./testdata/sample.go", "-d", "1000")
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	expected := "Delaying exec begin by 1000 milliseconds"
	if !strings.Contains(output, expected) {
		t.Fatal("failed; execution not delated as expected")
	}

}

func Test_LCwD_NoMod_SF(t *testing.T) {
	data := map[string]string{
		"extensions": ".go",
		"ignore":     "",
		"delay":      "1000",
		"env":        "",
		"args":       "",
	}
	createLocalConfig(t, data)

	t.Cleanup(func() {
		cleanupLocalConfig(t)
	})

	output, err := runInteractiveCommand(t, 2, "-f", "./testdata/sample.go")
	if err != nil {
		t.Errorf("error executing command %v", err)
	}

	expected := "Delaying exec begin by 1000 milliseconds"
	if !strings.Contains(output, expected) {
		t.Fatal("failed; execution not delated as expected")
	}
}

func Test_NC_Args(t *testing.T) {
	output, err := runInteractiveCommand(t, 1, "-f", "./testdata/sample.go", "-a", "arg1")
	if err != nil {
		t.Errorf("error during command execution: %v", err)
	}
	pattern := `(?m)Starting execution: *.*testdata/sample.go ([0-9a-zA-Z]+)`

	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(output)
	if len(match) > 1 {
		if !(match[1] == "arg1") || !re.MatchString(output) {
			t.Errorf("output doesnt match expected pattern : %v", pattern)
		}
	} else {
		t.Errorf("unable to find arg passed")
	}
}

func Test_NC_Env(t *testing.T) {

	inititalEnvLength := len(os.Environ())
	output, err := runInteractiveCommand(t, 1, "-f", "./testdata/sample.go", "-e", "key=val")
	if err != nil {
		t.Errorf("error during command execution: %v", err)
	}
	pattern := `(?m)[0-9]+> env: *([0-9]+)`

	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(output)
	if len(match) > 1 {
		finalEnvLength, err := strconv.Atoi(match[1])
		if err != nil {
			t.Errorf("failed to convert env length from output")
		}
		if !(finalEnvLength == inititalEnvLength+1) || !re.MatchString(output) {
			t.Errorf("failed verifying env values; expected number %d, got %d", inititalEnvLength+1, finalEnvLength)
		}
	} else {
		t.Errorf("unable to get expected output")
	}
}

// func Test_NC_EnvIncorrect() {

// }

// tests to be added; check if logic can be reused
//
// no config simple file with args
// no config simple file with env
// local config simple file with delay, no mod  -- done
// local config simple file with args, no mod
// local config simple file with env, no mod
// local confif simple file, modify
// local config simple file, ignore dir, modify
// local config simple file, extensions, modify

//negative scenarios
// no config, simple file -f error cases
// see what else
// restart command test
// quit test
