package main

import (
	"context"
	"testing"
	"time"
)

func NewTestTask() *Task {
	return &Task{
		ElementAmount: 10,
		Delta:         1.25,
		FirstElement:  2.11,
		Interval:      0.1,
		TTL:           50,
	}
}

func TestPoolEnqueue(t *testing.T) {
	testPool := NewPool(1)
	testTask := NewTestTask()

	testPool.Enqueue(testTask)

	if testTask.id != 1 {
		t.Errorf("Expected id %d, got %d", 1, testTask.id)
	}

	if testPool.Queue.Len() < 1 {
		t.Errorf("Expected Queue lenght %d, got %d", 1, testPool.Queue.Len())
	}

	if testTask.Status != WAIT {
		t.Errorf("Expected status %s, got %s", WAIT, testTask.Status)
	}
}

func TestPoolDequeue(t *testing.T) {
	testPool := NewPool(1)
	testTask := NewTestTask()

	testPool.Queue.PushBack(1)
	testPool.TaskHistory[1] = testTask
	testPool.Dequeue(1)

	if testPool.Queue.Len() != 0 {
		t.Errorf("Expected Queue lenght %d, got %d", 0, testPool.Queue.Len())
	}
}

func TestPoolWorker(t *testing.T) {
	testPool := NewPool(1)
	testTask := NewTestTask()

	testPool.Enqueue(testTask)

	go testPool.Worker(context.Background())
	testPool.TaskChan <- testTask

	if testTask.Status != WAIT {
		t.Errorf("Expected status %s, got %s", WAIT, testTask.Status)
	}

	time.Sleep(time.Second * 2)

	if testTask.Status != DONE {
		t.Errorf("Expected status %s, got %s", DONE, testTask.Status)
	}

	if testPool.Queue.Len() > 0 {
		t.Errorf("Expected Queue lenght %d, got %d", 0, testPool.Queue.Len())
	}

	if testTask.QueuePosition > 0 {
		t.Errorf("Expected Queue position %d, got %d", 0, testTask.QueuePosition)
	}
}
