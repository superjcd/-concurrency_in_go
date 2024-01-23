package main

import (
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"
)

const workers = 4

func main() {
	workerChans := make([]chan struct{}, 4)

	for i := 0; i < workers; i++ {
		workerChans[i] = make(chan struct{}, 1)
	}

	for i := 0; i < workers; i++ {
		go func(i int) {
			for {
				<-workerChans[i]
				fmt.Println(i + 1)
				time.Sleep(time.Second)
				workerChans[(i+1)%workers] <- struct{}{}
			}
		}(i)
	}

	workerChans[0] <- struct{}{}
	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}()
	time.Sleep(time.Second * 60)
}
