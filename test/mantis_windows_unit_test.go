//go:build unit && windows

// Tests for mantis.

package main_test

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/nn-advith/mantis/mantis"
	"github.com/nn-advith/mantis/testutils"
)

func TestGetGlobalConfigPath(t *testing.T) {
	t.Setenv("APPDATA", "C:\\Users\\Test\\AppData")

	expected := filepath.Join("C:\\Users\\Test\\AppData", "mantis", "mantis.json")
	result := mantis.GetGlobalConfigPath()

	if result != expected {
		t.Errorf("Expected %s; Got %s", expected, result)
	}
}

func TestCommandConstruct(t *testing.T) {
	globalargs := map[string][]string{
		"-f": {"testdata/sample.go"},
		"-a": {""},
		"-e": {},
	}
	mantis.SetGlobalArgs(globalargs)

	cmd, err := mantis.CommandConstruct()
	if err != nil {
		t.Fatalf("failed command construct, %v", err)
	}
	expectedPath := `^.*go\.exe$`
	re := regexp.MustCompile(expectedPath)
	expectedArgs := []string{"go", "run", "testdata/sample.go", ""}

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

	if !re.MatchString(cmd.Path) {
		t.Errorf("go binary not present in path")
	}
	if !sliceCompare(expectedArgs, cmd.Args) {
		t.Errorf("expected args for cmd %v; got %v", expectedArgs, cmd.Args)
	}

}

func TestKillProcess(t *testing.T) {

	cmd := testutils.CreateDummyProcess(t)

	if err := cmd.Start(); err != nil {
		t.Errorf("error starting dummy process")
	}

	mantis.SetCProcess(cmd.Process)
	mantis.SuppressLog(1)
	mantis.KillProcess()
	mantis.SuppressLog(0)
	if mantis.GetCProcess() != nil {
		t.Errorf("failed; expected cprocess to be nil")
	}

}
