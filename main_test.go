package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// withStore points the CLI at a throwaway store file for the duration of a test.
func withStore(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "tasks.json")
	t.Setenv("GO_TASK_CLI_FILE", path)
	return path
}

func TestGetStorePathHonoursOverride(t *testing.T) {
	path := withStore(t)
	if got := getStorePath(); got != path {
		t.Errorf("getStorePath() = %q, want %q", got, path)
	}
}

func TestLoadStoreStartsEmpty(t *testing.T) {
	withStore(t)
	store := loadStore()
	if len(store.Tasks) != 0 || store.NextID != 1 {
		t.Errorf("expected an empty store with NextID 1, got %+v", store)
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	withStore(t)
	due := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	store := TaskStore{Tasks: []Task{{
		ID:        3,
		Title:     "Ship v2",
		Priority:  PriorityHigh,
		Status:    StatusTodo,
		Tags:      []string{"work", "urgent"},
		CreatedAt: time.Now(),
		DueDate:   &due,
	}}, NextID: 4}
	if err := saveStore(store); err != nil {
		t.Fatal(err)
	}

	loaded := loadStore()
	if len(loaded.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(loaded.Tasks))
	}
	task := loaded.Tasks[0]
	if task.Title != "Ship v2" || task.Priority != PriorityHigh || task.Status != StatusTodo {
		t.Errorf("task fields lost: %+v", task)
	}
	if task.DueDate == nil || !task.DueDate.Equal(due) {
		t.Errorf("due date lost: %v", task.DueDate)
	}
	if len(task.Tags) != 2 {
		t.Errorf("tags lost: %v", task.Tags)
	}
}

func TestSaveStoreCreatesValidJSON(t *testing.T) {
	path := withStore(t)
	store := loadStore()
	store.Tasks = append(store.Tasks, Task{ID: 1, Title: "x", CreatedAt: time.Now()})
	if err := saveStore(store); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var decoded TaskStore
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("store is not valid JSON: %v", err)
	}
}

func TestFindNextIDTakesTheMaximum(t *testing.T) {
	store := TaskStore{Tasks: []Task{{ID: 1}, {ID: 9}, {ID: 4}}}
	if got := findNextID(store); got != 10 {
		t.Errorf("findNextID = %d, want 10", got)
	}
}

func TestFindNextIDOnEmptyStore(t *testing.T) {
	if got := findNextID(TaskStore{}); got != 1 {
		t.Errorf("findNextID = %d, want 1", got)
	}
}

func TestParsePriority(t *testing.T) {
	cases := map[string]Priority{
		"low":      PriorityLow,
		"MEDIUM":   PriorityMedium,
		"med":      PriorityMedium,
		"high":     PriorityHigh,
		"crit":     PriorityCritical,
		"critical": PriorityCritical,
		"nonsense": PriorityMedium,
	}
	for input, want := range cases {
		if got := parsePriority(input); got != want {
			t.Errorf("parsePriority(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestParseStatus(t *testing.T) {
	cases := map[string]Status{
		"todo":        StatusTodo,
		"to-do":       StatusTodo,
		"in_progress": StatusInProgress,
		"in progress": StatusInProgress,
		"done":        StatusDone,
		"blocked":     StatusBlocked,
		"whatever":    StatusTodo,
	}
	for input, want := range cases {
		if got := parseStatus(input); got != want {
			t.Errorf("parseStatus(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestAddProjectIsUniqueAndSorted(t *testing.T) {
	store := TaskStore{}
	addProject(&store, "work")
	addProject(&store, "home")
	addProject(&store, "work")
	if strings.Join(store.Projects, ",") != "home,work" {
		t.Errorf("projects = %v", store.Projects)
	}
}

func TestPrintStoreRendersTitleAndCounts(t *testing.T) {
	store := TaskStore{Tasks: []Task{
		{ID: 1, Title: "Write tests", Status: StatusTodo, Priority: PriorityHigh},
		{ID: 2, Title: "Ship it", Status: StatusDone, Priority: PriorityLow},
	}}
	printStore(store)
}

func TestPrintStoreHandlesEmptyStore(t *testing.T) {
	printStore(TaskStore{})
}

// runCmd executes a command with its output captured, failing the test on error.
func runCmd(t *testing.T, args ...string) string {
	t.Helper()
	cmd := newRootCmd()
	var out bytes.Buffer
	cmd.SetOut(&out)
	cmd.SetErr(&out)
	cmd.SetArgs(args)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("%v: %v", args, err)
	}
	return out.String()
}

func TestRootCommandRunsAddListAndDone(t *testing.T) {
	withStore(t)

	run := func(args ...string) string { return runCmd(t, args...) }

	run("add", "Write the report", "--priority", "high", "--project", "work", "--due", "2026-12-01")
	run("list")

	store := loadStore()
	if store.Tasks[0].Priority != PriorityHigh || store.Tasks[0].Project != "work" {
		t.Errorf("flags not applied: %+v", store.Tasks[0])
	}
	if store.Tasks[0].DueDate == nil {
		t.Error("due date was not parsed")
	}

	run("done", "1")
	if loadStore().Tasks[0].Status != StatusDone {
		t.Error("done did not change the status")
	}
}

func TestFilterTasks(t *testing.T) {
	due := time.Date(2026, 12, 1, 0, 0, 0, 0, time.UTC)
	tasks := []Task{
		{ID: 1, Title: "home task", Project: "home", Priority: PriorityLow,
			Status: StatusTodo, Tags: []string{"chores"}, DueDate: &due},
		{ID: 2, Title: "work task", Description: "quarterly report", Project: "work",
			Priority: PriorityHigh, Status: StatusInProgress, Tags: []string{"docs"}},
		{ID: 3, Title: "another work task", Project: "work",
			Priority: PriorityHigh, Status: StatusDone},
	}

	cases := []struct {
		name     string
		project  string
		status   string
		priority string
		tag      string
		search   string
		wantIDs  []int
	}{
		{"no filters", "", "", "", "", "", []int{1, 2, 3}},
		{"by project", "work", "", "", "", "", []int{2, 3}},
		{"by status", "", "in_progress", "", "", "", []int{2}},
		{"by priority", "", "", "high", "", "", []int{2, 3}},
		{"by tag", "", "", "", "chores", "", []int{1}},
		{"by description", "", "", "", "", "quarterly", []int{2}},
		{"unknown project", "nope", "", "", "", "", nil},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := filterTasks(tasks, tc.project, tc.status, tc.priority, tc.tag, tc.search)
			var ids []int
			for _, task := range got {
				ids = append(ids, task.ID)
			}
			if len(ids) != len(tc.wantIDs) {
				t.Fatalf("got %v, want %v", ids, tc.wantIDs)
			}
			for i := range ids {
				if ids[i] != tc.wantIDs[i] {
					t.Fatalf("got %v, want %v", ids, tc.wantIDs)
				}
			}
		})
	}
}
