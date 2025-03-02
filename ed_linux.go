//go:build linux

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func killProcess() {
	fmt.Println("Linux kill")
	if cprocess != nil {

		pgid, err := syscall.Getpgid(cprocess.Pid)
		if err != nil {
			fmt.Println("unable to find pgid: ", err)
			cprocess = nil
			return
		}
		// killing pgid

		// kchan := make(chan struct{}) // channel for waiting

		err = syscall.Kill(-pgid, syscall.SIGKILL)
		if err != nil {
			fmt.Println("error killing process group:", err)
			return
		}

		err = cprocess.Kill() //fallthorugh kill
		if err != nil {
			fmt.Println("error killing process:", err)
			return
		}
		_, err = cprocess.Wait()
		if err != nil {
			fmt.Println("error waiting for process:", err)
			return
		} else {
			fmt.Println("process has terminated")
		}

		cprocess = nil
	}
	return
}

func commandConstruct(filepath string) *exec.Cmd {
	executor := exec.Command("go", "run", filepath)
	executor.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // flag to run process and children as a group; avoid orphans on sigkill
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	return executor
}
