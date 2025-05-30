//go:build windows

package mantis

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

func GetGlobalConfigPath() string {
	globalConfigPath := filepath.Join(os.Getenv("APPDATA"), "mantis")
	return filepath.Join(globalConfigPath, "mantis.json")
}

func KillProcess() error {
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

func CommandConstruct() (*exec.Cmd, error) {

	var executor *exec.Cmd

	if len(globalargs["-a"]) != 0 {
		executor = exec.Command("go", append(append([]string{"run"}, globalargs["-f"]...), globalargs["-a"]...)...)
	} else {
		argsEnv := strings.Split(mantis_config.Args, ",")
		if len(argsEnv) > 0 {
			executor = exec.Command("go", append(append([]string{"run"}, globalargs["-f"]...), argsEnv...)...)
		} else {
			executor = exec.Command("go", append([]string{"run"}, globalargs["-f"]...)...)
		}
	}
	if len(globalargs["-e"]) > 0 {
		executor.Env = append(os.Environ(), globalargs["-e"]...)
	} else {
		configEnv := strings.Split(mantis_config.Env, ",")
		if len(configEnv) > 0 {
			executor.Env = append(os.Environ(), configEnv...)
		}
	}

	// moving this to ExecutionDriver for better(informative maybe) logging

	// executor.Stdout = os.Stdout
	// executor.Stderr = os.Stderr

	return executor, nil
}
