# Go Task CLI

> Feature-rich command-line task manager built in Go with CSV import/export, scheduling, caching, and webhook notifications.

[![CI](https://github.com/Ankitavasudev/go-task-cli/actions/workflows/ci.yml/badge.svg)](https://github.com/Ankitavasudev/go-task-cli/actions)
[![Go 1.21+](https://img.shields.io/badge/go-1.21+-00ADD8.svg)](https://golang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

## Features

- **Task Management** - Create, complete, delete, sort tasks by priority/due date
- **CSV Import/Export** - Bulk import/export tasks via CSV files
- **Interactive Mode** - TUI for visual task management
- **Scheduler** - Overdue detection, upcoming tasks, daily summaries
- **Webhooks** - HTTP notifications on task events
- **Caching** - File-based caching for offline support
- **Statistics** - Productivity scores, completion rates

## Quick Start

```bash
# Install
git clone https://github.com/Ankitavasudev/go-task-cli.git
cd go-task-cli
go build -o task-cli .

# Add tasks
./task-cli add "Buy groceries" -p 3 -t "shopping;personal"
./task-cli add "Deploy v2.0" -p 5 -d 2026-08-15 -t "work;urgent"

# List tasks
./task-cli list
./task-cli list --sort priority
./task-cli list --filter "shopping"

# Complete tasks
./task-cli complete 1

# CSV operations
./task-cli export tasks.csv
./task-cli import tasks.csv

# Scheduler
./task-cli schedule
./task-cli overdue
./task-cli upcoming

# Stats
./task-cli stats
./task-cli productivity
```

## Interactive Mode

```bash
./task-cli interactive
```

Navigate with arrow keys, press Enter to complete, 'd' to delete, 'a' to add.

## CSV Format

```csv
ID,Title,Description,Priority,Tags,DueDate,Completed,CreatedAt,CompletedAt
1,Buy groceries,Weekly shopping,3,shopping;personal,2026-08-15,false,2026-08-09T00:00:00Z,
2,Deploy v2.0,Production release,5,work;urgent,2026-08-20,false,2026-08-09T00:00:00Z,
```

## Webhooks

```bash
# Start webhook server
./task-cli webhook --url http://localhost:8080/tasks

# Task events sent as POST JSON:
# { "event": "created", "task": {...} }
# { "event": "completed", "task": {...} }
# { "event": "deleted", "task": {...} }
```

## Architecture

```
go-task-cli/
├── main.go           # CLI entry point, commands
├── scheduler.go      # Overdue/upcoming detection, stats
├── csv_export.go     # CSV import/export
├── interactive.go    # TUI mode
├── webhook.go        # HTTP webhook notifications
├── cache.go          # File-based caching
└── *_test.go         # Unit tests
```

## Tech Stack

- **Go 1.21+** - Core language
- **encoding/csv** - CSV parsing
- **net/http** - Webhook server
- **testing** - Unit tests

## Contributing

1. Fork the repo
2. Create a feature branch (`git checkout -b feature/amazing`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing`)
5. Open a Pull Request

## License

MIT License - see [LICENSE](LICENSE) for details.