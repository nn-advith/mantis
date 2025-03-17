package main_test

import (
	"bytes"
	"log"
	"os"
	"testing"

	"github.com/nn-advith/mantis/mantis"
)

var cprocess = &os.Process{Pid: 1234}

func TestLogProcessInfo(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)

	mantis.LogProcessInfo(cprocess, "Sample value")

	log.SetOutput(os.Stdout)
	// cprocess.Pid = originalPid

	expected := "1234> Sample value\n"
	actual := buf.String()
	if expected != actual {
		t.Errorf("Expected %s, Got %s", expected, actual)
	}

}
