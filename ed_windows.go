//go:build windows

package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

func getGlobalConfigPath() string {

	globalConfigPath := filepath.Join(os.Getenv("APPDATA"), "mantis")
	return filepath.Join(globalConfigPath, "mantis.json")
}

func killProcess() {
	if cprocess != nil {

		// os.FindProcess is stupid. it just finds the process but not the running status.
		// so emulating ps -aef | grep -i pid ; but windows version
		chckCmd := exec.Command("tasklist", "/FI", "PID eq "+strconv.Itoa(cprocess.Pid))
		chckCmd.Stderr = os.Stderr
		var out bytes.Buffer
		chckCmd.Stdout = &out

		if err := chckCmd.Run(); err != nil {
			fmt.Println("error running check")
			return
		}

		output := out.String()
		if output == "" || bytes.Contains(out.Bytes(), []byte("No tasks are running")) {
			fmt.Println("process has exited already")
			return
		} else {
			killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cprocess.Pid))
			killCmd.Stdout = os.Stdout
			killCmd.Stderr = os.Stderr

			if err := killCmd.Run(); err != nil { //Run() is blocking so will wait for killcmd to complete
				fmt.Println("taskkill err:", err)
				return
			}
		}
		cprocess = nil

	}

}

func commandConstruct(args map[string][]string) (*exec.Cmd, error) {

	var executor *exec.Cmd

	if len(args["-a"]) != 0 {
		executor = exec.Command("go", append(append([]string{"run"}, args["-f"]...), args["-a"]...)...)
	} else {
		executor = exec.Command("go", append([]string{"run"}, args["-f"]...)...)
	}
	if len(args["-e"]) > 0 {
		executor.Env = append(os.Environ(), args["-e"]...)
	}

	// executor.Dir = WDIR
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	fmt.Println(executor)

	return executor, nil
}
