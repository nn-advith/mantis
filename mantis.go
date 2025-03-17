package main

import (
	"bufio"
	"fmt"
	"log"
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

var mantis_config MantisConfig
var config_file string
var monitor_list map[string][]int = map[string][]int{}
var wdirectory string

var globalargs map[string][]string

var mlock sync.Mutex
var cprocess *os.Process

func ExecutionDriver(args map[string][]string) error {

	mlock.Lock()
	err := KillProcess()
	if err != nil {

	}

	executor, err := CommandConstruct(args)
	if err != nil {
		return fmt.Errorf("error constructing command: %V", err)
	}

	// capture output from pipe and use logProcessInfo within a goroutine to avoid  waiting
	stdout, err := executor.StdoutPipe()
	if err != nil {
		log.Fatal(err)
	}
	stderr, err := executor.StderrPipe()
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			LogProcessInfo(cprocess, scanner.Text())
		}
	}()
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			LogProcessInfo(cprocess, scanner.Text())
		}
	}()

	execDelay, err := strconv.Atoi(mantis_config.Delay)
	if err != nil {
		return fmt.Errorf("error in delay conversion; ensure delay is mentioned in milliseconds: %v", err)
	}
	if execDelay > 0 {
		log.Printf("Delaying exec begin by %v milliseconds; sleeping now\n", execDelay)
	}
	time.Sleep(time.Duration(execDelay) * time.Millisecond)

	log.Printf("Starting execution: %v", executor)
	err = executor.Start()
	if err != nil {
		return fmt.Errorf("error starting command: %v", err)

	}
	cprocess = executor.Process
	log.Printf("%v", fmt.Sprintf("started new process %v\n", cprocess.Pid))

	mlock.Unlock()
	return nil
}

func CheckIfModified(channel chan Event) {

	for {
		modified := false
		var wg sync.WaitGroup
		for k, v := range monitor_list {
			wg.Add(1)
			go func(k string, v []int) {
				defer wg.Done()

				fileinfo, err := os.Stat(k)
				if err != nil {
					log.Printf("error while checking for moditifications: %v", err)
				}
				newsize := int(fileinfo.Size())
				newmodtime := int(fileinfo.ModTime().Unix())

				if newsize != v[0] {
					modified = true
					monitor_list[k] = []int{newsize, newmodtime}
				} else {
					if newmodtime > v[1] {
						modified = true
						monitor_list[k] = []int{newsize, newmodtime}
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

func ListenForInput(inputChannel chan int) {
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

	initLogger()

	err := PreExec()
	if err != nil {
		log.Fatal("error during preexec:", err)
		return
	}

	filechannel := make(chan Event)
	inputchannel := make(chan int)

	go ListenForInput(inputchannel)
	go CheckIfModified(filechannel)

	err = ExecutionDriver(globalargs)
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			select {
			case data, ok := <-inputchannel:
				// fmt.Println(data)
				if !ok {
					log.Printf("Input channel closed")
					os.Exit(1)
				}

				switch data {
				case 0:
					KillProcess()
					close(inputchannel)
					close(filechannel)
					os.Exit(0)
				case 1:
					log.Printf("Restarting ... ")
					ExecutionDriver(globalargs)
					log.Printf("Restarted")
				case -1:
					log.Printf("unknown input;")
					runtimeCommandsLegend()
				}

			case data, ok := <-filechannel:
				if !ok {
					log.Printf("File channel is closed; exiting")
					os.Exit(1)
				}
				switch data.eventcode {
				case 101:
					ExecutionDriver(globalargs)
				}

			}
		}
	}()

	select {}
}
