//go:build linux

package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

func getGlobalConfigPath() string {
	var globalConfigPath string
	if home, err := os.UserHomeDir(); err != nil {
		fmt.Println("unable to get user home dir", err)
	} else {
		globalConfigPath = filepath.Join(home, ".config", "mantis")
	}
	return filepath.Join(globalConfigPath, "mantis.json")
}

func killProcess() error {

	if cprocess != nil {

		pgid, err := syscall.Getpgid(cprocess.Pid)
		if err != nil {
			cprocess = nil
			return fmt.Errorf("unable to find pgid: %v", err)

		}
		// killing pgid

		// kchan := make(chan struct{}) // channel for waiting

		err = syscall.Kill(-pgid, syscall.SIGKILL)
		if err != nil {
			return fmt.Errorf("error killing process group: %v", err)

		}

		err = cprocess.Kill() //fallthorugh kill
		if err != nil {
			return fmt.Errorf("error killing process group: %v", err)
		}
		_, err = cprocess.Wait()
		if err != nil {
			return fmt.Errorf("error waiting for process: %v", err)

		} else {
			log.Printf("process has terminated")
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
	executor.SysProcAttr = &syscall.SysProcAttr{Setpgid: true} // flag to run process and children as a group; avoid orphans on sigkill

	// moving this to executionDriver for better(informative maybe) logging

	// executor.Stdout = os.Stdout
	// executor.Stderr = os.Stderr

	return executor, nil
}
