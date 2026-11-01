package main

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

func sampleTasks() []Task {
	due := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	created := time.Date(2026, 8, 1, 9, 30, 0, 0, time.UTC)
	return []Task{
		{
			ID: 1, Title: "Write report", Description: "Quarterly, with charts",
			Project: "work", Priority: PriorityHigh, Status: StatusTodo,
			Tags: []string{"docs", "urgent"}, CreatedAt: created, UpdatedAt: created,
			DueDate: &due,
		},
		{
			ID: 2, Title: "Buy milk", Priority: PriorityLow, Status: StatusDone,
			CreatedAt: created, UpdatedAt: created,
		},
	}
}

func TestCSVRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	if err := tasksToCSV(&buf, sampleTasks()); err != nil {
		t.Fatal(err)
	}

	imported, err := tasksFromCSV(&buf, 10)
	if err != nil {
		t.Fatal(err)
	}
	if len(imported) != 2 {
		t.Fatalf("expected 2 tasks, got %d", len(imported))
	}

	first := imported[0]
	if first.ID != 10 || imported[1].ID != 11 {
		t.Errorf("IDs should be renumbered from 10, got %d and %d", first.ID, imported[1].ID)
	}
	if first.Title != "Write report" || first.Project != "work" || first.Priority != PriorityHigh {
		t.Errorf("fields lost on round trip: %+v", first)
	}
	if len(first.Tags) != 2 || first.Tags[0] != "docs" {
		t.Errorf("tags lost: %v", first.Tags)
	}
	if first.DueDate == nil || first.DueDate.Year() != 2026 {
		t.Errorf("due date lost: %v", first.DueDate)
	}
	if !first.CreatedAt.Equal(time.Date(2026, 8, 1, 9, 30, 0, 0, time.UTC)) {
		t.Errorf("created_at lost timezone: %v", first.CreatedAt)
	}
}

func TestCSVQuotesCommaseparatedValues(t *testing.T) {
	var buf bytes.Buffer
	if err := tasksToCSV(&buf, []Task{{
		ID: 1, Title: "Fix a, b and c", Description: `He said "go"`,
		CreatedAt: time.Now(),
	}}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(buf.String(), `"Fix a, b and c"`) {
		t.Errorf("comma in a title was not quoted:\n%s", buf.String())
	}
	if ts, err := tasksFromCSV(&buf, 1); err != nil || len(ts) != 1 || ts[0].Description != `He said "go"` {
		t.Errorf("quoted values did not round trip: %v %v", ts, err)
	}
}

func TestCSVEmptyListWritesHeaderOnly(t *testing.T) {
	var buf bytes.Buffer
	if err := tasksToCSV(&buf, nil); err != nil {
		t.Fatal(err)
	}
	tasks, err := tasksFromCSV(&buf, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(tasks) != 0 {
		t.Errorf("expected no tasks, got %d", len(tasks))
	}
}

func TestCSVRejectsWrongHeader(t *testing.T) {
	input := "name,title\nfoo,bar\n"
	if _, err := tasksFromCSV(strings.NewReader(input), 1); err == nil {
		t.Fatal("expected a header error")
	}
}

func TestCSVRejectsEmptyTitle(t *testing.T) {
	var buf bytes.Buffer
	if err := tasksToCSV(&buf, []Task{{ID: 1, Title: "ok", CreatedAt: time.Now()}}); err != nil {
		t.Fatal(err)
	}
	// Blank out the title of the data row.
	rows := strings.Split(buf.String(), "\n")
	rows[1] = "1,,,work,high,,docs,2026-08-01T09:30:00Z,,"
	input := strings.Join(rows, "\n")
	if _, err := tasksFromCSV(strings.NewReader(input), 1); err == nil {
		t.Fatal("expected an empty-title error")
	}
}

func TestCSVImportDefaultsUnknownPriority(t *testing.T) {
	var buf bytes.Buffer
	if err := tasksToCSV(&buf, []Task{{ID: 1, Title: "t", Priority: "nonsense", CreatedAt: time.Now()}}); err != nil {
		t.Fatal(err)
	}
	tasks, err := tasksFromCSV(&buf, 1)
	if err != nil {
		t.Fatal(err)
	}
	if tasks[0].Priority != PriorityMedium {
		t.Errorf("expected the default priority, got %q", tasks[0].Priority)
	}
}

func TestCSVAcceptsDateOnlyDueDates(t *testing.T) {
	var buf bytes.Buffer
	tasksToCSV(&buf, []Task{{ID: 1, Title: "t", CreatedAt: time.Now()}})
	rows := strings.Split(buf.String(), "\n")
	rows[1] = "1,t,,,medium,todo,,,2026-09-05,"
	if _, err := tasksFromCSV(strings.NewReader(strings.Join(rows, "\n")), 1); err != nil {
		t.Fatal(err)
	}
}

func TestCSVCommandsRoundTripThroughTheStore(t *testing.T) {
	withStore(t)

	store := loadStore()
	store.Tasks = append(store.Tasks, sampleTasks()...)
	store.NextID = findNextID(store)
	if err := saveStore(store); err != nil {
		t.Fatal(err)
	}

	csvPath := t.TempDir() + "/tasks.csv"
	runCmd(t, "csv-export", csvPath)

	// Wipe the store, then import into an empty one.
	if err := saveStore(TaskStore{NextID: 1}); err != nil {
		t.Fatal(err)
	}
	runCmd(t, "csv-import", csvPath)

	reloaded := loadStore()
	if len(reloaded.Tasks) != 2 {
		t.Fatalf("expected 2 tasks after import, got %d", len(reloaded.Tasks))
	}
	if reloaded.Tasks[0].Title != "Write report" {
		t.Errorf("unexpected first task: %+v", reloaded.Tasks[0])
	}
	if reloaded.NextID != 3 {
		t.Errorf("NextID = %d, want 3", reloaded.NextID)
	}
}

func TestCSVExportFailsOnBadPath(t *testing.T) {
	withStore(t)
	err := tasksToCSV(&failingWriter{}, sampleTasks())
	if err == nil {
		t.Fatal("expected the writer error to surface")
	}
}

type failingWriter struct{}

func (f *failingWriter) Write(p []byte) (int, error) { return 0, errWrite }

var errWrite = &writeError{}

type writeError struct{}

func (w *writeError) Error() string { return "disk full" }
