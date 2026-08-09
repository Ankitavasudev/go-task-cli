package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Task struct {
	ID          int         + "" + `json:"`id"` + "" + 
	Title       string      + "" + `json:"`title"` + "" + 
	Description string      + "" + `json:"`description"` + "" + 
	Priority    string      + "" + `json:"`priority"` + "" + 
	Tags        []string    + "" + `json:"`tags"` + "" + 
	DueDate     string      + "" + `json:"`due_date"` + "" + 
	Completed   bool        + "" + `json:"`completed"` + "" + 
	CreatedAt   time.Time   + "" + `json:"`created_at"` + "" + 
	CompletedAt *time.Time  + "" + `json:"`completed_at,omitempty"` + "" + 
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	switch os.Args[1] {
	case "interactive", "i":
		RunInteractive()
	case "export":
		exportCmd := flag.NewFlagSet("export", flag.ExitOnError)
		filename := exportCmd.String("file", "tasks.csv", "Output filename")
		exportCmd.Parse(os.Args[2:])
		tasks := loadTasks()
		if err := ExportToCSV(tasks, *filename); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
	case "import":
		importCmd := flag.NewFlagSet("import", flag.ExitOnError)
		filename := importCmd.String("file", "", "Input CSV filename")
		importCmd.Parse(os.Args[2:])
		if *filename == "" {
			fmt.Println("Error: --file is required")
			os.Exit(1)
		}
		tasks := loadTasks()
		imported, err := ImportFromCSV(*filename)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		tasks = append(tasks, imported...)
		saveTasks(tasks)
		fmt.Printf("Imported %d tasks\n", len(imported))
	default:
		runCLI()
	}
}

func printUsage() {
	fmt.Println("Go Task CLI - Manage your tasks efficiently")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  add         Add a new task")
	fmt.Println("  list        List all tasks")
	fmt.Println("  done        Mark a task as completed")
	fmt.Println("  delete      Delete a task")
	fmt.Println("  search      Search tasks")
	fmt.Println("  stats       Show task statistics")
	fmt.Println("  priority    List tasks by priority")
	fmt.Println("  due         List tasks by due date")
	fmt.Println("  export      Export tasks to CSV")
	fmt.Println("  import      Import tasks from CSV")
	fmt.Println("  interactive Start interactive mode")
	fmt.Println()
	fmt.Println("Flags:")
	fmt.Println("  --json      Output as JSON (list command)")
	fmt.Println("  --csv       Output as CSV (list command)")
}

func runCLI() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	title := addCmd.String("title", "", "Task title")
	desc := addCmd.String("desc", "", "Task description")
	priority := addCmd.String("priority", "3", "Priority (1-5)")
	tags := addCmd.String("tags", "", "Comma-separated tags")
	due := addCmd.String("due", "", "Due date (YYYY-MM-DD)")

	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	jsonOutput := listCmd.Bool("json", false, "Output as JSON")
	csvOutput := listCmd.Bool("csv", false, "Output as CSV")

	doneCmd := flag.NewFlagSet("done", flag.ExitOnError)
	doneID := doneCmd.Int("id", 0, "Task ID to mark as done")

	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteID := deleteCmd.Int("id", 0, "Task ID to delete")

	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	query := searchCmd.String("q", "", "Search query")

	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		if *title == "" {
			fmt.Println("Error: --title is required")
			os.Exit(1)
		}
		tasks := loadTasks()
		newTask := Task{
			ID:          len(tasks) + 1,
			Title:       *title,
			Description: *desc,
			Priority:    *priority,
			Tags:        strings.Split(*tags, ","),
			DueDate:     *due,
			CreatedAt:   time.Now(),
		}
		tasks = append(tasks, newTask)
		saveTasks(tasks)
		fmt.Printf("Task added: %s (ID: %d)\n", newTask.Title, newTask.ID)

	case "list":
		listCmd.Parse(os.Args[2:])
		tasks := loadTasks()
		if *jsonOutput {
			printJSON(tasks)
		} else if *csvOutput {
			printCSV(tasks)
		} else {
			printTasks(tasks)
		}

	case "done":
		doneCmd.Parse(os.Args[2:])
		if *doneID == 0 {
			fmt.Println("Error: --id is required")
			os.Exit(1)
		}
		tasks := loadTasks()
		for i := range tasks {
			if tasks[i].ID == *doneID {
				tasks[i].Completed = true
				now := time.Now()
				tasks[i].CompletedAt = &now
				saveTasks(tasks)
				fmt.Printf("Completed: %s\n", tasks[i].Title)
				return
			}
		}
		fmt.Printf("Task with ID %d not found\n", *doneID)

	case "delete":
		deleteCmd.Parse(os.Args[2:])
		if *deleteID == 0 {
			fmt.Println("Error: --id is required")
			os.Exit(1)
		}
		tasks := loadTasks()
		for i := range tasks {
			if tasks[i].ID == *deleteID {
				title := tasks[i].Title
				tasks = append(tasks[:i], tasks[i+1:]...)
				saveTasks(tasks)
				fmt.Printf("Deleted: %s\n", title)
				return
			}
		}
		fmt.Printf("Task with ID %d not found\n", *deleteID)

	case "search":
		searchCmd.Parse(os.Args[2:])
		tasks := loadTasks()
		queryLower := strings.ToLower(*query)
		found := false
		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Title), queryLower) ||
				strings.Contains(strings.ToLower(t.Description), queryLower) {
				fmt.Printf("  [%d] %s (priority: %s)\n", t.ID, t.Title, t.Priority)
				found = true
			}
		}
		if !found {
			fmt.Println("No matching tasks found.")
		}

	case "stats":
		tasks := loadTasks()
		printStats(tasks)

	case "priority":
		tasks := loadTasks()
		sortByPriority(tasks)
		printTasks(tasks)

	case "due":
		tasks := loadTasks()
		sortByDueDate(tasks)
		printTasks(tasks)

	default:
		fmt.Printf("Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func loadTasks() []Task {
	var tasks []Task
	data, err := os.ReadFile("tasks.json")
	if err != nil {
		return tasks
	}
	json.Unmarshal(data, &tasks)
	return tasks
}

func saveTasks(tasks []Task) {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	os.WriteFile("tasks.json", data, 0644)
}

func printTasks(tasks []Task) {
	if len(tasks) == 0 {
		fmt.Println("No tasks found.")
		return
	}
	fmt.Println("\nID | Title | Priority | Tags | Due | Status")
	fmt.Println("---|-------|----------|------|-----|-------")
	for _, t := range tasks {
		status := "pending"
		if t.Completed {
			status = "done"
		}
		fmt.Printf("%d | %s | %s | %s | %s | %s\n",
			t.ID, t.Title, t.Priority, strings.Join(t.Tags, ","), t.DueDate, status)
	}
}

func printJSON(tasks []Task) {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	fmt.Println(string(data))
}

func printCSV(tasks []Task) {
	fmt.Println("ID,Title,Description,Priority,Tags,DueDate,Completed")
	for _, t := range tasks {
		completed := "false"
		if t.Completed {
			completed = "true"
		}
		fmt.Printf("%d,\"%s\",\"%s\",%s,\"%s\",%s,%s\n",
			t.ID, t.Title, t.Description, t.Priority,
			strings.Join(t.Tags, ";"), t.DueDate, completed)
	}
}

func printStats(tasks []Task) {
	total := len(tasks)
	done := 0
	for _, t := range tasks {
		if t.Completed {
			done++
		}
	}
	pending := total - done

	fmt.Println("\nTask Statistics")
	fmt.Println(strings.Repeat("=", 30))
	fmt.Printf("Total:     %d\n", total)
	fmt.Printf("Completed: %d\n", done)
	fmt.Printf("Pending:   %d\n", pending)
	if total > 0 {
		fmt.Printf("Progress:  %.1f%%\n", float64(done)/float64(total)*100)
	}
}

func sortByPriority(tasks []Task) {
	sort.Slice(tasks, func(i, j int) bool {
		pi, _ := strconv.Atoi(tasks[i].Priority)
		pj, _ := strconv.Atoi(tasks[j].Priority)
		return pi > pj
	})
}

func sortByDueDate(tasks []Task) {
	sort.Slice(tasks, func(i, j int) bool {
		return tasks[i].DueDate < tasks[j].DueDate
	})
}