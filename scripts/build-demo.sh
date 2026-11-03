#!/usr/bin/env bash
# Regenerate the board published on GitHub Pages.
set -euo pipefail
cd "$(dirname "$0")/.."
mkdir -p docs
GO_TASK_CLI_FILE=examples/tasks.json go run . board > docs/index.html
echo "wrote docs/index.html"
