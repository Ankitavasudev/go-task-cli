# go-task-cli

A minimal task manager CLI built in Go with [cobra](https://github.com/spf13/cobra).

## Features

- Add tasks
- List tasks
- Mark tasks as done
- Remove tasks

## Installation

`ash
go install github.com/Ankitavasudev/go-task-cli@latest
`

## Usage

`ash
# Add a task
task add "Buy groceries"

# List tasks
task list

# Mark task as done
task done 1

# Remove a task
task remove 1
`

## Building from Source

`ash
git clone https://github.com/Ankitavasudev/go-task-cli.git
cd go-task-cli
go build -o task .
`

## License

MIT License