package main

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	_ "net/http/pprof"
	"time"
)

const (
	MaxWorkers = 100
	MaxQueue   = 1000
)

var JobQueue chan Job
var DemoPayloads [10000]Payload

type Job struct {
	Payload Payload
}

// 模拟
func PayloadHandler() {

	// Go through each payload and queue items individually to be posted to Worker
	for _, payload := range DemoPayloads {

		// let's create a job with the payload
		work := Job{Payload: payload}

		// Push the work onto the queue.
		JobQueue <- work             // 程序会阻塞在这里
		time.Sleep(time.Microsecond) // 每分钟6万次
	}
}

type Payload struct {
	Data string
}

func (p *Payload) UploadToS3() error {
	// 模拟传输到S3的耗时操作
	fmt.Printf("Uploading data: %s \n", p.Data)
	// time.Sleep(time.Second * 1)
	return nil
}

type Worker struct {
	quit chan bool
}

func NewWorker() Worker {
	return Worker{
		quit: make(chan bool)}
}

func (w Worker) Start() {
	go func() {
		for {
			select {
			case job := <-JobQueue:
				// we have received a work request.
				if err := job.Payload.UploadToS3(); err != nil {
					fmt.Println("Error uploading to S3:" + err.Error())
				}

			case <-w.quit:
				// we have received a signal to stop
				return
			}
		}
	}()
}

func (w Worker) Stop() {
	go func() {
		w.quit <- true
	}()
}

func init() {
	data := [4]string{
		"cat",
		"mouse",
		"dog",
		"bird",
	}
	for i := 0; i < 10000; i++ {
		DemoPayloads[i] = Payload{Data: data[rand.Intn(4)]}
	}

	JobQueue = make(chan Job, MaxQueue)
}

func main() {
	// 模拟Handler接受请求
	go PayloadHandler()

	for i := 0; i < MaxWorkers; i++ {
		worker := NewWorker()

		worker.Start()
	}

	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatal(err)
		}
	}()
	select {}
}
