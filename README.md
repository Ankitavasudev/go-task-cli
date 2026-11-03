# go-task-cli

A project-aware todo list that lives in one JSON file. Cobra on the command
line, no database, no server, no account. It can also print the store as a
read-only HTML kanban board you can publish or paste into a review.

```bash
go-task-cli add "Write the report" --project work --priority high --due 2026-12-01 --tag docs --tag urgent
go-task-cli list
go-task-cli done 1
go-task-cli board > board.html
```

[See a generated board](https://ankitavasudev.github.io/go-task-cli/), built
from [`examples/tasks.json`](examples/tasks.json).

The store is `~/.go-task-cli.json`; set `GO_TASK_CLI_FILE` to point somewhere
else (the tests rely on that to stay out of your home directory).

## Install

```bash
go install github.com/Ankitavasudev/go-task-cli@v1.1.0
```

Go 1.21 or newer. `go install` names the binary after the module, so the command
is `go-task-cli`. From a clone, `go build -o go-task-cli .` produces the same
binary.

`go-task-cli version` prints the version it was built from: the module version
when installed with `go install`, or the value compiled into the source when
built from a checkout.

## Commands

| Command | What it does |
|---|---|
| `add <title>` | `-p/--project`, `-P/--priority`, `-t/--tag` (repeatable), `--due YYYY-MM-DD`, `--desc` |
| `list` (`ls`) | filter with `-p`, `-s/--status`, `-P`, `-t`, `-S/--search` |
| `done <id>` | marks the task done and records `completed_at` |
| `progress <id>` | sets the status to `in_progress` |
| `delete <id>` (`rm`) | removes the task |
| `edit <id>` | change title/description/project/priority/status/tags |
| `projects` (`proj`) | every project name in the store, with a task count |
| `search <query>` | title + description substring search |
| `stats` | counts by status, priority and project |
| `export` / `import <file>` | the whole store as JSON |
| `csv-export <file>` / `csv-import <file>` | the task list as CSV |
| `board` | a self-contained HTML kanban board on stdout |
| `version` | the version this binary was built from |
| `interactive` (`i`) | prompts for a new task, one field at a time |

Priorities are `low`, `medium`, `high`, `critical`; statuses are `todo`,
`in_progress`, `done`, `blocked`. `parseStatus` also accepts `in-progress`,
`wip`, `completed`, and underscores or spaces.

## The HTML board

`board` reads the same store as every other command and writes one file to
stdout:

```bash
go-task-cli board > board.html
GO_TASK_CLI_FILE=examples/tasks.json go-task-cli board > demo.html
```

The page has inline CSS, no JavaScript and no external requests, so it opens
from disk, from a `file://` URL, or from any static host. Titles, descriptions,
projects and tags are escaped by `html/template`, so a task called
`<script>alert(1)</script>` is rendered as text rather than executed.

Cards are grouped into Todo, In progress, Blocked and Done. Inside a column they
are ordered by priority, then by due date (earliest first, no due date last),
then by ID, so the same store always produces the same order. A task with a past
due date is marked overdue unless it is done.

The board is read-only by design: a static file cannot write to your store, and
adding an API for that would need a server. Use the CLI to change a task, then
regenerate the board.

## CSV format

`csv-export` writes this header, and `csv-import` requires exactly this order:

```csv
id,title,description,project,priority,status,tags,created_at,due_date,completed_at
1,Buy groceries,Weekly shopping,home,medium,todo,shopping;personal,2026-08-09T10:00:00Z,2026-08-15T00:00:00Z,
```

Tags are separated by `;` (a comma is already the field separator). Dates are
RFC 3339; `YYYY-MM-DD` is accepted for due dates on import. IDs in the file are
ignored — imported rows get fresh IDs so they never collide with the local
store. Rows with an empty title are rejected instead of silently dropped.

## Layout

```
main.go        cobra commands, TaskStore load/save, filtering, text rendering
board.go       HTML kanban board, column ordering, version reporting
csv.go         CSV encode/decode + the csv-import / csv-export commands
main_test.go   store round-trip, parsers, filterTasks, command-level tests
board_test.go  grouping, ordering, overdue rules, escaping, offline checks
csv_test.go    quoting, renumbering, bad headers, empty titles
examples/      a 12-task store used by the README, CI and the published board
scripts/       build-demo.sh, which regenerates the published board
docs/          the board served by GitHub Pages
```

## Development

```bash
go test ./...      # no network, no home-directory writes
go vet ./...
gofmt -l .         # should print nothing
./scripts/build-demo.sh   # regenerate docs/index.html
```

## Licence

MIT — see [LICENSE](LICENSE).
