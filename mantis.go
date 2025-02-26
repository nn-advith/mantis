package main

import (
	"bufio"
	"fmt"
	"mantis/filewatcher"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

//mantis filename.go
//
//on startup: compute file hash and start program
//on file save: recompute hash, if hash different restart file
//on error: redirect error
//command line utils to restart, terminate etc

//loop
//file system watcher - fsnotify..build custom if possible - run this as a go routine
//debounce delay
//hash compute and compare
//restart if changes

// func checkForFileChanges(filepath string) bool {

// }

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

func executionDriver(filepath string) {

	mlock.Lock()
	defer mlock.Unlock()

	if cprocess != nil {

		if runtime.GOOS == "windows" {
			killCmd := exec.Command("taskkill", "/F", "/T", "/PID", strconv.Itoa(cprocess.Pid))
			killCmd.Stdout = os.Stdout
			killCmd.Stderr = os.Stderr

			if err := killCmd.Run(); err != nil {
				fmt.Println("taskkill err:", err)
				return
			}
		} else {

			cprocess.Release()
			cprocess.Kill()
			cprocess = nil
		}
	}

	// ctx, cancel := context.WithCancel(context.Background())
	// defer cancel()

	executor := exec.Command("go", "run", filepath)
	executor.Stdout = os.Stdout
	executor.Stderr = os.Stderr

	err := executor.Start()
	if err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		return
	}
	cprocess = executor.Process

	// err = executor.Wait()
	// if err != nil {
	// 	fmt.Println("asdasdasd")
	// }

	fmt.Printf("started new process %v", cprocess.Pid)

}

func checkIfModified(filepath string, channel chan Event) {
	for {

		if fsize, _ := filewatcher.GetFileSize(filepath); fsize != INIT_FILE_SIZE {
			fmt.Println("File has been modified")
			INIT_FILE_SIZE = fsize
			fmodt, _ := filewatcher.GetModTime(filepath)
			INIT_MOD_TIME = fmodt
			channel <- Event{eventcode: 101, eventname: "modified"}

		} else {
			fmodt, _ := filewatcher.GetModTime(filepath)
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
	err := filewatcher.CheckFileExists(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s", err)
		os.Exit(1)
	}

	INIT_FILE_SIZE, err = filewatcher.GetFileSize(filepath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
	}
	INIT_MOD_TIME, err = filewatcher.GetModTime(filepath)
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
					close(inputchannel)
					close(filechannel)
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

	// for {
	// 	err := executor.Run()
	// 	if err!=nil {
	// 		fmt.Fprintf(os.Stderr, "error executing command: %v\n", err)
	// 		os.Exit(1)
	// 	}
	// }

	// modtime := filewatcher.GetModTime("./filewatcher/sample.txt")
	// if modtime == -1 {
	// 	fmt.Printf("Error in modtime")
	// 	return
	// }

	// fmt.Println(modtime)
}
