package main

import (
	"fmt"
	"strings"
)

// SearchTasks searches tasks by title or description
func SearchTasks(tasks []Task, query string) []Task {
	var results []Task
	query = strings.ToLower(query)
	for _, task := range tasks {
		if strings.Contains(strings.ToLower(task.Title), query) ||
			strings.Contains(strings.ToLower(task.Description), query) {
			results = append(results, task)
		}
	}
	return results
}

// SearchByTag searches tasks by tag
func SearchByTag(tasks []Task, tag string) []Task {
	var results []Task
	tag = strings.ToLower(tag)
	for _, task := range tasks {
		for _, t := range task.Tags {
			if strings.ToLower(t) == tag {
				results = append(results, task)
				break
			}
		}
	}
	return results
}

// SearchByPriority searches tasks by priority level
func SearchByPriority(tasks []Task, priority string) []Task {
	var results []Task
	for _, task := range tasks {
		if task.Priority == priority {
			results = append(results, task)
		}
	}
	return results
}

// SearchByDueDate searches tasks due within specified days
func SearchByDueDate(tasks []Task, days int) []Task {
	var results []Task
	for _, task := range tasks {
		if task.DueDate != "" && !task.Completed {
			results = append(results, task)
		}
	}
	return results
}

// FormatSearchResults formats search results for display
func FormatSearchResults(tasks []Task, query string) string {
	if len(tasks) == 0 {
		return fmt.Sprintf("No tasks found matching '%s'", query)
	}
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Found %d task(s) matching '%s':\n\n", len(tasks), query))
	for _, task := range tasks {
		status := "[ ]"
		if task.Completed {
			status = "[x]"
		}
		sb.WriteString(fmt.Sprintf("%s #%d [%s] %s\n", status, task.ID, task.Priority, task.Title))
		if task.Description != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", task.Description))
		}
		if len(task.Tags) > 0 {
			sb.WriteString(fmt.Sprintf("    Tags: %s\n", strings.Join(task.Tags, ", ")))
		}
		if task.DueDate != "" {
			sb.WriteString(fmt.Sprintf("    Due: %s\n", task.DueDate))
		}
	}
	return sb.String()
}
