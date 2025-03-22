package testutils

import (
	"bytes"
	"encoding/json"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"
)

func CreateLocalConfig(t *testing.T, data map[string]string) {
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

func CleanupLocalConfig(t *testing.T) {
	t.Helper()
	file := "./testdata/mantis.json"
	err := os.Remove(file)
	if err != nil {
		t.Fatalf("failed to remove temp config file %v", err)
	}
}

func ModifyTestFile(t *testing.T, file string, content string) {
	t.Helper()
	// newcomment := "//comment to simulate modification"
	f, err := os.OpenFile(file, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		t.Fatalf("unable to open file for modification: %v", err)
	}
	defer f.Close()

	if _, err := f.WriteString(content); err != nil {
		t.Fatalf("unable to simulate modification of file %v", err)
	}
}

func ResetModification(t *testing.T, file string) {
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

func CreateDummyProcess(t *testing.T) *exec.Cmd {
	t.Helper()
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("timeout", "10")
	} else {
		cmd = exec.Command("sleep", "10")
	}
	return cmd
	// if err := cmd.Start(); err != nil {
	// 	t.Errorf("error starting dummy process")
	// }
	// *cprocess = cmd.Process
}
