package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"
)

//mantis -- filename.go

// go install github.com/path@latest -> this is how it will work with installing; assuming bin is added to path

var INIT_FILE_SIZE int
var INIT_MOD_TIME int

var mlock sync.Mutex
var cprocess *os.Process

type Event struct {
	eventcode int
	eventname string
}

func usage() {
	fmt.Printf("\nUsage:\n\nmantis -- <filename or filepath>\n")
	os.Exit(1)
}

func checkFileExists(filepath string) error {
	fmt.Println(filepath)
	_, err := os.Stat(filepath)
	if os.IsNotExist(err) {
		return fmt.Errorf("file doesn't exist: %s", err)
	}
	return nil
}

func getModTime(filepath string) (int, error) {
	fileInfo, err := os.Stat(filepath)
	if err != nil {
		return 0, fmt.Errorf("error stat %v", err)
	}
	// fmt.Printf("%+v", fileInfo)
	return int(fileInfo.ModTime().Unix()), nil

}

func getFileSize(filepath string) (int, error) {
	fileinfo, err := os.Stat(filepath)
	if err != nil {

		return 0, fmt.Errorf("error stat %v", err)

	}
	return int(fileinfo.Size()), nil
}

func killProcess() {
	if cprocess != nil {

		//check if process is running

		// in case of error in filepath this will fail since it will search for pid even when pid has terminated
		// add some kind of checks here to avoid this
		if runtime.GOOS == "windows" {

			p, err := os.FindProcess(cprocess.Pid)
			if err != nil {
				fmt.Println("error finding porocess", err)
				cprocess = nil
			}
			if p.Pid != cprocess.Pid {
				fmt.Println("killed already", p)
				cprocess = nil
				return
			} else {

				killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cprocess.Pid))
				killCmd.Stdout = os.Stdout
				killCmd.Stderr = os.Stderr

				if err := killCmd.Run(); err != nil {
					fmt.Println("taskkill err:", err)
					return
				}
			}
		} else {
			err := cprocess.Signal(syscall.Signal(0))
			if err != nil {
				fmt.Println("Killed already.", err)
				cprocess = nil
				return
			} else {

				cprocess.Release()
				cprocess.Kill()
				cprocess = nil
			}
		}
	}
}

func executionDriver(filepath string) {

	mlock.Lock()

	killProcess()

	executor := exec.Command("go", "run", filepath)
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	err := executor.Start()
	if err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		return
	}
	cprocess = executor.Process
	fmt.Printf("started new process %v", cprocess.Pid)

	mlock.Unlock()

}

func checkIfModified(filepath string, channel chan Event) {
	for {

		if fsize, _ := getFileSize(filepath); fsize != INIT_FILE_SIZE {
			fmt.Println("File has been modified")
			INIT_FILE_SIZE = fsize
			fmodt, _ := getModTime(filepath)
			INIT_MOD_TIME = fmodt
			channel <- Event{eventcode: 101, eventname: "modified"}

		} else {
			fmodt, _ := getModTime(filepath)
			if fmodt != INIT_MOD_TIME {
				channel <- Event{eventcode: 101, eventname: "modified"}
				INIT_MOD_TIME = fmodt

			} else {
				channel <- Event{eventcode: 100, eventname: "crickets"}
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func listenForInput(inputChannel chan int) {
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _ := reader.ReadString('\n')
		text = strings.TrimSpace(text)
		if text == "c" {
			fmt.Println("here")
			inputChannel <- 0
		}
		inputChannel <- 1 //add other cases here
	}
}

func main() {
	fmt.Println("Starting mantis:")

	if len(os.Args) <= 2 {
		usage()
	}

	filepath := os.Args[2]
	//check if file exists; either here or in filwatcher
	err := checkFileExists(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}

	INIT_FILE_SIZE, err = getFileSize(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	INIT_MOD_TIME, err = getModTime(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}

	//create an exec command for executing the file

	//channel creation
	filechannel := make(chan Event)
	inputchannel := make(chan int)
	// execChannel := make(chan int)

	// go changesListener(filechannel)
	go listenForInput(inputchannel)
	go checkIfModified(filepath, filechannel)
	// go executionDriver(filepath, execChannel)

	executionDriver(filepath)
	go func() {
		for {
			select {
			case data, ok := <-inputchannel:
				// fmt.Println(data)
				if !ok {
					fmt.Println("Input channel closed")
					os.Exit(1)
				}
				if data == 0 {
					//terminate the execution of go code as well
					fmt.Println("c called", strconv.Itoa(cprocess.Pid))
					killProcess()
					close(inputchannel)
					close(filechannel)
					// terminate all running processes ----- PENDING
					os.Exit(0)
				} else if data == 1 {
					fmt.Println("Some other operation")
				}

			case data, ok := <-filechannel:
				// fmt.Println(data)
				if !ok {
					fmt.Println("File channel is closed; exiting")
					os.Exit(1)
				}
				if data.eventcode == 101 {
					//modified, restart
					fmt.Println("PID", cprocess.Pid)

					executionDriver(filepath)
				}

			}
		}
	}()

	select {}
}
