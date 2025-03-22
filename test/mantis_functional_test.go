//go:build functional

package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

var FILEPATH string = "./testdata/sample.go"
var MODIFY_FILEPATH string

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

func modifyTestFile(t *testing.T, file string, content string) {
	t.Helper()
	// newcomment := "//comment to simulate modification"
	f, err := os.OpenFile(file, os.O_APPEND, 0777)
	if err != nil {
		t.Fatalf("unable to open file for modification: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("unable to simulate modification of file %v", err)
	}
}

func resetModification(t *testing.T, file string) {
	t.Helper()
	data, err := os.ReadFile(file)
	if err != nil {
		t.Fatalf("error reading file %v", err)
	}

	var nl string
	if bytes.Contains(data, []byte("\r\n")) {
		nl = "\r\n"
	} else {
		nl = "\n"
	}

	lines := strings.Split(strings.TrimRight(string(data), "\r\n"), nl)
	if len(lines) > 0 {
		lines = lines[0 : len(lines)-1]
	}
	err = os.WriteFile(file, []byte(strings.Join(lines, nl)+nl), 0644)
	if err != nil {
		t.Fatalf("unable to write back %v", err)
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

func runInteractiveCommand(t *testing.T, sec int, sendStop bool, restart bool, modify bool, args ...string) (string, error) {
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
	if modify {
		modifyTestFile(t, MODIFY_FILEPATH, "//comment to simulate modification")
		time.Sleep(time.Duration(sec) * time.Second)

	}
	if restart {
		stdin.Write([]byte("r\n")) //Send restart
		time.Sleep(time.Duration(sec) * time.Second)
	}
	if sendStop {
		stdin.Write([]byte("q\n")) // Send stop command
		err = cmd.Wait()
		if err != nil {
			t.Fatalf("command stop failed: %v", err)
		}
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

func Test_NC_Quit(t *testing.T) {
	output, err := runInteractiveCommand(t, 1, true, false, false, "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	pattern := `(?m) *.* process has terminated`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(output) {
		t.Errorf("failed to quit, expected pattern in output: %v", pattern)
	}
}

func Test_NC_Restart(t *testing.T) {
	output, err := runInteractiveCommand(t, 1, true, true, false, "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	pattern := `(?m) *.*Restarting[\s\S]*Restarted`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(output) {
		t.Errorf("expected process to restart; got %v", output)
	}
}

func Test_NC_NoMod_SF(t *testing.T) {

	output, err := runInteractiveCommand(t, 1, true, false, false, "-f", FILEPATH)
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

func Test_NC_NoMod_Dir(t *testing.T) {
	output, err := runInteractiveCommand(t, 1, true, false, false, "-f", filepath.Dir(FILEPATH)+"/")
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
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

}

func Test_NC_NoMod_MF(t *testing.T) {
	output, err := runInteractiveCommand(t, 1, true, false, false, "-f", filepath.Dir(FILEPATH)+"/sample.go", filepath.Dir(FILEPATH)+"/helper.go")
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
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

}

func Test_NC_Delay(t *testing.T) {
	output, err := runInteractiveCommand(t, 2, true, false, false, "-f", FILEPATH, "-d", "1000")
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

	output, err := runInteractiveCommand(t, 2, true, false, false, "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command %v", err)
	}

	expected := "Delaying exec begin by 1000 milliseconds"
	if !strings.Contains(output, expected) {
		t.Fatal("failed; execution not delated as expected")
	}
}

func Test_NC_Args(t *testing.T) {
	output, err := runInteractiveCommand(t, 1, true, false, false, "-f", FILEPATH, "-a", "arg1")
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
	output, err := runInteractiveCommand(t, 1, true, false, false, "-f", FILEPATH, "-e", "key=val")
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

func Test_NC_EnvIncorrect(t *testing.T) {
	output, err := runInteractiveCommand(t, 1, false, false, false, "-f", FILEPATH, "-e", "key")
	if err != nil {
		t.Errorf("error during command execution: %v", err)
	}

	if runtime.GOOS == "windows" {
		pattern := `(?m) *.* The parameter is incorrect`
		re := regexp.MustCompile(pattern)
		if !re.MatchString(output) {
			t.Errorf("[windows]failed; expected output containing pattern %v", pattern)
		}
	} else {
		t.Logf("skipping test check; linux allows empty ENV vars")
	}
}

func Test_LCwA_NoMod_SF(t *testing.T) {
	data := map[string]string{
		"extensions": ".go",
		"ignore":     "",
		"delay":      "",
		"env":        "",
		"args":       "arg1,arg2",
	}
	createLocalConfig(t, data)

	t.Cleanup(func() {
		cleanupLocalConfig(t)
	})

	output, err := runInteractiveCommand(t, 1, true, false, false, "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	pattern := `(?m)Starting execution: *.*testdata/sample.go arg1 arg2`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(output) {
		t.Errorf("expected output to match pattern %v", pattern)
	}
}

func Test_LCwE_NoMod_SF(t *testing.T) {
	data := map[string]string{
		"extensions": ".go",
		"ignore":     "",
		"delay":      "",
		"env":        "key1=val1,key2=val2",
		"args":       "",
	}
	createLocalConfig(t, data)

	t.Cleanup(func() {
		cleanupLocalConfig(t)
	})
	inititalEnvLength := len(os.Environ())
	output, err := runInteractiveCommand(t, 1, true, false, false, "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	pattern := `(?m)[0-9]+> env: *([0-9]+)`

	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(output)
	if len(match) > 1 {
		finalEnvLength, err := strconv.Atoi(match[1])
		if err != nil {
			t.Errorf("failed to convert env length from output")
		}
		if !(finalEnvLength == inititalEnvLength+2) || !re.MatchString(output) {
			t.Errorf("failed verifying env values; expected number %d, got %d", inititalEnvLength+2, finalEnvLength)
		}
	} else {
		t.Errorf("unable to get expected output")
	}
}

func Test_NC_Mod_SF(t *testing.T) {
	MODIFY_FILEPATH = "./testdata/sample.go"
	t.Cleanup(func() { resetModification(t, MODIFY_FILEPATH) })
	output, err := runInteractiveCommand(t, 1, true, false, true, "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	pattern := `(?m) *.*Modified[\s\S]*started new process`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(output) {
		t.Errorf("modification not detected, process not restarted")
	}
}

func Test_LCwExt_Mod_SF(t *testing.T) {
	data := map[string]string{
		"extensions": ".go,.txt",
		"ignore":     "",
		"delay":      "",
		"env":        "",
		"args":       "",
	}
	MODIFY_FILEPATH = "./testdata/sample.txt"
	createLocalConfig(t, data)
	dirpath := filepath.Dir(FILEPATH)
	sampletxt := filepath.Join(dirpath, "sample.txt")
	file, err := os.Create(sampletxt)
	if err != nil {
		t.Fatalf("unable to create sample file")
	}
	defer file.Close()

	t.Cleanup(func() {
		cleanupLocalConfig(t)
		os.Remove(sampletxt)
	})

	output, err := runInteractiveCommand(t, 1, true, false, true, "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	pattern := `(?m) *.*Modified[\s\S]*started new process`
	re := regexp.MustCompile(pattern)
	if !re.MatchString(output) {
		t.Errorf("modification not detected, process not restarted")
	}
}

func Test_LCwIgnore_Mod_SF(t *testing.T) {
	data := map[string]string{
		"extensions": ".go",
		"ignore":     "sample.txt",
		"delay":      "",
		"env":        "",
		"args":       "",
	}
	MODIFY_FILEPATH = "./testdata/sample.txt"
	createLocalConfig(t, data)
	dirpath := filepath.Dir(FILEPATH)
	sampletxt := filepath.Join(dirpath, "sample.txt")
	file, err := os.Create(sampletxt)
	if err != nil {
		t.Fatalf("unable to create sample file")
	}
	defer file.Close()

	t.Cleanup(func() {
		cleanupLocalConfig(t)
		os.Remove(sampletxt)
	})

	output, err := runInteractiveCommand(t, 1, true, false, true, "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	pattern := `(?m) *.*Modified[\s\S]*started new process`
	re := regexp.MustCompile(pattern)
	if re.MatchString(output) {
		t.Errorf("modification detected, expected modification to be ignored")
	}
}

func Test_LCwIgnoreDir_Mod_SF(t *testing.T) {
	data := map[string]string{
		"extensions": ".go",
		"ignore":     "sampledir/",
		"delay":      "",
		"env":        "",
		"args":       "",
	}
	MODIFY_FILEPATH = "./testdata/sampledir/sample.txt"
	createLocalConfig(t, data)
	dirpath := filepath.Dir(MODIFY_FILEPATH)
	err := os.Mkdir(dirpath, 0777)
	if err != nil {
		t.Fatalf("unable to create sample dir")
	}
	sampletxt := filepath.Join(dirpath, "sample.txt")
	file, err := os.Create(sampletxt)
	if err != nil {
		t.Fatalf("unable to create sample file")
	}
	defer file.Close()

	t.Cleanup(func() {
		cleanupLocalConfig(t)
		os.RemoveAll(dirpath)
	})

	output, err := runInteractiveCommand(t, 1, true, false, true, "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	pattern := `(?m) *.*Modified[\s\S]*started new process`
	re := regexp.MustCompile(pattern)
	if re.MatchString(output) {
		t.Errorf("modification detected, expected modification to be ignored")
	}
}

func Test_Empty_FArg(t *testing.T) {
	output, err := runCommand(t, "-f")
	if err != nil {
		t.Errorf("error executing command %v", err)
	}
	expected := "parse error"
	if !strings.Contains(output, expected) {
		t.Errorf("expected parse error; got otherwise")
	}
}

func Test_Empty_DArg(t *testing.T) {
	output, err := runInteractiveCommand(t, 2, true, false, false, "-f", FILEPATH, "-d")
	if err != nil {
		t.Errorf("error executing command: %v", err)
	}
	expected := "empty values for -d"
	if !strings.Contains(output, expected) {
		t.Errorf("expected normal execution ignoring -d flag")
	}
}

func Test_Empty_AArg(t *testing.T) {
	output, err := runInteractiveCommand(t, 2, true, false, false, "-f", FILEPATH, "-a")
	if err != nil {
		t.Errorf("error executing command: %v", err)
	}
	expected := "empty values for -a"
	if !strings.Contains(output, expected) {
		t.Errorf("expected normal execution ignoring -a flag")
	}
}

func Test_Empty_EArg(t *testing.T) {
	output, err := runInteractiveCommand(t, 2, true, false, false, "-f", FILEPATH, "-e")
	if err != nil {
		t.Errorf("error executing command: %v", err)
	}
	expected := "empty values for -e"
	if !strings.Contains(output, expected) {
		t.Errorf("expected normal execution ignoring -e flag")
	}
}

func Test_Reordered_Flags(t *testing.T) {
	output, err := runInteractiveCommand(t, 2, true, false, false, "-a", "args", "-d", "1000", "-f", FILEPATH)
	if err != nil {
		t.Errorf("error executing command: %v", err)
	}
	expected := "started new process"
	if !strings.Contains(output, expected) {
		t.Errorf("processes failed to start; expected normal execution")
	}
}
