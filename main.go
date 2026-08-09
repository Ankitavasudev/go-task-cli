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
	ID          int        `json:"id"`
	Title       string     `json:"title"`
	Description string     `json:"description"`
	Priority    string     `json:"priority"`
	Tags        []string   `json:"tags"`
	DueDate     string     `json:"due_date"`
	Completed   bool       `json:"completed"`
	CreatedAt   time.Time  `json:"created_at"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
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
	fmt.Println("Go Task CLI")
	fmt.Println("Commands: add, list, done, delete, search, stats, export, import, interactive")
}

func runCLI() {
	addCmd := flag.NewFlagSet("add", flag.ExitOnError)
	title := addCmd.String("title", "", "Task title")
	desc := addCmd.String("desc", "", "Task description")
	priority := addCmd.String("priority", "3", "Priority 1-5")
	tags := addCmd.String("tags", "", "Comma-separated tags")
	due := addCmd.String("due", "", "Due date YYYY-MM-DD")
	listCmd := flag.NewFlagSet("list", flag.ExitOnError)
	jsonOutput := listCmd.Bool("json", false, "Output as JSON")
	csvOutput := listCmd.Bool("csv", false, "Output as CSV")
	doneCmd := flag.NewFlagSet("done", flag.ExitOnError)
	doneID := doneCmd.Int("id", 0, "Task ID")
	deleteCmd := flag.NewFlagSet("delete", flag.ExitOnError)
	deleteID := deleteCmd.Int("id", 0, "Task ID")
	searchCmd := flag.NewFlagSet("search", flag.ExitOnError)
	query := searchCmd.String("q", "", "Search query")
	switch os.Args[1] {
	case "add":
		addCmd.Parse(os.Args[2:])
		if *title == "" {
			fmt.Println("Error: --title required")
			os.Exit(1)
		}
		tasks := loadTasks()
		newTask := Task{ID: len(tasks) + 1, Title: *title, Description: *desc, Priority: *priority, Tags: strings.Split(*tags, ","), DueDate: *due, CreatedAt: time.Now()}
		tasks = append(tasks, newTask)
		saveTasks(tasks)
		fmt.Printf("Added: %s\n", newTask.Title)
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
		tasks := loadTasks()
		for i := range tasks {
			if tasks[i].ID == *doneID {
				tasks[i].Completed = true
				now := time.Now()
				tasks[i].CompletedAt = &now
				saveTasks(tasks)
				fmt.Printf("Done: %s\n", tasks[i].Title)
				return
			}
		}
	case "delete":
		deleteCmd.Parse(os.Args[2:])
		tasks := loadTasks()
		for i := range tasks {
			if tasks[i].ID == *deleteID {
				tasks = append(tasks[:i], tasks[i+1:]...)
				saveTasks(tasks)
				fmt.Println("Deleted")
				return
			}
		}
	case "search":
		searchCmd.Parse(os.Args[2:])
		tasks := loadTasks()
		for _, t := range tasks {
			if strings.Contains(strings.ToLower(t.Title), strings.ToLower(*query)) {
				fmt.Printf("[%d] %s\n", t.ID, t.Title)
			}
		}
	case "stats":
		printStats(loadTasks())
	case "priority":
		tasks := loadTasks()
		sortByPriority(tasks)
		printTasks(tasks)
	case "due":
		tasks := loadTasks()
		sortByDueDate(tasks)
		printTasks(tasks)
	default:
		fmt.Printf("Unknown: %s\n", os.Args[1])
		printUsage()
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
	for _, t := range tasks {
		status := "pending"
		if t.Completed {
			status = "done"
		}
		fmt.Printf("[%d] %s | P%s | %s\n", t.ID, t.Title, t.Priority, status)
	}
}

func printJSON(tasks []Task) {
	data, _ := json.MarshalIndent(tasks, "", "  ")
	fmt.Println(string(data))
}

func printCSV(tasks []Task) {
	fmt.Println("ID,Title,Priority,Completed")
	for _, t := range tasks {
		fmt.Printf("%d,%s,%s,%v\n", t.ID, t.Title, t.Priority, t.Completed)
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
	fmt.Printf("Total: %d | Done: %d | Pending: %d\n", total, done, total-done)
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
