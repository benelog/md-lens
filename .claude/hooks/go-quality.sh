#!/usr/bin/env bash
# PostToolUse hook: Go quality gate for edited files.
# Reads the hook JSON on stdin; for *.go files inside the project it
#   1. auto-formats with gofmt + goimports (silent fixes)
#   2. runs go vet + golangci-lint on the file's package
# Exits 2 with findings on stderr so Claude sees them and fixes them.
set -u

input=$(cat)
file=$(printf '%s' "$input" | jq -r '.tool_input.file_path // empty' 2>/dev/null)

[ -n "$file" ] || exit 0
case "$file" in *.go) ;; *) exit 0 ;; esac
[ -f "$file" ] || exit 0

proj="${CLAUDE_PROJECT_DIR:-$PWD}"
case "$file" in
  "$proj"/archive/*) exit 0 ;; # preserved Java-era sources — never touch
  "$proj"/*) ;;
  *) exit 0 ;; # outside the project (e.g. /tmp scratch)
esac

cd "$proj" || exit 0

# 1) Auto-format: fix style quietly instead of complaining about it.
gofmt -w "$file" 2>/dev/null
command -v goimports >/dev/null 2>&1 && goimports -w "$file" 2>/dev/null

# 2) Static analysis on the file's package only (keeps the hook fast).
rel="${file#"$proj"/}"
pkg="./$(dirname "$rel")"

issues=""
if ! vet_out=$(go vet "$pkg" 2>&1); then
  issues="${issues}--- go vet ${pkg} ---\n${vet_out}\n"
fi
if command -v golangci-lint >/dev/null 2>&1; then
  if ! lint_out=$(golangci-lint run "$pkg" 2>&1); then
    issues="${issues}--- golangci-lint ${pkg} ---\n${lint_out}\n"
  fi
fi

if [ -n "$issues" ]; then
  printf '%b' "$issues" >&2
  exit 2
fi
exit 0
