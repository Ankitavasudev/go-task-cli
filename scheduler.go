package main

import (
	"fmt"
	"sort"
	"time"
)

type Scheduler struct {
	tasks []Task
}

func NewScheduler(tasks []Task) *Scheduler {
	return &Scheduler{tasks: tasks}
}

func (s *Scheduler) GetOverdueTasks() []Task {
	var overdue []Task
	now := time.Now()

	for _, task := range s.tasks {
		if task.DueDate != "" && !task.Completed {
			due, err := time.Parse("2006-01-02", task.DueDate)
			if err == nil && now.After(due) {
				overdue = append(overdue, task)
			}
		}
	}

	return overdue
}

func (s *Scheduler) GetUpcomingTasks(days int) []Task {
	var upcoming []Task
	now := time.Now()
	future := now.AddDate(0, 0, days)

	for _, task := range s.tasks {
		if task.DueDate != "" && !task.Completed {
			due, err := time.Parse("2006-01-02", task.DueDate)
			if err == nil && due.After(now) && due.Before(future) {
				upcoming = append(upcoming, task)
			}
		}
	}

	sort.Slice(upcoming, func(i, j int) bool {
		dueI, _ := time.Parse("2006-01-02", upcoming[i].DueDate)
		dueJ, _ := time.Parse("2006-01-02", upcoming[j].DueDate)
		return dueI.Before(dueJ)
	})

	return upcoming
}

func (s *Scheduler) GetTaskStats() map[string]int {
	stats := map[string]int{
		"total":     0,
		"completed": 0,
		"pending":   0,
		"overdue":   0,
		"today":     0,
	}

	now := time.Now()
	today := now.Format("2006-01-02")

	for _, task := range s.tasks {
		stats["total"]++
		if task.Completed {
			stats["completed"]++
		} else {
			stats["pending"]++
			if task.DueDate == today {
				stats["today"]++
			}
			if task.DueDate != "" {
				due, err := time.Parse("2006-01-02", task.DueDate)
				if err == nil && now.After(due) {
					stats["overdue"]++
				}
			}
		}
	}

	return stats
}

func (s *Scheduler) GetProductivityScore() float64 {
	if len(s.tasks) == 0 {
		return 0
	}

	completed := 0
	for _, task := range s.tasks {
		if task.Completed {
			completed++
		}
	}

	return float64(completed) / float64(len(s.tasks)) * 100
}

func (s *Scheduler) GetDailySummary() string {
	stats := s.GetTaskStats()
	overdue := s.GetOverdueTasks()
	upcoming := s.GetUpcomingTasks(7)
	score := s.GetProductivityScore()

	summary := "Daily Summary\n"
	summary += "=============\n"
	summary += fmt.Sprintf("Total tasks: %d\n", stats["total"])
	summary += fmt.Sprintf("Completed: %d\n", stats["completed"])
	summary += fmt.Sprintf("Pending: %d\n", stats["pending"])
	summary += fmt.Sprintf("Overdue: %d\n", stats["overdue"])
	summary += fmt.Sprintf("Due today: %d\n", stats["today"])
	summary += fmt.Sprintf("Due this week: %d\n", len(upcoming))
	summary += fmt.Sprintf("Productivity score: %.1f%%\n", score)

	if len(overdue) > 0 {
		summary += "\nOverdue Tasks:\n"
		for _, task := range overdue {
			summary += fmt.Sprintf("  - [%s] %s (due: %s)\n", task.Priority, task.Title, task.DueDate)
		}
	}

	if len(upcoming) > 0 {
		summary += "\nUpcoming Tasks:\n"
		for _, task := range upcoming {
			summary += fmt.Sprintf("  - [%s] %s (due: %s)\n", task.Priority, task.Title, task.DueDate)
		}
	}

	return summary
}