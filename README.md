# go-task-cli

A project-aware todo list that lives in one JSON file. Built with cobra, no
database, no server, no account.

```bash
go build -o task-cli .
./task-cli add "Write the report" --project work --priority high --due 2026-12-01 --tag docs --tag urgent
./task-cli list
./task-cli done 1
```

The store is `~/.go-task-cli.json`; set `GO_TASK_CLI_FILE` to point somewhere
else (the tests rely on that to stay out of your home directory).

## Commands

| Command | What it does |
|---|---|
| `add <title>` | `-p/--project`, `-P/--priority`, `-t/--tag` (repeatable), `--due YYYY-MM-DD`, `--desc` |
| `list` (`ls`) | filter with `-p`, `-s/--status`, `-P`, `-t`, `-S/--search` |
| `done <id>` | marks the task done and records `completed_at` |
| `progress <id>` | sets the status to `in_progress` |
| `delete <id>` | removes the task |
| `edit <id>` | change title/description/project/priority/due date |
| `projects` | every project name in the store |
| `search <query>` | title + description substring search |
| `stats` | counts by status, priority and project |
| `export` / `import <file>` | the whole store as JSON |
| `csv-export <file>` / `csv-import <file>` | the task list as CSV |
| `interactive` (`i`) | prompts for a new task, one field at a time |

Priorities are `low`, `medium`, `high`, `critical`; statuses are `todo`,
`in_progress`, `done`, `blocked`. `parseStatus` also accepts `in-progress`,
`wip`, `completed`, and underscores or spaces.

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
main.go        cobra commands, TaskStore load/save, filtering, rendering
csv.go         CSV encode/decode + the csv-import / csv-export commands
main_test.go   store round-trip, parsers, filterTasks, command-level tests
csv_test.go    quoting, renumbering, bad headers, empty titles
```

## Development

```bash
go test ./...      # no network, no home-directory writes
go vet ./...
gofmt -l .         # should print nothing
```

## Licence

MIT — see [LICENSE](LICENSE).
