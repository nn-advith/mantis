package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

//mantis -- filename.go

// go install github.com/path@latest -> this is how it will work with installing; assuming bin is added to path

type Event struct {
	eventcode int
	eventname string
}

type MantisConfig struct {
	Monitor    string `json:"monitor"`
	Extensions string `json:"extensions"`
	Main       string `json:"main"`
	Ignore     string `json:"ignore"`
	Delay      string `json:"delay"`
	Env        string `json:"env"`
	Flags      string `json:"flags"`
}

var MANTIS_CONFIG MantisConfig

var globalargs map[string][]string
var CONFIG_FILE string
var WDIR string

var MONITOR_LIST map[string][]int = map[string][]int{}

var mlock sync.Mutex
var cprocess *os.Process

func usage() {
	fmt.Printf("\nUsage:\n\nmantis -f <files> -a <args> -e <key=value>\n")
}

func executionDriver(args map[string][]string) {

	mlock.Lock()

	killProcess()

	executor, err := commandConstruct(args)
	if err != nil {
		fmt.Println("error constructing command", err)
	}

	err = executor.Start()
	if err != nil {
		fmt.Printf("Error starting command: %v\n", err)
		return
	}
	cprocess = executor.Process
	fmt.Printf("started new process %v", cprocess.Pid)

	mlock.Unlock()

}

// fileMap := map[string]FileMeta{
// 	"path/to/file1": {size, modtime},
// 	"path/to/file2": {size, modtime},
// }

//check if dir has modified,

// build a list of files and directories relative to wdir
// check if the directory / files in wdir have modified
// if directory has modified, then move in and check if modified file is ignores/matches extension

// this should check for all files, not just one
func checkIfModified(channel chan Event) {

	for {

		modified := false

		var wg sync.WaitGroup

		for k, v := range MONITOR_LIST {
			wg.Add(1)

			go func(k string, v []int) {
				defer wg.Done()

				fileinfo, err := os.Stat(k)
				if err != nil {
					fmt.Println("error stat")
				}
				newsize := int(fileinfo.Size())
				newmodtime := int(fileinfo.ModTime().Unix())

				if newsize != v[0] {
					modified = true
					MONITOR_LIST[k] = []int{newsize, newmodtime}
				} else {
					if newmodtime > v[1] {
						modified = true
						MONITOR_LIST[k] = []int{newsize, newmodtime}

					}
				}

			}(k, v)
		}

		wg.Wait()
		if modified {
			channel <- Event{eventcode: 101, eventname: "modified"}
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
		// inputChannel <- 1 //add other cases here
	}
}

func main() {

	fmt.Println("Starting mantis:")

	err := preExec()
	if err != nil {
		fmt.Println("error during preexec ", err)
		return
	}

	filechannel := make(chan Event)
	inputchannel := make(chan int)

	go listenForInput(inputchannel)
	go checkIfModified(filechannel)

	executionDriver(globalargs)
	fmt.Println(MONITOR_LIST)
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

					executionDriver(globalargs)
				}

			}
		}
	}()

	select {}
}
