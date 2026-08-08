package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type Priority int

const (
	Low Priority = iota
	Medium
	High
	Urgent
)

func (p Priority) String() string {
	return [...]string{"Low", "Medium", "High", "Urgent"}[p]
}

func (p Priority) Color() string {
	return [...]string{"[green]", "[yellow]", "[orange]", "[red]"}[p]
}

type Task struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Completed   bool      `json:"completed"`
	Priority    Priority  `json:"priority"`
	DueDate     string    `json:"due_date,omitempty"`
	Tags        []string  `json:"tags,omitempty"`
	CreatedAt   string    `json:"created_at"`
	CompletedAt string    `json:"completed_at,omitempty"`
}

type TaskStore struct {
	Tasks  []Task `json:"tasks"`
	NextID int    `json:"next_id"`
}

var storeFile string

func init() {
	home, _ := os.UserHomeDir()
	storeFile = home + "/.tasks.json"
}

func loadStore() TaskStore {
	var store TaskStore
	data, err := os.ReadFile(storeFile)
	if err != nil {
		return TaskStore{NextID: 1}
	}
	json.Unmarshal(data, &store)
	return store
}

func saveStore(store TaskStore) {
	data, _ := json.MarshalIndent(store, "", "  ")
	os.WriteFile(storeFile, data, 0644)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "task",
		Short: "Minimal task manager with persistence, priorities, and due dates",
		Long:  "A production-grade CLI task manager with JSON persistence, priorities, due dates, tags, and filtering.",
	}

	var addCmd = &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			priority, _ := cmd.Flags().GetInt("priority")
			due, _ := cmd.Flags().GetString("due")
			tags, _ := cmd.Flags().GetStringSlice("tags")

			store := loadStore()
			task := Task{
				ID:        store.NextID,
				Title:     strings.Join(args, " "),
				Priority:  Priority(priority),
				DueDate:   due,
				Tags:      tags,
				CreatedAt: time.Now().Format(time.RFC3339),
			}
			store.Tasks = append(store.Tasks, task)
			store.NextID++
			saveStore(store)

			fmt.Printf("Added task #%d: %s [Priority: %s]\n", task.ID, task.Title, task.Priority)
			if due != "" {
				fmt.Printf("  Due: %s\n", due)
			}
			if len(tags) > 0 {
				fmt.Printf("  Tags: %s\n", strings.Join(tags, ", "))
			}
		},
	}
	addCmd.Flags().IntP("priority", "p", 1, "Priority: 0=Low, 1=Medium, 2=High, 3=Urgent")
	addCmd.Flags().StringP("due", "d", "", "Due date (YYYY-MM-DD)")
	addCmd.Flags().StringSliceP("tags", "t", nil, "Tags (comma-separated)")

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List tasks with optional filters",
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			showAll, _ := cmd.Flags().GetBool("all")
			priorityFilter, _ := cmd.Flags().GetInt("priority")
			tagFilter, _ := cmd.Flags().GetString("tag")
			overdueOnly, _ := cmd.Flags().GetBool("overdue")

			var filtered []Task
			for _, t := range store.Tasks {
				if !showAll && t.Completed {
					continue
				}
				if priorityFilter >= 0 && int(t.Priority) != priorityFilter {
					continue
				}
				if tagFilter != "" && !containsTag(t.Tags, tagFilter) {
					continue
				}
				if overdueOnly && !isOverdue(t.DueDate) {
					continue
				}
				filtered = append(filtered, t)
			}

			if len(filtered) == 0 {
				fmt.Println("No tasks found.")
				return
			}

			sort.Slice(filtered, func(i, j int) bool {
				if filtered[i].Priority != filtered[j].Priority {
					return filtered[i].Priority > filtered[j].Priority
				}
				return filtered[i].ID < filtered[j].ID
			})

			fmt.Printf("\n%-4s %-6s %-40s %-12s %-20s %s\n", "ID", "Pri", "Title", "Due", "Tags", "Status")
			fmt.Println(strings.Repeat("-", 100))
			for _, t := range filtered {
				status := "[ ]"
				if t.Completed {
					status = "[x]"
				}
				due := t.DueDate
				if due == "" {
					due = "-"
				}
				if isOverdue(t.DueDate) && !t.Completed {
					due = "[RED]" + due + "[/RED]"
				}
				tags := strings.Join(t.Tags, ",")
				if tags == "" {
					tags = "-"
				}
				fmt.Printf("%-4d %s %-40s %-12s %-20s %s\n",
					t.ID, t.Priority.Color()+t.Priority.String()+"[/]", t.Title, due, tags, status)
			}
			fmt.Printf("\nTotal: %d tasks\n", len(filtered))
		},
	}
	listCmd.Flags().BoolP("all", "a", false, "Show completed tasks too")
	listCmd.Flags().IntP("priority", "p", -1, "Filter by priority")
	listCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	listCmd.Flags().Bool("overdue", false, "Show only overdue tasks")

	var doneCmd = &cobra.Command{
		Use:   "done [id]",
		Short: "Mark task as completed",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := strconv.Atoi(args[0])
			store := loadStore()
			for i, t := range store.Tasks {
				if t.ID == id {
					store.Tasks[i].Completed = true
					store.Tasks[i].CompletedAt = time.Now().Format(time.RFC3339)
					saveStore(store)
					fmt.Printf("Task #%d completed: %s\n", id, t.Title)
					return
				}
			}
			fmt.Printf("Task #%d not found\n", id)
		},
	}

	var removeCmd = &cobra.Command{
		Use:   "remove [id]",
		Short: "Remove a task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := strconv.Atoi(args[0])
			store := loadStore()
			for i, t := range store.Tasks {
				if t.ID == id {
					store.Tasks = append(store.Tasks[:i], store.Tasks[i+1:]...)
					saveStore(store)
					fmt.Printf("Task #%d removed: %s\n", id, t.Title)
					return
				}
			}
			fmt.Printf("Task #%d not found\n", id)
		},
	}

	var editCmd = &cobra.Command{
		Use:   "edit [id]",
		Short: "Edit task title",
		Args:  cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := strconv.Atoi(args[0])
			newTitle := strings.Join(args[1:], " ")
			store := loadStore()
			for i, t := range store.Tasks {
				if t.ID == id {
					store.Tasks[i].Title = newTitle
					saveStore(store)
					fmt.Printf("Task #%d updated: %s\n", id, newTitle)
					return
				}
			}
			fmt.Printf("Task #%d not found\n", id)
		},
	}

	var statsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Show detailed statistics",
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			total := len(store.Tasks)
			completed := 0
			byPriority := make(map[string]int)
			byTag := make(map[string]int)
			overdue := 0

			for _, t := range store.Tasks {
				if t.Completed {
					completed++
				}
				byPriority[t.Priority.String()]++
				for _, tag := range t.Tags {
					byTag[tag]++
				}
				if !t.Completed && isOverdue(t.DueDate) {
					overdue++
				}
			}

			fmt.Println("\n=== Task Statistics ===")
			fmt.Printf("Total:     %d\n", total)
			fmt.Printf("Completed: %d\n", completed)
			fmt.Printf("Pending:   %d\n", total-completed)
			fmt.Printf("Overdue:   %d\n", overdue)
			if total > 0 {
				fmt.Printf("Progress:  %.0f%%\n", float64(completed)/float64(total)*100)
			}

			fmt.Println("\nBy Priority:")
			for p, c := range byPriority {
				fmt.Printf("  %s: %d\n", p, c)
			}

			if len(byTag) > 0 {
				fmt.Println("\nBy Tag:")
				for tag, c := range byTag {
					fmt.Printf("  %s: %d\n", tag, c)
				}
			}
		},
	}

	var searchCmd = &cobra.Command{
		Use:   "search [query]",
		Short: "Search tasks by title",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			query := strings.ToLower(strings.Join(args, " "))
			store := loadStore()
			found := false
			for _, t := range store.Tasks {
				if strings.Contains(strings.ToLower(t.Title), query) {
					status := "[ ]"
					if t.Completed { status = "[x]" }
					fmt.Printf("#%d %s %s [Priority: %s]\n", t.ID, status, t.Title, t.Priority)
					found = true
				}
			}
			if !found {
				fmt.Println("No matching tasks found.")
			}
		},
	}

	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export tasks as JSON",
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			data, _ := json.MarshalIndent(store.Tasks, "", "  ")
			fmt.Println(string(data))
		},
	}

	rootCmd.AddCommand(addCmd, listCmd, doneCmd, removeCmd, editCmd, statsCmd, searchCmd, exportCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func containsTag(tags []string, tag string) bool {
	for _, t := range tags {
		if strings.EqualFold(t, tag) {
			return true
		}
	}
	return false
}

func isOverdue(due string) bool {
	if due == "" {
		return false
	}
	t, err := time.Parse("2006-01-02", due)
	if err != nil {
		return false
	}
	return time.Now().After(t)
}
