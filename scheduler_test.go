package main

import (
	"testing"
	"time"
)

func TestGetOverdueTasks(t *testing.T) {
	past := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
	tasks := []Task{
		{ID: 1, Title: "Overdue", DueDate: past, Completed: false},
		{ID: 2, Title: "Not overdue", DueDate: "2099-01-01", Completed: false},
		{ID: 3, Title: "Completed", DueDate: past, Completed: true},
	}

	scheduler := NewScheduler(tasks)
	overdue := scheduler.GetOverdueTasks()

	if len(overdue) != 1 {
		t.Errorf("Expected 1 overdue task, got %d", len(overdue))
	}
}

func TestGetUpcomingTasks(t *testing.T) {
	tomorrow := time.Now().AddDate(0, 0, 1).Format("2006-01-02")
	nextWeek := time.Now().AddDate(0, 0, 5).Format("2006-01-02")
	farFuture := time.Now().AddDate(0, 0, 30).Format("2006-01-02")

	tasks := []Task{
		{ID: 1, Title: "Tomorrow", DueDate: tomorrow, Completed: false},
		{ID: 2, Title: "Next week", DueDate: nextWeek, Completed: false},
		{ID: 3, Title: "Far future", DueDate: farFuture, Completed: false},
	}

	scheduler := NewScheduler(tasks)
	upcoming := scheduler.GetUpcomingTasks(7)

	if len(upcoming) != 2 {
		t.Errorf("Expected 2 upcoming tasks, got %d", len(upcoming))
	}
}

func TestGetTaskStats(t *testing.T) {
	today := time.Now().Format("2006-01-02")
	past := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	tasks := []Task{
		{ID: 1, Title: "Completed", Completed: true},
		{ID: 2, Title: "Pending", Completed: false, DueDate: today},
		{ID: 3, Title: "Overdue", Completed: false, DueDate: past},
	}

	scheduler := NewScheduler(tasks)
	stats := scheduler.GetTaskStats()

	if stats["total"] != 3 {
		t.Errorf("Expected 3 total, got %d", stats["total"])
	}
	if stats["completed"] != 1 {
		t.Errorf("Expected 1 completed, got %d", stats["completed"])
	}
	if stats["overdue"] != 1 {
		t.Errorf("Expected 1 overdue, got %d", stats["overdue"])
	}
}

func TestGetProductivityScore(t *testing.T) {
	tasks := []Task{
		{ID: 1, Completed: true},
		{ID: 2, Completed: true},
		{ID: 3, Completed: false},
		{ID: 4, Completed: false},
	}

	scheduler := NewScheduler(tasks)
	score := scheduler.GetProductivityScore()

	if score != 50 {
		t.Errorf("Expected 50%%, got %.1f%%", score)
	}
}

func TestGetDailySummary(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Task 1", Completed: true},
		{ID: 2, Title: "Task 2", Completed: false},
	}

	scheduler := NewScheduler(tasks)
	summary := scheduler.GetDailySummary()

	if summary == "" {
		t.Error("Expected non-empty summary")
	}
}