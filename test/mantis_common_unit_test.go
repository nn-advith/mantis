//go:build unit

package main_test

import (
	"bytes"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"testing"

	"github.com/nn-advith/mantis/mantis"
	"github.com/nn-advith/mantis/testutils"
)

// simulating global variables
var cprocess *os.Process
var globalargs map[string][]string

func argCompare(t *testing.T, a, b map[string][]string) {
	t.Helper()

	var sliceCompare = func(a, b []string) bool {
		if len(a) != len(b) {
			return false
		}
		for i := range a {
			if a[i] != b[i] {
				return false
			}
		}
		return true
	}

	if len(a) != len(b) {
		t.Errorf("arg compare failed; not equal")
	}
	for k, _ := range a {
		if !sliceCompare(a[k], b[k]) {
			t.Errorf("arg conmpare failed; maps not equal")
		}
	}

}

func TestLogProcessInfo(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	cprocess = &os.Process{Pid: 1234}

	mantis.LogProcessInfo(cprocess, "Sample value")

	log.SetOutput(os.Stdout)
	// cprocess.Pid = originalPid

	expected := "1234> Sample value\n"
	actual := buf.String()
	if expected != actual {
		t.Errorf("Expected %s, Got %s", expected, actual)
	}

}

func TestParseArgs(t *testing.T) {

	globalargs = map[string][]string{
		"-f": make([]string, 0),
		"-a": make([]string, 0),
		"-e": make([]string, 0),
	}
	args := []string{"./mantis", "-f", "sample.go", "-a", "arg1", "-e", "key=val", "-d", "1000"}
	err := mantis.ParseArgs(globalargs, args)
	if err != nil {
		t.Fatalf("error parsing args: %v", err)
	}
	expectedGArgs := map[string][]string{
		"-f": {"sample.go"},
		"-a": {"arg1"},
		"-e": {"key=val"},
	}

	argCompare(t, expectedGArgs, globalargs)
	if mantis.GetGlobalDelay() != 1000 {
		t.Errorf("delay not parsed")
	}
}

func TestCleanFileArgs(t *testing.T) {
	globalargs = map[string][]string{
		"-f": {"testdata/"},
		"-a": {},
		"-e": {},
	}
	mantis.SetGlobalArgs(globalargs)
	err := mantis.CleanFileArgs()
	if err != nil {
		t.Fatalf("failed; error cleaning args %v", err)
	}
	pattern := `^.+[\.]go$`
	re := regexp.MustCompile(pattern)
	fargs := mantis.GetGlobalArgs()["-f"]
	for i := range fargs {
		if !re.MatchString(fargs[i]) {
			t.Errorf("failed; clean file args")
		}
	}

}

func TestCheckForGlobalConfig(t *testing.T) {
	filepath := mantis.GetGlobalConfigPath()
	err := mantis.CheckForGlobalConfig()
	if err != nil {
		t.Fatalf("failed; function call failed %v", err)
	}
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Errorf("expected global config to be created; not created")
	}
}

func TestCheckForLocalConfig(t *testing.T) {

	data := map[string]string{
		"extensions": ".go",
		"ignore":     "",
		"delay":      "",
		"env":        "",
		"args":       "",
	}
	testutils.CreateLocalConfig(t, data)

	t.Cleanup(func() {
		testutils.CleanupLocalConfig(t)
	})

	ftags := []string{"./testdata/sample.go"}
	output, err := mantis.CheckForLocalConfig(ftags)
	if err != nil {
		t.Fatalf("failed; expected%v", err)
	}
	if !output {
		t.Errorf("expected true(config file present); got false(config file not present)")
	}
}

func TestGetFilesToMonitor(t *testing.T) {
	sample_mc := mantis.MantisConfig{
		Extensions: ".go",
		Ignore:     "helper.go",
		Delay:      "",
		Args:       "",
		Env:        "",
	}
	mantis.SetMantisConfig(sample_mc)
	wdtemp, _ := filepath.Abs("./testdata")
	mantis.SetWDirectory(wdtemp)
	mantis.GetFilesToMonitor()

	output := mantis.GetMonitorList()
	pattern := `^.*sample\.go$`
	re := regexp.MustCompile(pattern)

	for k := range output {
		if !re.MatchString(k) {
			t.Errorf("expected to find sample.go in key")
		}
	}
}

func TestDecodeMantisConfig(t *testing.T) {
	data := map[string]string{
		"extensions": ".go",
		"ignore":     "test/",
		"delay":      "10",
		"env":        "k=v",
		"args":       "a1",
	}
	testutils.CreateLocalConfig(t, data)

	t.Cleanup(func() {
		testutils.CleanupLocalConfig(t)
	})

	cftemp, _ := filepath.Abs("./testdata/mantis.json")
	mantis.SetConfigFile(cftemp)

	err := mantis.DecodeMantisConfig()
	if err != nil {
		t.Fatalf("failed during function call %v", err)
	}

	expected := mantis.MantisConfig{Extensions: ".go", Ignore: "test/", Delay: "10", Env: "k=v", Args: "a1"}
	output := mantis.GetMantisConfig()

	if !reflect.DeepEqual(output, expected) {
		t.Errorf("decode test failed; expected %v - got %v", expected, output)
	}

}
