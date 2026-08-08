package main

import (
	"encoding/csv"
	"fmt"
	"os"
	"strings"
)

// ExportToCSV exports tasks to a CSV file
func ExportToCSV(tasks []Task, filename string) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	// Header
	header := []string{"ID", "Title", "Description", "Priority", "Tags", "DueDate", "Completed", "CreatedAt", "CompletedAt"}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Data
	for _, task := range tasks {
		completed := "false"
		if task.Completed {
			completed = "true"
		}
		record := []string{
			fmt.Sprintf("%d", task.ID),
			task.Title,
			task.Description,
			task.Priority,
			strings.Join(task.Tags, ";"),
			task.DueDate,
			completed,
			task.CreatedAt.Format("2006-01-02T15:04:05Z"),
			"",
		}
		if task.CompletedAt != nil {
			record[8] = task.CompletedAt.Format("2006-01-02T15:04:05Z")
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write record: %w", err)
		}
	}

	fmt.Printf("Exported %d tasks to %s\n", len(tasks), filename)
	return nil
}

// ImportFromCSV imports tasks from a CSV file
func ImportFromCSV(filename string) ([]Task, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("failed to read CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or has no data rows")
	}

	var tasks []Task
	for i, record := range records {
		if i == 0 {
			continue // Skip header
		}
		if len(record) < 7 {
			continue
		}
		task := Task{
			Title:       record[1],
			Description: record[2],
			Priority:    record[3],
			Tags:        strings.Split(record[4], ";"),
			DueDate:     record[5],
			Completed:   record[6] == "true",
		}
		tasks = append(tasks, task)
	}

	fmt.Printf("Imported %d tasks from %s\n", len(tasks), filename)
	return tasks, nil
}