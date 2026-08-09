package main

import (
	"os"
	"testing"
	"time"
)

func TestLoadTasks(t *testing.T) {
	os.Remove("tasks.json")
	tasks := loadTasks()
	if len(tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestSaveAndLoadTasks(t *testing.T) {
	os.Remove("tasks.json")
	tasks := []Task{
		{ID: 1, Title: "Test Task", Priority: "3", Completed: false, CreatedAt: time.Now()},
		{ID: 2, Title: "Done Task", Priority: "1", Completed: true, CreatedAt: time.Now()},
	}
	saveTasks(tasks)
	loaded := loadTasks()
	if len(loaded) != 2 {
		t.Errorf("Expected 2 tasks, got %d", len(loaded))
	}
	if loaded[0].Title != "Test Task" {
		t.Errorf("Expected 'Test Task', got '%s'", loaded[0].Title)
	}
	os.Remove("tasks.json")
}

func TestSortByPriority(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Low", Priority: "1"},
		{ID: 2, Title: "High", Priority: "5"},
		{ID: 3, Title: "Medium", Priority: "3"},
	}
	sortByPriority(tasks)
	if tasks[0].Priority != "5" {
		t.Errorf("Expected highest priority first, got %s", tasks[0].Priority)
	}
}

func TestSortByDueDate(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Later", DueDate: "2099-01-01"},
		{ID: 2, Title: "Sooner", DueDate: "2026-01-01"},
	}
	sortByDueDate(tasks)
	if tasks[0].DueDate != "2026-01-01" {
		t.Errorf("Expected sooner date first, got %s", tasks[0].DueDate)
	}
}

func TestPrintStats(t *testing.T) {
	tasks := []Task{
		{ID: 1, Completed: true},
		{ID: 2, Completed: false},
		{ID: 3, Completed: false},
	}
	printStats(tasks)
}

func TestPrintTasks(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Task 1", Priority: "3", Completed: false},
	}
	printTasks(tasks)
}
