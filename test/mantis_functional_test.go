//go:build functional

package main_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"
)

func createLocalConfig(data map[string]string) error {
	// data := map[string]string{
	// 	"extensions": ".go",
	// 	"ignore":     "",
	// 	"delay":      "0",
	// 	"env":        "",
	// 	"args":       "",
	// }

	jsondata, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		return err
	}
	err = os.WriteFile("./testdata/mantis.json", jsondata, 0644)
	if err != nil {
		return err
	}
	return nil
}

func cleanupLocalConfig() error {
	file := "./testdata/mantis.json"
	err := os.Remove(file)
	if err != nil {
		return err
	}
	return nil
}

func TestNoArgs(t *testing.T) {
	cmd := exec.Command("./mantis")

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

	cmd := exec.Command("./mantis", "-v")

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

	cmd := exec.Command("./mantis", "-h")

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
	cmd := exec.Command("./mantis", "-f", "./testdata/sample.go")
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf

	stdin, err := cmd.StdinPipe() //input pipe to send quit command
	if err != nil {
		t.Fatal("unable to configure command")
	}

	// err = createLocalConfig()
	// if err != nil {
	// 	t.Errorf("some error")
	// }
	// t.Cleanup(func() {
	// 	err = cleanupLocalConfig()
	// 	if err != nil {
	// 		t.Fatal("error cleaning up dummy config json; please remove manually")
	// 	}
	// })

	expected1 := "using global mantis config"
	expected2 := `run ./testdata/sample.go`
	pattern := `(?m) started new process ([0-9]+)`

	if err := cmd.Start(); err != nil { //Start over Run to wait
		t.Fatalf("error executing command %v", err)
	}

	time.Sleep(1 * time.Second)

	stdin.Write([]byte("q\n")) //send stop
	err = cmd.Wait()
	if err != nil {
		t.Fatalf("error waiting for stop: %v", err)
	}
	output := buf.String()

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

func Test_LCwD_NoMod_SF(t *testing.T) {

}

// tests to be added; check if logic can be reused
//
// no config simple file with args
// no config simple file with env
// local config simple file with delay, no mod
// local config simple file with args, no mod
// local config simple file with env, no mod
// local confif simple file, modify
// local config simple file, ignore dir, modify
// local config simple file, extensions, modify

//negative scenarios
// no config, simple file -f error cases
// see what else
