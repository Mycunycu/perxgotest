package main

import (
	"container/list"
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

const (
	WAIT    = "WAITING"
	PROCESS = "PROCESSING"
	DONE    = "DONE"
)

type Task struct {
	id               int
	QueuePosition    int     `json:"queue_position,omitempty"`
	Status           string  `json:"status"`
	ElementAmount    int     `json:"element_amount"`
	Delta            float32 `json:"delta"`
	FirstElement     float32 `json:"first_element"`
	Interval         float32 `json:"interval"`
	TTL              float32 `json:"ttl"`
	CurrentIteration int     `json:"current_iteration,omitempty"`
	EnqueueTime      string  `json:"enqueue_time"`
	StartTime        string  `json:"start_time,omitempty"`
	DoneTime         string  `json:"done_time,omitempty"`
}

type PoolWorkers struct {
	TaskId      int
	MaxWorkers  int
	Queue       *list.List
	TaskHistory map[int]*Task
	TaskChan    chan *Task
	wg          *sync.WaitGroup
}

func NewPoolWorkers(maxWorkers int) *PoolWorkers {
	return &PoolWorkers{
		MaxWorkers:  maxWorkers,
		Queue:       list.New(),
		TaskHistory: make(map[int]*Task),
		TaskChan:    make(chan *Task, maxWorkers),
		wg:          &sync.WaitGroup{},
	}
}

func (pw *PoolWorkers) WorkersRun() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	for i := 0; i < pw.MaxWorkers; i++ {
		pw.wg.Add(1)
		fmt.Printf("Worker %d waiting a tasks\n", i+1)
		go pw.Worker(ctx)
	}

	go func() {
		pool.wg.Wait()
		fmt.Print("Graceful shutdown (:")
		os.Exit(0)
	}()
}

func (pw *PoolWorkers) Worker(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			fmt.Println("Signal to cancel received")
			pw.wg.Done()
			return
		case t := <-pw.TaskChan:
			fmt.Printf("Worker started work on task: %d\n", t.id)
			t.Status = PROCESS
			t.StartTime = time.Now().Format(time.Stamp)

			result := t.FirstElement
			current := t.FirstElement

			for i := 1; i < t.ElementAmount; i++ {
				t.CurrentIteration = i

				next := current + t.Delta
				result += next
				current = next

				time.Sleep(time.Duration(t.Interval*1000) * time.Millisecond)
			}

			t.Status = DONE
			t.CurrentIteration = 0
			t.DoneTime = time.Now().Format(time.Stamp)

			time.AfterFunc(time.Duration(t.TTL*1000)*time.Millisecond, func() { pw.RemoveTaskHistory(t.id) })

			pw.Dequeue(t.id)
			pw.SyncQueuePosition()

			fmt.Printf("Worker ended on task %d and result: %f\n", t.id, result)
		}
	}
}

func (pw *PoolWorkers) Enqueue(task *Task) {
	task.id = taskId
	taskId++

	fmt.Printf("Enqueue Task with id: %d\n", task.id)

	pw.Queue.PushBack(task.id)
	pw.TaskHistory[task.id] = task

	task.QueuePosition = pw.Queue.Len()
	task.Status = WAIT
	task.EnqueueTime = time.Now().Format(time.Stamp)

	go func() {
		pw.TaskChan <- task
	}()
}

func (pw *PoolWorkers) Dequeue(id int) {
	fmt.Printf("Dequeue Task with id: %d\n", id)

	for e := pw.Queue.Front(); e != nil; e = e.Next() {
		if e.Value.(int) == id {
			pw.Queue.Remove(e)
			continue
		}
	}

	task := pw.TaskHistory[id]
	task.QueuePosition = 0
}

func (pw *PoolWorkers) RemoveTaskHistory(id int) {
	fmt.Printf("Removing Task with id: %d from history by lifetime", id)
	delete(pw.TaskHistory, id)
}

func (pw *PoolWorkers) SyncQueuePosition() {
	position := 1

	for e := pw.Queue.Front(); e != nil; e = e.Next() {
		id := e.Value.(int)
		task, ok := pw.TaskHistory[id]

		if !ok {
			fmt.Printf("Task with id %d don't find in history map", id)
			continue
		}

		task.QueuePosition = position
		position++
	}
}
