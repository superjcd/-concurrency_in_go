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
	MaxQueue   = 500
)

var JobQueue chan Job

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

var DemoPayloads [10000]Payload

type Payload struct {
	Data string
}

func (p *Payload) UploadToS3() error {
	// 模拟传输到S3的耗时操作
	fmt.Printf("Uploading data: %s\n", p.Data)
	time.Sleep(time.Second * 1)
	return nil
}

type Worker struct {
	WorkerPool chan chan Job
	JobChannel chan Job
	quit       chan bool
}

func NewWorker(workerPool chan chan Job) Worker {
	return Worker{
		WorkerPool: workerPool,
		JobChannel: make(chan Job),
		quit:       make(chan bool)}
}

func (w Worker) Start() {
	go func() {
		for {
			w.WorkerPool <- w.JobChannel
			fmt.Println("len of JobChan:", len(w.WorkerPool))
			select {
			case job := <-w.JobChannel:
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

type Dispatcher struct {
	WorkerPool chan chan Job
}

func NewDispatcher(maxWorkers int) *Dispatcher {
	pool := make(chan chan Job, maxWorkers)
	return &Dispatcher{WorkerPool: pool}
}

func (d *Dispatcher) Run() {
	for i := 0; i < MaxWorkers; i++ {
		worker := NewWorker(d.WorkerPool)
		worker.Start()
	}

	go d.dispatch()
}

func (d *Dispatcher) dispatch() {
	for {
		select {
		case job := <-JobQueue:
			// a job request has been received
			go func(job Job) {
				// try to obtain a worker job channel that is available.
				// this will block until a worker is idle
				jobChannel := <-d.WorkerPool

				// dispatch the job to the worker job channel
				jobChannel <- job
			}(job)
		}
	}
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

	dispathcer := NewDispatcher(MaxWorkers)
	dispathcer.Run()

	go func() {
		if err := http.ListenAndServe(":6060", nil); err != nil {
			log.Fatal(err)
		}
	}()
	select {}
}
