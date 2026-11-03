package main

import (
	"fmt"
	"html/template"
	"runtime/debug"
	"sort"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

// version is used when the binary is built from a checkout. When it is
// installed from a module version, versionString reports that instead, so the
// number in the board matches what was downloaded.
var version = "v1.1.0"

func versionString() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return version
}

// boardColumns is the order the columns appear in, left to right.
var boardColumns = []struct {
	Status Status
	Label  string
	Key    string
}{
	{StatusTodo, "Todo", "todo"},
	{StatusInProgress, "In progress", "progress"},
	{StatusBlocked, "Blocked", "blocked"},
	{StatusDone, "Done", "done"},
}

// priorityRank orders cards inside a column, most urgent first.
var priorityRank = map[Priority]int{
	PriorityCritical: 0,
	PriorityHigh:     1,
	PriorityMedium:   2,
	PriorityLow:      3,
}

type boardCard struct {
	ID          int
	Title       string
	Description string
	Project     string
	Priority    string
	Tags        []string
	Due         string
	Overdue     bool
	Completed   string
}

type boardColumn struct {
	Label string
	Key   string
	Count int
	Cards []boardCard
}

type boardData struct {
	Generated string
	Version   string
	Summary   string
	Columns   []boardColumn
}

func plural(n int, singular, plural string) string {
	if n == 1 {
		return singular
	}
	return plural
}

// sortForBoard orders a column by priority, then by due date (earliest first,
// no due date last), then by ID. Two runs over the same store therefore produce
// the same board, which matters when the file is committed or diffed.
func sortForBoard(tasks []Task) {
	sort.SliceStable(tasks, func(i, j int) bool {
		a, b := tasks[i], tasks[j]
		if priorityRank[a.Priority] != priorityRank[b.Priority] {
			return priorityRank[a.Priority] < priorityRank[b.Priority]
		}
		switch {
		case a.DueDate == nil && b.DueDate != nil:
			return false
		case a.DueDate != nil && b.DueDate == nil:
			return true
		case a.DueDate != nil && b.DueDate != nil && !a.DueDate.Equal(*b.DueDate):
			return a.DueDate.Before(*b.DueDate)
		}
		return a.ID < b.ID
	})
}

func toCard(t Task, now time.Time) boardCard {
	card := boardCard{
		ID:          t.ID,
		Title:       t.Title,
		Description: t.Description,
		Project:     t.Project,
		Priority:    string(t.Priority),
		Tags:        t.Tags,
	}
	if t.DueDate != nil {
		card.Due = t.DueDate.Format("Jan 02, 2006")
		card.Overdue = t.Status != StatusDone && t.DueDate.Before(now)
	}
	if t.CompletedAt != nil {
		card.Completed = t.CompletedAt.Format("Jan 02, 2006")
	}
	return card
}

// buildBoard groups tasks into columns. Unknown statuses fall into Todo, which
// is what parseStatus does too, so an edited store cannot make a task vanish.
func buildBoard(tasks []Task) []boardColumn {
	now := time.Now()
	known := make(map[Status]int, len(boardColumns))
	grouped := make([][]Task, len(boardColumns))
	for i, column := range boardColumns {
		known[column.Status] = i
	}

	for _, task := range tasks {
		index, ok := known[task.Status]
		if !ok {
			index = 0
		}
		grouped[index] = append(grouped[index], task)
	}

	columns := make([]boardColumn, len(boardColumns))
	for i, column := range boardColumns {
		sortForBoard(grouped[i])
		columns[i] = boardColumn{Label: column.Label, Key: column.Key}
		for _, task := range grouped[i] {
			columns[i].Cards = append(columns[i].Cards, toCard(task, now))
		}
		columns[i].Count = len(columns[i].Cards)
	}
	return columns
}

// boardTemplate is one self-contained page: inline styles, no JavaScript and no
// external requests, so it opens from disk and can be published as a static
// file. Titles, descriptions, projects and tags are escaped by html/template.
var boardTemplate = template.Must(template.New("board").Parse(`<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>go-task-cli board</title>
<style>
:root{--bg:#0b1020;--panel:#141a2e;--card:#1b2340;--text:#e7ecf7;--muted:#8fa0c0;--line:#28324f;--low:#3fb950;--medium:#d29922;--high:#f0883e;--critical:#f85149;--accent:#58a6ff}
*{box-sizing:border-box}
body{margin:0;background:var(--bg);color:var(--text);font:15px/1.5 ui-sans-serif,system-ui,-apple-system,"Segoe UI",Roboto,sans-serif}
main{max-width:1280px;margin:0 auto;padding:36px 20px 64px}
h1{font-size:24px;margin:0 0 4px}
.sub{color:var(--muted);font-size:13px;margin-bottom:26px}
.board{display:grid;grid-template-columns:repeat(4,minmax(0,1fr));gap:14px;align-items:start}
@media (max-width:900px){.board{grid-template-columns:repeat(2,minmax(0,1fr))}}
@media (max-width:560px){.board{grid-template-columns:1fr}}
.col{background:var(--panel);border:1px solid var(--line);border-radius:12px;padding:12px}
.col h2{font-size:13px;text-transform:uppercase;letter-spacing:.06em;color:var(--muted);margin:2px 0 12px;display:flex;justify-content:space-between}
.col h2 .count{color:var(--text);background:var(--card);border:1px solid var(--line);border-radius:999px;padding:0 8px;font-size:12px}
.col.done h2 .count{color:var(--low)}
.card{background:var(--card);border:1px solid var(--line);border-left:3px solid var(--muted);border-radius:9px;padding:11px 12px;margin-bottom:10px}
.card:last-child{margin-bottom:0}
.card.low{border-left-color:var(--low)}
.card.medium{border-left-color:var(--medium)}
.card.high{border-left-color:var(--high)}
.card.critical{border-left-color:var(--critical)}
.card .top{display:flex;gap:8px;align-items:baseline;justify-content:space-between}
.card .id{color:var(--muted);font:12px ui-monospace,SFMono-Regular,Consolas,monospace}
.card .title{font-weight:600;font-size:14px;margin:2px 0 4px;word-break:break-word}
.card .desc{color:var(--muted);font-size:13px;margin:0 0 8px;word-break:break-word}
.meta{display:flex;flex-wrap:wrap;gap:6px;font-size:12px;color:var(--muted)}
.pill{border:1px solid var(--line);border-radius:999px;padding:0 7px;font-size:11px}
.pill.project{color:var(--accent);border-color:color-mix(in srgb,var(--accent) 45%,var(--line))}
.pill.tag{color:var(--muted)}
.pill.due{color:var(--medium)}
.pill.overdue{color:var(--critical);border-color:color-mix(in srgb,var(--critical) 45%,var(--line));font-weight:600}
.pill.completed{color:var(--low)}
.empty{color:var(--muted);font-size:13px;padding:6px 2px}
footer{color:var(--muted);font-size:12px;margin-top:26px}
</style>
</head>
<body>
<main>
<h1>Task board</h1>
<div class="sub">Generated {{.Generated}} &middot; {{.Summary}} &middot; {{.Version}}</div>
<div class="board">
{{range .Columns}}<section class="col {{.Key}}">
<h2>{{.Label}} <span class="count">{{.Count}}</span></h2>
{{if .Cards}}{{range .Cards}}<article class="card {{.Priority}}">
<div class="top"><span class="id">#{{.ID}}</span></div>
<div class="title">{{.Title}}</div>
{{if .Description}}<p class="desc">{{.Description}}</p>{{end}}
<div class="meta">
<span class="pill {{.Priority}}">{{.Priority}}</span>
{{if .Project}}<span class="pill project">{{.Project}}</span>{{end}}
{{if .Due}}{{if .Overdue}}<span class="pill overdue">overdue {{.Due}}</span>{{else}}<span class="pill due">due {{.Due}}</span>{{end}}{{end}}
{{if .Completed}}<span class="pill completed">done {{.Completed}}</span>{{end}}
{{range .Tags}}<span class="pill tag">#{{.}}</span>{{end}}
</div>
</article>{{end}}{{else}}<p class="empty">Nothing here.</p>{{end}}
</section>{{end}}
</div>
<footer>Generated by go-task-cli {{.Version}}. The board is read-only; run the CLI to change a task.</footer>
</main>
</body>
</html>
`))

// renderBoard renders the tasks as a self-contained kanban page.
func renderBoard(tasks []Task) (string, error) {
	projects := make(map[string]bool)
	for _, task := range tasks {
		if task.Project != "" {
			projects[task.Project] = true
		}
	}

	var b strings.Builder
	err := boardTemplate.Execute(&b, boardData{
		Generated: time.Now().Format("2006-01-02 15:04 MST"),
		Version:   versionString(),
		Summary: fmt.Sprintf("%d %s across %d %s",
			len(tasks), plural(len(tasks), "task", "tasks"),
			len(projects), plural(len(projects), "project", "projects")),
		Columns: buildBoard(tasks),
	})
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(b.String(), "\n"), nil
}

var boardCmd = &cobra.Command{
	Use:   "board",
	Short: "Print a self-contained HTML kanban board",
	Long: "Print a read-only HTML kanban board for the current store.\n\n" +
		"The page has inline CSS, no JavaScript and no external requests, so it\n" +
		"opens from disk and can be published as a static file.",
	Run: func(cmd *cobra.Command, args []string) {
		store := loadStore()
		page, err := renderBoard(store.Tasks)
		if err != nil {
			fmt.Fprintf(cmd.ErrOrStderr(), "Error: %s\n", err)
			return
		}
		fmt.Fprintln(cmd.OutOrStdout(), page)
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Fprintln(cmd.OutOrStdout(), versionString())
	},
}
