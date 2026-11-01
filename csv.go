package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// csvHeader is the column order written by `csv-export` and required, in this
// order, by `csv-import`.
var csvHeader = []string{
	"id", "title", "description", "project", "priority",
	"status", "tags", "created_at", "due_date", "completed_at",
}

// tasksToCSV writes tasks as RFC 4180 CSV. Tags are joined with ";" because a
// comma is already the field separator; dates are RFC 3339.
func tasksToCSV(w io.Writer, tasks []Task) error {
	cw := csv.NewWriter(w)
	if err := cw.Write(csvHeader); err != nil {
		return err
	}

	for _, t := range tasks {
		record := []string{
			strconv.Itoa(t.ID),
			t.Title,
			t.Description,
			t.Project,
			string(t.Priority),
			string(t.Status),
			strings.Join(t.Tags, ";"),
			t.CreatedAt.Format(time.RFC3339),
			formatDatePointer(t.DueDate),
			formatDatePointer(t.CompletedAt),
		}
		if err := cw.Write(record); err != nil {
			return err
		}
	}

	cw.Flush()
	return cw.Error()
}

// tasksFromCSV reads tasks written by tasksToCSV. IDs in the file are ignored;
// startID assigns fresh ones so imports never collide with the local store.
func tasksFromCSV(r io.Reader, startID int) ([]Task, error) {
	cr := csv.NewReader(r)
	cr.FieldsPerRecord = len(csvHeader)

	header, err := cr.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	for i, column := range csvHeader {
		if strings.TrimSpace(strings.ToLower(header[i])) != column {
			return nil, fmt.Errorf("column %d should be %q, got %q", i+1, column, header[i])
		}
	}

	var tasks []Task
	next := startID
	for {
		record, err := cr.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row: %w", err)
		}

		created, err := time.Parse(time.RFC3339, record[7])
		if err != nil {
			created = time.Now()
		}

		task := Task{
			ID:          next,
			Title:       strings.TrimSpace(record[1]),
			Description: record[2],
			Project:     strings.TrimSpace(record[3]),
			Priority:    parsePriority(record[4]),
			Status:      parseStatus(record[5]),
			Tags:        splitTags(record[6]),
			CreatedAt:   created,
			UpdatedAt:   created,
			DueDate:     parseDatePointer(record[8]),
			CompletedAt: parseDatePointer(record[9]),
		}
		if task.Title == "" {
			return nil, fmt.Errorf("row %d has an empty title", next-startID+2)
		}
		tasks = append(tasks, task)
		next++
	}

	return tasks, nil
}

func splitTags(raw string) []string {
	var tags []string
	for _, tag := range strings.Split(raw, ";") {
		tag = strings.TrimSpace(tag)
		if tag != "" {
			tags = append(tags, tag)
		}
	}
	return tags
}

func formatDatePointer(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func parseDatePointer(raw string) *time.Time {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		if day, dayErr := time.Parse("2006-01-02", raw); dayErr == nil {
			return &day
		}
		return nil
	}
	return &parsed
}

var csvExportCmd = &cobra.Command{
	Use:   "csv-export [file]",
	Short: "Export all tasks to a CSV file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		store := loadStore()
		file, err := os.Create(args[0])
		if err != nil {
			fmt.Printf("\033[31mError: %s\033[0m\n", err)
			return
		}
		defer file.Close()

		if err := tasksToCSV(file, store.Tasks); err != nil {
			fmt.Printf("\033[31mError: %s\033[0m\n", err)
			return
		}
		fmt.Printf("\033[32mExported %d task(s) to %s\033[0m\n", len(store.Tasks), args[0])
	},
}

var csvImportCmd = &cobra.Command{
	Use:   "csv-import [file]",
	Short: "Import tasks from a CSV file",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		file, err := os.Open(args[0])
		if err != nil {
			fmt.Printf("\033[31mError: %s\033[0m\n", err)
			return
		}
		defer file.Close()

		store := loadStore()
		imported, err := tasksFromCSV(file, findNextID(store))
		if err != nil {
			fmt.Printf("\033[31mError: %s\033[0m\n", err)
			return
		}
		for _, task := range imported {
			store.Tasks = append(store.Tasks, task)
			if task.Project != "" {
				addProject(&store, task.Project)
			}
		}
		store.NextID = findNextID(store)
		if err := saveStore(store); err != nil {
			fmt.Printf("\033[31mError: %s\033[0m\n", err)
			return
		}
		fmt.Printf("\033[32mImported %d task(s) from %s\033[0m\n", len(imported), args[0])
	},
}
