//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
)

func killProcess() {
	if cprocess != nil {

		_, err := os.FindProcess(cprocess.Pid)
		if err != nil {
			fmt.Println("error finding process; likely terminated: ", err)
			cprocess = nil
			return
		}

		killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cprocess.Pid))
		killCmd.Stdout = os.Stdout
		killCmd.Stderr = os.Stderr

		if err := killCmd.Run(); err != nil { //Run() is blocking so will wait for killcmd to complete
			fmt.Println("taskkill err:", err)
			return
		}
		cprocess = nil

	}

}

func commandConstruct(filepath string) *exec.Cmd {
	executor := exec.Command("go", "run", filepath)
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	return executor
}
