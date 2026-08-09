package main

import (
	"os"
	"strings"
	"testing"
)

func TestExportToCSV(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Test Task", Description: "Desc", Priority: "3", Tags: []string{"bug", "urgent"}, DueDate: "2026-08-15"},
		{ID: 2, Title: "Another Task", Description: "Desc2", Priority: "1", Tags: []string{"feature"}},
	}

	tmpFile := "test_export.csv"
	defer os.Remove(tmpFile)

	err := ExportToCSV(tasks, tmpFile)
	if err != nil {
		t.Fatalf("ExportToCSV failed: %v", err)
	}

	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Fatal("Export file was not created")
	}
}

func TestExportToCSVWithCommas(t *testing.T) {
	tasks := []Task{
		{ID: 1, Title: "Task", Description: "Fix bug, add feature", Priority: "3"},
	}

	tmpFile := "test_comma.csv"
	defer os.Remove(tmpFile)

	err := ExportToCSV(tasks, tmpFile)
	if err != nil {
		t.Fatalf("ExportToCSV failed: %v", err)
	}

	tasks2, err := ImportFromCSV(tmpFile)
	if err != nil {
		t.Fatalf("ImportFromCSV failed: %v", err)
	}

	if len(tasks2) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks2))
	}

	if !strings.Contains(tasks2[0].Description, ",") {
		t.Errorf("Description lost comma: %s", tasks2[0].Description)
	}
}

func TestImportFromCSV(t *testing.T) {
	content := "ID,Title,Description,Priority,Tags,DueDate,Completed,CreatedAt,CompletedAt\n1,Test Task,Desc,3,bug;urgent,2026-08-15,false,2026-08-08T00:00:00Z,\n"
	tmpFile := "test_import.csv"
	os.WriteFile(tmpFile, []byte(content), 0644)
	defer os.Remove(tmpFile)

	tasks, err := ImportFromCSV(tmpFile)
	if err != nil {
		t.Fatalf("ImportFromCSV failed: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("Expected 1 task, got %d", len(tasks))
	}
	if tasks[0].Title != "Test Task" {
		t.Errorf("Expected title 'Test Task', got '%s'", tasks[0].Title)
	}
}

func TestImportFromCSV_Empty(t *testing.T) {
	content := "ID,Title,Description,Priority,Tags,DueDate,Completed,CreatedAt,CompletedAt\n"
	tmpFile := "test_empty.csv"
	os.WriteFile(tmpFile, []byte(content), 0644)
	defer os.Remove(tmpFile)

	tasks, err := ImportFromCSV(tmpFile)
	if err != nil {
		t.Fatalf("ImportFromCSV failed: %v", err)
	}
	if len(tasks) != 0 {
		t.Fatalf("Expected 0 tasks, got %d", len(tasks))
	}
}

func TestRoundTrip(t *testing.T) {
	original := []Task{
		{ID: 1, Title: "Task 1", Description: "D1", Priority: "5", Tags: []string{"a", "b"}, DueDate: "2026-08-20"},
	}

	tmpFile := "test_roundtrip.csv"
	defer os.Remove(tmpFile)

	if err := ExportToCSV(original, tmpFile); err != nil {
		t.Fatalf("Export failed: %v", err)
	}

	imported, err := ImportFromCSV(tmpFile)
	if err != nil {
		t.Fatalf("Import failed: %v", err)
	}

	if len(imported) != len(original) {
		t.Fatalf("Expected %d tasks, got %d", len(original), len(imported))
	}
	if imported[0].Title != original[0].Title {
		t.Errorf("Title mismatch: %s vs %s", imported[0].Title, original[0].Title)
	}
}
