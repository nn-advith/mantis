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

type Event struct {
	eventcode int
	eventname string
}

type MantisConfig struct {
	Extensions string `json:"extensions"`
	Ignore     string `json:"ignore"`
	Delay      string `json:"delay"`
	Env        string `json:"env"`
	Args       string `json:"args"`
}

var MANTIS_CONFIG MantisConfig
var CONFIG_FILE string
var MONITOR_LIST map[string][]int = map[string][]int{}
var WDIR string

var globalargs map[string][]string

var mlock sync.Mutex
var cprocess *os.Process

func executionDriver(args map[string][]string) error {

	mlock.Lock()
	err := killProcess()
	if err != nil {

	}

	executor, err := commandConstruct(args)
	if err != nil {
		return fmt.Errorf("error constructing command: %V", err)
	}
	execDelay, err := strconv.Atoi(MANTIS_CONFIG.Delay)
	if err != nil {
		return fmt.Errorf("error in delay conversion; ensure delay is mentioned in milliseconds: %v", err)
	}
	if execDelay > 0 {
		fmt.Printf("\nDelaying exec begin by %v milliseconds; sleeping now\n", execDelay)
	}
	time.Sleep(time.Duration(execDelay) * time.Millisecond)

	err = executor.Start()
	if err != nil {
		return fmt.Errorf("error starting command: %v", err)

	}
	cprocess = executor.Process
	fmt.Printf("started new process %v\n", cprocess.Pid)

	mlock.Unlock()
	return nil
}

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
					fmt.Println("error while checking for moditifications:", err)
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
		switch text {
		case "q":
			// quit
			inputChannel <- 0
		case "r":
			// restart
			inputChannel <- 1
		default:
			// default
			inputChannel <- -1
		}

	}
}

func main() {

	err := preExec()
	if err != nil {
		fmt.Println("error during preexec ", err)
		return
	}

	filechannel := make(chan Event)
	inputchannel := make(chan int)

	go listenForInput(inputchannel)
	go checkIfModified(filechannel)

	err = executionDriver(globalargs)
	if err != nil {
		fmt.Println(err)
	}

	go func() {
		for {
			select {
			case data, ok := <-inputchannel:
				// fmt.Println(data)
				if !ok {
					fmt.Println("Input channel closed")
					os.Exit(1)
				}

				switch data {
				case 0:
					killProcess()
					close(inputchannel)
					close(filechannel)
					os.Exit(0)
				case 1:
					executionDriver(globalargs)
				case -1:
					fmt.Println("unknown input;")
					runtimeCommandsLegend()
				}

			case data, ok := <-filechannel:
				if !ok {
					fmt.Println("File channel is closed; exiting")
					os.Exit(1)
				}
				switch data.eventcode {
				case 101:
					executionDriver(globalargs)
				}

			}
		}
	}()

	select {}
}
