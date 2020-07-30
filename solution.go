package main

import (
	"flag"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

var workers int
var endpoints string

type result struct {
	m         sync.RWMutex
	urlsDone  int
	urlsError int
}

func main() {
	flag.IntVar(&workers, "workers", 5, "number of workers")
	flag.StringVar(&endpoints, "endpoints", "", "list of endpoints comma delimited")
	flag.Parse()

	res := result{}

	tasks := make(chan string)
	successChannel := make(chan int)
	errorChannel := make(chan int)
	done := make(chan bool)

	splittedEndpoints := strings.Split(endpoints, ",")

	for i := 0; i < workers; i++ {
		go worker(tasks, successChannel, errorChannel)
	}

	go func() {
		for i := 0; i < len(splittedEndpoints); i++ {
			select {
			case <-successChannel:
				res.m.Lock()
				res.urlsDone = res.urlsDone + 1
				res.m.Unlock()
				continue
			case <-errorChannel:
				res.m.Lock()
				res.urlsError = res.urlsError + 1
				res.m.Unlock()
				continue
			case <-time.After(time.Duration(len(splittedEndpoints)) * time.Second):
				fmt.Println("Timeout")
				break
			}
		}

		done <- true
	}()

	for _, endpoint := range splittedEndpoints {
		tasks <- endpoint
	}

	<-done

	close(tasks)
	close(successChannel)
	close(errorChannel)
	fmt.Printf("work done. successfully=%d errors=%d", res.urlsDone, res.urlsError)
}

func worker(urls chan string, successChannel, errorChannel chan int) error {
	for msg := range urls {
		_, err := http.Get(msg)
		if err != nil {
			errorChannel <- 1
			fmt.Printf("Error on getting url: %s. Detailed error: \n%s\n", msg, err.Error())
			continue
		}

		successChannel <- 1
	}

	return nil
}
