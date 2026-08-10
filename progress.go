package main

import "fmt"

func PrintProgressBar(completed, total int) {
	if total == 0 {
		fmt.Println("No tasks")
		return
	}
	pct := float64(completed) / float64(total) * 100
	filled := int(pct / 5)
	bar := ""
	for i := 0; i < 20; i++ {
		if i < filled {
			bar += "#"
		} else {
			bar += "-"
		}
	}
	fmt.Printf("[%s] %d/%d (%.0f%%)\n", bar, completed, total, pct)
}