//go:build windows

package main

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func getGlobalConfigPath() string {
	globalConfigPath := filepath.Join(os.Getenv("APPDATA"), "mantis")
	return filepath.Join(globalConfigPath, "mantis.json")
}

func killProcess() error {
	if cprocess != nil {

		// os.FindProcess is stupid. it just finds the process but not the running status.
		// so emulating ps -aef | grep -i pid ; but windows version
		chckCmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(cprocess.Pid))
		chckCmd.Stderr = os.Stderr
		var out bytes.Buffer
		chckCmd.Stdout = &out

		if err := chckCmd.Run(); err != nil {
			return fmt.Errorf("error while checking for existing processes")
		}

		output := out.String()
		if output == "" || bytes.Contains(out.Bytes(), []byte("No tasks are running")) {
			log.Printf("process has terminated")

		} else {
			killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cprocess.Pid))
			killCmd.Stdout = os.Stdout
			killCmd.Stderr = os.Stderr

			if err := killCmd.Run(); err != nil { //Run() is blocking so will wait for killcmd to complete
				return fmt.Errorf("error while killing existing process")
			}
		}
		cprocess = nil
	}
	return nil
}

func commandConstruct(args map[string][]string) (*exec.Cmd, error) {

	var executor *exec.Cmd

	if len(args["-a"]) != 0 {
		executor = exec.Command("go", append(append([]string{"run"}, args["-f"]...), args["-a"]...)...)
	} else {
		argsEnv := strings.Split(MANTIS_CONFIG.Args, ",")
		if len(argsEnv) > 0 {
			executor = exec.Command("go", append(append([]string{"run"}, args["-f"]...), argsEnv...)...)
		} else {
			executor = exec.Command("go", append([]string{"run"}, args["-f"]...)...)
		}
	}
	if len(args["-e"]) > 0 {
		executor.Env = append(os.Environ(), args["-e"]...)
	} else {
		configEnv := strings.Split(MANTIS_CONFIG.Env, ",")
		if len(configEnv) > 0 {
			executor.Env = append(os.Environ(), configEnv...)
		}
	}

	// moving this to executionDriver for better(informative maybe) logging

	// executor.Stdout = os.Stdout
	// executor.Stderr = os.Stderr

	return executor, nil
}
