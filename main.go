package main

import (
	"log"
	"os"

	"github.com/nn-advith/mantis/mantis"
)

func main() {

	mantis.InitLogger()

	err := mantis.PreExec()
	if err != nil {
		log.Fatal("error during preexec:", err)
		return
	}

	filechannel := make(chan mantis.Event)
	inputchannel := make(chan int)

	go mantis.ListenForInput(inputchannel)
	go mantis.CheckIfModified(filechannel)

	err = mantis.ExecutionDriver()
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
					mantis.KillProcess()
					close(inputchannel)
					close(filechannel)
					os.Exit(0)
				case 1:
					log.Printf("Restarting ... ")
					mantis.ExecutionDriver()
					log.Printf("Restarted")
				case -1:
					log.Printf("unknown input;")
					mantis.RuntimeCommandsLegend()
				}

			case data, ok := <-filechannel:
				if !ok {
					log.Printf("File channel is closed; exiting")
					os.Exit(1)
				}
				switch data.EventCode {
				case 101:
					mantis.ExecutionDriver()
				}

			}
		}
	}()

	select {}
}
