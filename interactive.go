package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

var reader = bufio.NewReader(os.Stdin)

func prompt(question string) string {
	fmt.Print(question)
	input, _ := reader.ReadString('\n')
	return strings.TrimSpace(input)
}

func promptInt(question string) int {
	for {
		input := prompt(question)
		val, err := strconv.Atoi(input)
		if err != nil {
			fmt.Println("Please enter a valid number.")
			continue
		}
		return val
	}
}

func RunInteractive() {
	fmt.Println("=== Go Task CLI - Interactive Mode ===")
	fmt.Println("Commands: add, list, done, delete, search, export, import, stats, quit")
	fmt.Println()

	var tasks []Task

	for {
		cmd := prompt("\n> ")
		switch strings.ToLower(cmd) {
		case "add":
			title := prompt("Title: ")
			desc := prompt("Description: ")
			priority := prompt("Priority (1-5): ")
			tags := prompt("Tags (comma-separated): ")
			due := prompt("Due date (YYYY-MM-DD, optional): ")

			task := Task{
				Title:       title,
				Description: desc,
				Priority:    priority,
				Tags:        strings.Split(tags, ","),
				DueDate:     due,
			}
			tasks = append(tasks, task)
			fmt.Printf("Added: %s\n", title)

		case "list":
			if len(tasks) == 0 {
				fmt.Println("No tasks.")
				continue
			}
			fmt.Println("\nID | Title | Priority | Tags | Due | Status")
			fmt.Println("---|-------|----------|------|-----|-------")
			for i, t := range tasks {
				status := "pending"
				if t.Completed {
					status = "done"
				}
				fmt.Printf("%d | %s | %s | %s | %s | %s\n",
					i+1, t.Title, t.Priority, strings.Join(t.Tags, ","), t.DueDate, status)
			}

		case "done":
			id := promptInt("Task ID: ")
			if id > 0 && id <= len(tasks) {
				tasks[id-1].Completed = true
				fmt.Printf("Completed: %s\n", tasks[id-1].Title)
			} else {
				fmt.Println("Invalid task ID.")
			}

		case "delete":
			id := promptInt("Task ID: ")
			if id > 0 && id <= len(tasks) {
				title := tasks[id-1].Title
				tasks = append(tasks[:id-1], tasks[id:]...)
				fmt.Printf("Deleted: %s\n", title)
			} else {
				fmt.Println("Invalid task ID.")
			}

		case "search":
			query := prompt("Search query: ")
			found := false
			for i, t := range tasks {
				if strings.Contains(strings.ToLower(t.Title), strings.ToLower(query)) ||
					strings.Contains(strings.ToLower(t.Description), strings.ToLower(query)) {
					fmt.Printf("  [%d] %s (priority: %s)\n", i+1, t.Title, t.Priority)
					found = true
				}
			}
			if !found {
				fmt.Println("No matching tasks.")
			}

		case "export":
			filename := prompt("Filename (e.g., tasks.csv): ")
			if err := ExportToCSV(tasks, filename); err != nil {
				fmt.Printf("Error: %v\n", err)
			}

		case "import":
			filename := prompt("Filename: ")
			imported, err := ImportFromCSV(filename)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			tasks = append(tasks, imported...)
			fmt.Printf("Imported %d tasks.\n", len(imported))

		case "stats":
			total := len(tasks)
			done := 0
			for _, t := range tasks {
				if t.Completed {
					done++
				}
			}
			fmt.Printf("\nTotal: %d | Done: %d | Pending: %d\n", total, done, total-done)
			if total > 0 {
				fmt.Printf("Completion: %.1f%%\n", float64(done)/float64(total)*100)
			}

		case "quit", "exit", "q":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command. Try: add, list, done, delete, search, export, import, stats, quit")
		}
	}
}