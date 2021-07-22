package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"time"
)

var MaxWorkers int
var Queue []*Task
var workerChan = make(chan *Task, MaxWorkers)

type TaskStatus int

func (d TaskStatus) String() string {
	return [...]string{"UNDEFINED", "WAITING", "PROCESSING", "DONE"}[d]
}

const (
	Wait TaskStatus = iota + 1
	Process
	Done
)

type Task struct {
	QueuePosition    int
	Status           TaskStatus
	ElementAmount    int     `json:"element_amount"`
	Delta            float32 `json:"delta"`
	FirstElement     float32 `json:"first_element"`
	Interval         float32 `json:"interval"`
	TTL              float32 `json:"ttl"`
	CurrentIteration int
	EnqueueTime      string
	StartTime        string
	DoneTime         string
}

func WorkersRun(maxWorkers int) {
	for i := 0; i < maxWorkers; i++ {
		fmt.Printf("Worker %d started", i+1)
		go Worker()
	}
}

func Worker() {
	for t := range workerChan {
		t.Status = 2
		t.StartTime = time.Now().Format(time.Stamp)
		log.Println(t)
	}
}

func taskProcess(w http.ResponseWriter, r *http.Request) {

	if r.URL.Path != "/api/v1/task" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":

	case "POST":
		var task Task

		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		Queue = append(Queue, &task)

		task.QueuePosition = len(Queue)
		task.Status = 1
		task.EnqueueTime = time.Now().Format(time.Stamp)

		for _, t := range Queue {
			if t.Status == 1 {
				workerChan <- t
			}
		}

	default:
		fmt.Fprintf(w, "Only GET and POST methods are supported.")
	}
}

func main() {
	flag.IntVar(&MaxWorkers, "N", 1, "max workers at the same time")
	flag.Parse()

	if MaxWorkers < 1 {
		log.Fatal("Count of workers should be more than zero")
	}

	WorkersRun(MaxWorkers)

	http.HandleFunc("/api/v1/task", taskProcess)

	fmt.Printf("Starting server at port 8080\n")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
