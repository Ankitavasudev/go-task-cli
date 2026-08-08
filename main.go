package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
)

var tasks []Task

type Task struct {
	ID        int    json:"id"
	Title     string json:"title"
	Completed bool   json:"completed"
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "task",
		Short: "Minimal task manager CLI",
		Long:  "A simple and minimal task manager built in Go",
	}

	var addCmd = &cobra.Command{
		Use:   "add [title]",
		Short: "Add a new task",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			title := strings.Join(args, " ")
			task := Task{ID: len(tasks) + 1, Title: title, Completed: false}
			tasks = append(tasks, task)
			fmt.Printf("Added task #%d: %s\n", task.ID, task.Title)
		},
	}

	var listCmd = &cobra.Command{
		Use:   "list",
		Short: "List all tasks",
		Run: func(cmd *cobra.Command, args []string) {
			if len(tasks) == 0 {
				fmt.Println("No tasks yet. Use 'task add' to create one.")
				return
			}
			for _, t := range tasks {
				status := "[ ]"
				if t.Completed {
					status = "[x]"
				}
				fmt.Printf("%s #%d: %s\n", status, t.ID, t.Title)
			}
		},
	}

	var doneCmd = &cobra.Command{
		Use:   "done [id]",
		Short: "Mark a task as completed",
		Args:  cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			id, _ := strconv.Atoi(args[0])
			for i, t := range tasks {
				if t.ID == id {
					tasks[i].Completed = true
					fmt.Printf("Task #%d completed!\n", id)
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
			for i, t := range tasks {
				if t.ID == id {
					tasks = append(tasks[:i], tasks[i+1:]...)
					fmt.Printf("Task #%d removed\n", id)
					return
				}
			}
			fmt.Printf("Task #%d not found\n", id)
		},
	}

	rootCmd.AddCommand(addCmd, listCmd, doneCmd, removeCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}