package main

import (
	"os"
	"testing"
)

func TestLoadStoreEmpty(t *testing.T) {
	os.Remove(storeFile)
	store := loadStore()
	if store.NextID != 1 {
		t.Errorf("Expected NextID=1, got %d", store.NextID)
	}
	if len(store.Tasks) != 0 {
		t.Errorf("Expected 0 tasks, got %d", len(store.Tasks))
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		p    Priority
		want string
	}{
		{Low, "Low"},
		{Medium, "Medium"},
		{High, "High"},
		{Urgent, "Urgent"},
	}
	for _, tt := range tests {
		if got := tt.p.String(); got != tt.want {
			t.Errorf("Priority(%d).String() = %s, want %s", tt.p, got, tt.want)
		}
	}
}

func TestContainsTag(t *testing.T) {
	tags := []string{"devops", "urgent", "backend"}
	if !containsTag(tags, "urgent") {
		t.Error("Expected to find 'urgent' tag")
	}
	if containsTag(tags, "frontend") {
		t.Error("Should not find 'frontend' tag")
	}
	if !containsTag(tags, "DevOps") {
		t.Error("Case-insensitive search should work")
	}
}

func TestIsOverdue(t *testing.T) {
	if isOverdue("") {
		t.Error("Empty date should not be overdue")
	}
	if isOverdue("2099-01-01") {
		t.Error("Future date should not be overdue")
	}
	if !isOverdue("2020-01-01") {
		t.Error("Past date should be overdue")
	}
	if isOverdue("invalid-date") {
		t.Error("Invalid date should not be overdue")
	}
}

func TestStorePersistence(t *testing.T) {
	os.Remove(storeFile)
	store := loadStore()
	task := Task{
		ID:        1,
		Title:     "Test task",
		Priority:  Medium,
		CreatedAt: "2026-08-08T10:00:00Z",
	}
	store.Tasks = append(store.Tasks, task)
	store.NextID = 2
	saveStore(store)

	loaded := loadStore()
	if len(loaded.Tasks) != 1 {
		t.Errorf("Expected 1 task after reload, got %d", len(loaded.Tasks))
	}
	if loaded.Tasks[0].Title != "Test task" {
		t.Errorf("Expected title 'Test task', got '%s'", loaded.Tasks[0].Title)
	}
	if loaded.NextID != 2 {
		t.Errorf("Expected NextID=2, got %d", loaded.NextID)
	}
	os.Remove(storeFile)
}