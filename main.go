package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"net/http"
	"sort"
)

var pool *PoolWorkers
var taskId = 1

func main() {
	err := appRun()

	if err != nil {
		log.Fatal(err.Error())
	}
}

func appRun() error {
	var maxWorkers int

	flag.IntVar(&maxWorkers, "N", 1, "max workers at the same time")
	flag.Parse()

	if maxWorkers < 1 {
		return fmt.Errorf("count of workers should be more than zero")
	}

	pool = NewPoolWorkers(maxWorkers)
	pool.WorkersRun()

	http.HandleFunc("/api/v1/task", taskProcess)

	fmt.Printf("Starting server at port 8080\n")
	err := http.ListenAndServe(":8080", nil)

	if err != nil {
		return err
	}

	return nil
}

func taskProcess(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/api/v1/task" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	switch r.Method {
	case "GET":
		result := make([]Task, 0)
		var fulfilled []Task

		for _, v := range pool.TaskHistory {
			if v.Status == DONE {
				fulfilled = append(fulfilled, *v)
			}

			if v.Status != DONE {
				result = append(result, *v)
			}
		}

		sort.Slice(result, func(i, j int) bool {
			return result[i].QueuePosition < result[j].QueuePosition
		})

		sort.Slice(fulfilled, func(i, j int) bool {
			return fulfilled[i].EnqueueTime < fulfilled[j].EnqueueTime
		})

		result = append(result, fulfilled...)

		w.Header().Add("content-type", "application/json")
		json.NewEncoder(w).Encode(result)
	case "POST":
		var task Task

		err := json.NewDecoder(r.Body).Decode(&task)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if task.ElementAmount < 1 {
			err := errors.New("numbers of elements should be at least one")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if task.Interval < 0 {
			err := errors.New("interval should be more than zero")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if task.TTL < 0 {
			err := errors.New("lifetime shouldn't be negative")
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		pool.Enqueue(&task)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Task queued successfully."))
	default:
		fmt.Fprintf(w, "Only GET and POST methods are supported.")
	}
}
