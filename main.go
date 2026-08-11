package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type Priority string

const (
	PriorityLow      Priority = "low"
	PriorityMedium   Priority = "medium"
	PriorityHigh     Priority = "high"
	PriorityCritical Priority = "critical"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in_progress"
	StatusDone       Status = "done"
	StatusBlocked    Status = "blocked"
)

type Task struct {
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description,omitempty"`
	Project     string     `json:"project,omitempty"`
	Priority    Priority   `json:"priority"`
	Status      Status     `json:"status"`
	Tags        []string   `json:"tags,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DueDate     *time.Time `json:"due_date,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
}

type TaskStore struct {
	Tasks    []Task   `json:"tasks"`
	NextID   int      `json:"next_id"`
	Projects []string `json:"projects,omitempty"`
}

func getStorePath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ".go-task-cli.json"
	}
	return filepath.Join(home, ".go-task-cli.json")
}

func loadStore() TaskStore {
	var store TaskStore
	data, err := os.ReadFile(getStorePath())
	if err != nil {
		return TaskStore{NextID: 1}
	}
	json.Unmarshal(data, &store)
	if store.NextID == 0 {
		store.NextID = 1
	}
	return store
}

func saveStore(store TaskStore) error {
	data, err := json.MarshalIndent(store, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(getStorePath(), data, 0644)
}

func addProject(store *TaskStore, project string) {
	for _, p := range store.Projects {
		if p == project {
			return
		}
	}
	store.Projects = append(store.Projects, project)
	sort.Strings(store.Projects)
}

func findNextID(store TaskStore) int {
	maxID := 0
	for _, t := range store.Tasks {
		if t.ID > maxID {
			maxID = t.ID
		}
	}
	return maxID + 1
}

func parsePriority(s string) Priority {
	switch strings.ToLower(s) {
	case "low":
		return PriorityLow
	case "medium", "med":
		return PriorityMedium
	case "high":
		return PriorityHigh
	case "critical", "crit":
		return PriorityCritical
	default:
		return PriorityMedium
	}
}

func parseStatus(s string) Status {
	switch strings.ToLower(s) {
	case "todo", "to-do":
		return StatusTodo
	case "in-progress", "progress", "wip":
		return StatusInProgress
	case "done", "completed":
		return StatusDone
	case "blocked":
		return StatusBlocked
	default:
		return StatusTodo
	}
}

func priorityColor(p Priority) string {
	switch p {
	case PriorityLow:
		return "\033[32m"
	case PriorityMedium:
		return "\033[33m"
	case PriorityHigh:
		return "\033[31m"
	case PriorityCritical:
		return "\033[1;31m"
	default:
		return "\033[0m"
	}
}

func statusIcon(s Status) string {
	switch s {
	case StatusTodo:
		return "\033[33m[ ]\033[0m"
	case StatusInProgress:
		return "\033[34m[>]\033[0m"
	case StatusDone:
		return "\033[32m[x]\033[0m"
	case StatusBlocked:
		return "\033[31m[!]\033[0m"
	default:
		return "[ ]"
	}
}

func printTask(t Task) {
	fmt.Printf("  %s \033[36m#%d\033[0m %s%s\033[0m",
		statusIcon(t.ID), t.ID, priorityColor(t.Priority), t.Title)
	if t.Project != "" {
		fmt.Printf(" \033[35m[%s]\033[0m", t.Project)
	}
	for _, tag := range t.Tags {
		fmt.Printf(" \033[36m#%s\033[0m", tag)
	}
	if t.DueDate != nil {
		days := time.Until(*t.DueDate).Hours() / 24
		if days < 0 {
			fmt.Printf(" \033[31mOVERDUE\033[0m")
		} else if days < 2 {
			fmt.Printf(" \033[33mDUE SOON\033[0m")
		}
	}
	fmt.Println()
	if t.Description != "" {
		fmt.Printf("    \033[90m%s\033[0m\n", t.Description)
	}
	fmt.Printf("    Created: %s", t.CreatedAt.Format("Jan 02, 2006 15:04"))
	if t.CompletedAt != nil {
		fmt.Printf(" | Completed: %s", t.CompletedAt.Format("Jan 02, 2006 15:04"))
	}
	fmt.Println()
}

func printStore(store TaskStore) {
	if len(store.Tasks) == 0 {
		fmt.Println("\033[33mNo tasks found.\033[0m")
		return
	}
	sorted := make([]Task, len(store.Tasks))
	copy(sorted, store.Tasks)
	sort.Slice(sorted, func(i, j int) bool {
		pOrder := map[Priority]int{PriorityCritical: 0, PriorityHigh: 1, PriorityMedium: 2, PriorityLow: 3}
		sOrder := map[Status]int{StatusInProgress: 0, StatusBlocked: 1, StatusTodo: 2, StatusDone: 3}
		if pOrder[sorted[i].Priority] != pOrder[sorted[j].Priority] {
			return pOrder[sorted[i].Priority] < pOrder[sorted[j].Priority]
		}
		return sOrder[sorted[i].Status] < sOrder[sorted[j].Status]
	})
	for _, t := range sorted {
		printTask(t)
	}
	todo, inProgress, done := 0, 0, 0
	for _, t := range store.Tasks {
		switch t.Status {
		case StatusTodo:
			todo++
		case StatusInProgress:
			inProgress++
		case StatusDone:
			done++
		}
	}
	fmt.Printf("\n\033[90m--- %d total | %d todo | %d in progress | %d done ---\033[0m\n",
		len(store.Tasks), todo, inProgress, done)
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "go-task-cli",
		Short: "A fast, project-aware task manager for developers",
	}

	var addCmd = &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			project, _ := cmd.Flags().GetString("project")
			priority, _ := cmd.Flags().GetString("priority")
			tags, _ := cmd.Flags().GetStringSlice("tag")
			due, _ := cmd.Flags().GetString("due")
			desc, _ := cmd.Flags().GetString("desc")
			task := Task{
				ID:          findNextID(store),
				Title:       args[0],
				Description: desc,
				Project:     project,
				Priority:    parsePriority(priority),
				Status:      StatusTodo,
				Tags:        tags,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if due != "" {
				if t, err := time.Parse("2006-01-02", due); err == nil {
					task.DueDate = &t
				}
			}
			if project != "" {
				addProject(&store, project)
			}
			store.Tasks = append(store.Tasks, task)
			store.NextID = task.ID + 1
			if err := saveStore(store); err != nil {
				fmt.Printf("\033[31mError: %s\033[0m\n", err)
				return
			}
			fmt.Printf("\033[32mTask #%d created\033[0m: %s\n", task.ID, task.Title)
		},
	}
	addCmd.Flags().StringP("project", "p", "", "Project name")
	addCmd.Flags().StringP("priority", "P", "medium", "Priority (low, medium, high, critical)")
	addCmd.Flags().StringSliceP("tag", "t", nil, "Tags (repeatable)")
	addCmd.Flags().String("due", "", "Due date (YYYY-MM-DD)")
	addCmd.Flags().String("desc", "", "Description")

	var listCmd = &cobra.Command{
		Use:     "list",
		Short:   "List all tasks",
		Aliases: []string{"ls"},
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			project, _ := cmd.Flags().GetString("project")
			status, _ := cmd.Flags().GetString("status")
			priority, _ := cmd.Flags().GetString("priority")
			tag, _ := cmd.Flags().GetString("tag")
			search, _ := cmd.Flags().GetString("search")
			filtered := store.Tasks
			if project != "" {
				var temp []Task
				for _, t := range filtered {
					if t.Project == project {
						temp = append(temp, t)
					}
				}
				filtered = temp
			}
			if status != "" {
				s := parseStatus(status)
				var temp []Task
				for _, t := range filtered {
					if t.Status == s {
						temp = append(temp, t)
					}
				}
				filtered = temp
			}
			if priority != "" {
				p := parsePriority(priority)
				var temp []Task
				for _, t := range filtered {
					if t.Priority == p {
						temp = append(temp, t)
					}
				}
				filtered = temp
			}
			if tag != "" {
				var temp []Task
				for _, t := range filtered {
					for _, tg := range t.Tags {
						if tg == tag {
							temp = append(temp, t)
							break
						}
					}
				}
				filtered = temp
			}
			if search != "" {
				lower := strings.ToLower(search)
				var temp []Task
				for _, t := range filtered {
					if strings.Contains(strings.ToLower(t.Title), lower) ||
						strings.Contains(strings.ToLower(t.Description), lower) {
						temp = append(temp, t)
					}
				}
				filtered = temp
			}
			store.Tasks = filtered
			printStore(store)
		},
	}
	listCmd.Flags().StringP("project", "p", "", "Filter by project")
	listCmd.Flags().StringP("status", "s", "", "Filter by status")
	listCmd.Flags().StringP("priority", "P", "", "Filter by priority")
	listCmd.Flags().StringP("tag", "t", "", "Filter by tag")
	listCmd.Flags().StringP("search", "S", "", "Search in title/description")

	var doneCmd = &cobra.Command{
		Use:   "done [id]",
		Short: "Mark task as done",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			var id int
			fmt.Sscanf(args[0], "%d", &id)
			for i, t := range store.Tasks {
				if t.ID == id {
					now := time.Now()
					store.Tasks[i].Status = StatusDone
					store.Tasks[i].CompletedAt = &now
					store.Tasks[i].UpdatedAt = now
					saveStore(store)
					fmt.Printf("\033[32mTask #%d marked as done\033[0m\n", id)
					return
				}
			}
			fmt.Printf("\033[31mTask #%d not found\033[0m\n", id)
		},
	}

	var progressCmd = &cobra.Command{
		Use:   "progress [id]",
		Short: "Mark task as in-progress",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			var id int
			fmt.Sscanf(args[0], "%d", &id)
			for i, t := range store.Tasks {
				if t.ID == id {
					store.Tasks[i].Status = StatusInProgress
					store.Tasks[i].UpdatedAt = time.Now()
					saveStore(store)
					fmt.Printf("\033[34mTask #%d marked as in-progress\033[0m\n", id)
					return
				}
			}
			fmt.Printf("\033[31mTask #%d not found\033[0m\n", id)
		},
	}

	var deleteCmd = &cobra.Command{
		Use:     "delete [id]",
		Short:   "Delete a task",
		Aliases: []string{"rm"},
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			var id int
			fmt.Sscanf(args[0], "%d", &id)
			for i, t := range store.Tasks {
				if t.ID == id {
					store.Tasks = append(store.Tasks[:i], store.Tasks[i+1:]...)
					saveStore(store)
					fmt.Printf("\033[32mTask #%d deleted\033[0m\n", id)
					return
				}
			}
			fmt.Printf("\033[31mTask #%d not found\033[0m\n", id)
		},
	}

	var editCmd = &cobra.Command{
		Use:   "edit [id]",
		Short: "Edit a task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			var id int
			fmt.Sscanf(args[0], "%d", &id)
			for i, t := range store.Tasks {
				if t.ID == id {
					title, _ := cmd.Flags().GetString("title")
					desc, _ := cmd.Flags().GetString("desc")
					project, _ := cmd.Flags().GetString("project")
					priority, _ := cmd.Flags().GetString("priority")
					status, _ := cmd.Flags().GetString("status")
					tags, _ := cmd.Flags().GetStringSlice("tag")
					if title != "" {
						store.Tasks[i].Title = title
					}
					if desc != "" {
						store.Tasks[i].Description = desc
					}
					if project != "" {
						store.Tasks[i].Project = project
						addProject(&store, project)
					}
					if priority != "" {
						store.Tasks[i].Priority = parsePriority(priority)
					}
					if status != "" {
						store.Tasks[i].Status = parseStatus(status)
					}
					if len(tags) > 0 {
						store.Tasks[i].Tags = tags
					}
					store.Tasks[i].UpdatedAt = time.Now()
					saveStore(store)
					fmt.Printf("\033[32mTask #%d updated\033[0m\n", id)
					return
				}
			}
			fmt.Printf("\033[31mTask #%d not found\033[0m\n", id)
		},
	}
	editCmd.Flags().String("title", "", "New title")
	editCmd.Flags().String("desc", "", "New description")
	editCmd.Flags().StringP("project", "p", "", "New project")
	editCmd.Flags().StringP("priority", "P", "", "New priority")
	editCmd.Flags().StringP("status", "s", "", "New status")
	editCmd.Flags().StringSliceP("tag", "t", nil, "New tags")

	var projectsCmd = &cobra.Command{
		Use:     "projects",
		Short:   "List all projects",
		Aliases: []string{"proj"},
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			if len(store.Projects) == 0 {
				fmt.Println("\033[33mNo projects found.\033[0m")
				return
			}
			for _, p := range store.Projects {
				count := 0
				for _, t := range store.Tasks {
					if t.Project == p {
						count++
					}
				}
				fmt.Printf("  \033[35m%s\033[0m (%d tasks)\n", p, count)
			}
		},
	}

	var searchCmd = &cobra.Command{
		Use:   "search [query]",
		Short: "Search tasks by title or description",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			query := strings.ToLower(args[0])
			var results []Task
			for _, t := range store.Tasks {
				if strings.Contains(strings.ToLower(t.Title), query) ||
					strings.Contains(strings.ToLower(t.Description), query) {
					results = append(results, t)
				}
			}
			if len(results) == 0 {
				fmt.Println("\033[33mNo matching tasks found.\033[0m")
				return
			}
			for _, t := range results {
				printTask(t)
			}
		},
	}

	var exportCmd = &cobra.Command{
		Use:   "export",
		Short: "Export tasks as JSON",
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			data, _ := json.MarshalIndent(store, "", "  ")
			fmt.Println(string(data))
		},
	}

	var importCmd = &cobra.Command{
		Use:   "import [file]",
		Short: "Import tasks from JSON file",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			data, err := os.ReadFile(args[0])
			if err != nil {
				fmt.Printf("\033[31mError reading file: %s\033[0m\n", err)
				return
			}
			var imported TaskStore
			if err := json.Unmarshal(data, &imported); err != nil {
				fmt.Printf("\033[31mError parsing JSON: %s\033[0m\n", err)
				return
			}
			store := loadStore()
			for _, t := range imported.Tasks {
				t.ID = findNextID(store)
				store.Tasks = append(store.Tasks, t)
				store.NextID = t.ID + 1
				if t.Project != "" {
					addProject(&store, t.Project)
				}
			}
			saveStore(store)
			fmt.Printf("\033[32mImported %d tasks\033[0m\n", len(imported.Tasks))
		},
	}

	var statsCmd = &cobra.Command{
		Use:   "stats",
		Short: "Show task statistics",
		Run: func(cmd *cobra.Command, args []string) {
			store := loadStore()
			total := len(store.Tasks)
			todo, inProgress, done, blocked := 0, 0, 0, 0
			pLow, pMed, pHigh, pCrit := 0, 0, 0, 0
			projects := make(map[string]int)
			for _, t := range store.Tasks {
				switch t.Status {
				case StatusTodo:
					todo++
				case StatusInProgress:
					inProgress++
				case StatusDone:
					done++
				case StatusBlocked:
					blocked++
				}
				switch t.Priority {
				case PriorityLow:
					pLow++
				case PriorityMedium:
					pMed++
				case PriorityHigh:
					pHigh++
				case PriorityCritical:
					pCrit++
				}
				if t.Project != "" {
					projects[t.Project]++
				}
			}
			fmt.Printf("\n  \033[1mTask Statistics\033[0m\n")
			fmt.Printf("  Total:     %d\n", total)
			fmt.Printf("  Todo:      \033[33m%d\033[0m\n", todo)
			fmt.Printf("  In Progress: \033[34m%d\033[0m\n", inProgress)
			fmt.Printf("  Done:      \033[32m%d\033[0m\n", done)
			fmt.Printf("  Blocked:   \033[31m%d\033[0m\n", blocked)
			fmt.Printf("\n  Priority:\n")
			fmt.Printf("    Low:      %d\n", pLow)
			fmt.Printf("    Medium:   %d\n", pMed)
			fmt.Printf("    High:     %d\n", pHigh)
			fmt.Printf("    Critical: %d\n", pCrit)
			if len(projects) > 0 {
				fmt.Printf("\n  Projects:\n")
				for _, p := range store.Projects {
					if c, ok := projects[p]; ok {
						fmt.Printf("    \033[35m%s\033[0m: %d\n", p, c)
					}
				}
			}
			fmt.Println()
		},
	}

	var interactiveCmd = &cobra.Command{
		Use:     "interactive",
		Short:   "Interactive mode - add tasks with prompts",
		Aliases: []string{"i"},
		Run: func(cmd *cobra.Command, args []string) {
			reader := bufio.NewReader(os.Stdin)
			store := loadStore()
			fmt.Print("Task title: ")
			title, _ := reader.ReadString('\n')
			title = strings.TrimSpace(title)
			if title == "" {
				fmt.Println("\033[31mTitle cannot be empty\033[0m")
				return
			}
			fmt.Print("Description (optional): ")
			desc, _ := reader.ReadString('\n')
			desc = strings.TrimSpace(desc)
			fmt.Print("Project (optional): ")
			project, _ := reader.ReadString('\n')
			project = strings.TrimSpace(project)
			fmt.Print("Priority (low/medium/high/critical): ")
			priority, _ := reader.ReadString('\n')
			priority = strings.TrimSpace(priority)
			fmt.Print("Tags (comma-separated): ")
			tagsStr, _ := reader.ReadString('\n')
			var tags []string
			for _, t := range strings.Split(strings.TrimSpace(tagsStr), ",") {
				t = strings.TrimSpace(t)
				if t != "" {
					tags = append(tags, t)
				}
			}
			fmt.Print("Due date (YYYY-MM-DD, optional): ")
			dueStr, _ := reader.ReadString('\n')
			dueStr = strings.TrimSpace(dueStr)
			task := Task{
				ID:          findNextID(store),
				Title:       title,
				Description: desc,
				Project:     project,
				Priority:    parsePriority(priority),
				Status:      StatusTodo,
				Tags:        tags,
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			}
			if dueStr != "" {
				if t, err := time.Parse("2006-01-02", dueStr); err == nil {
					task.DueDate = &t
				}
			}
			if project != "" {
				addProject(&store, project)
			}
			store.Tasks = append(store.Tasks, task)
			store.NextID = task.ID + 1
			saveStore(store)
			fmt.Printf("\033[32mTask #%d created\033[0m: %s\n", task.ID, task.Title)
		},
	}

	rootCmd.AddCommand(addCmd, listCmd, doneCmd, progressCmd, deleteCmd, editCmd,
		projectsCmd, searchCmd, exportCmd, importCmd, statsCmd, interactiveCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
