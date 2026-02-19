# AGENTS.md

Agent instructions for this repository.

---

## Repository Overview

This is a monorepo containing two independent tools with a shared purpose: removing comments from source code.

```
nvim-remove-comments/
├── plugin/          # Neovim auto-load entry point
├── lua/             # Neovim plugin (Lua)
│   └── nvim-remove-comments/
│       ├── init.lua
│       ├── core.lua
│       └── config.lua
└── cli/             # Go CLI tool (planned/in-progress)
    ├── main.go
    ├── go.mod
    └── internal/
```

---

## Lua Plugin (Neovim)

### Build & Test Commands

There is no build step for the Lua plugin. Neovim loads it directly.

Manual testing: open Neovim with a supported file and run `:lua require("nvim-remove-comments").remove_comments()` or press `<leader>rc`.

Lint:

```bash
luacheck lua/ plugin/
```

Format:

```bash
stylua lua/ plugin/
```

There are no automated tests for the Lua plugin at this time.

---

## Go CLI (`cli/`)

### Build & Test Commands

```bash
cd cli

go build ./...

go test ./...

go test ./internal/parser/...

go test -run TestRemoveComments ./internal/remover/...

go test -v -run TestWalker ./internal/walker/...

go vet ./...

golangci-lint run ./...
```

Run the CLI in dry-run mode (default):

```bash
go run . .
```

Run with write mode:

```bash
go run . --write .
```

Run a single test by name:

```bash
go test -run <TestFunctionName> ./<package>/...
```

---

## Code Style

### General

- No comments unless the code is genuinely non-obvious and cannot be made clearer by renaming. Absolutely no doc comments, JSDoc, GoDoc blocks, or explanatory prose in code.
- No TODO, FIXME, HACK, or NOTE comments in committed code.
- Prefer clarity via naming over comments.

### Lua (Plugin)

**Formatting:**

- Tabs for indentation (not spaces) — match the existing codebase.
- Use `stylua` for formatting. Config follows the existing style in the repo.
- Max line length: 120 characters.

**Naming:**

- Modules exported as `local M = {}` / `return M` pattern.
- Functions: `snake_case`.
- Local variables: `snake_case`.
- Constants: `SCREAMING_SNAKE_CASE` only when genuinely constant and module-level.

**Imports:**

- All `require()` calls at the top of the file, assigned to locals.
- Never inline `require()` inside functions unless lazy-loading is necessary.

```lua
local ts = vim.treesitter
local parsers = require("nvim-treesitter.parsers")
local config = require("nvim-remove-comments.config")
```

**Error handling:**

- Use early returns for guard clauses. No deeply nested conditionals.
- Silent returns (bare `return`) are acceptable when a missing parser means there is nothing to do.

```lua
if not parsers.has_parser(lang) then
    return
end
```

**Tables used as sets:**

- Use `table[key] = true` for set membership. Do not use arrays when you need O(1) lookup.

**Iteration:**

- Prefer `ipairs` for ordered arrays, `pairs` for hash tables.
- Do not use numeric `for` loops over tables unless index arithmetic is needed.

---

### Go (CLI)

**Formatting:**

- `gofmt` / `goimports` — enforced. No exceptions.
- Use tabs for indentation.
- Max line length: 100 characters (soft limit; prefer readability).

**Naming:**

- Packages: short, lowercase, no underscores (`walker`, `parser`, `remover`).
- Exported types/functions: `PascalCase`.
- Unexported: `camelCase`.
- Acronyms follow Go convention: `URL`, `ID`, `HTTP` (not `Url`, `Id`, `Http`).
- Error variables: `err` for immediate use, `ErrXxx` for sentinel errors.

**Imports:**

- Group with a blank line between: stdlib → external → internal.
- Use `goimports` to manage automatically.

```go
import (
    "fmt"
    "os"

    "github.com/spf13/cobra"

    "github.com/KashifKhn/nvim-remove-comments/cli/internal/parser"
)
```

**Types:**

- Define explicit types for domain concepts (e.g., `CommentRange`, `FileEntry`). Avoid bare `map[string]interface{}`.
- Use `struct` for anything with more than two related fields.
- Prefer value receivers unless mutation or interface satisfaction requires pointer receivers.

**Error handling:**

- Always handle errors. Never assign to `_` unless the error is provably irrelevant.
- Wrap errors with context: `fmt.Errorf("parse %s: %w", path, err)`.
- Do not use `panic` in library/internal code. `panic` only in `main` for unrecoverable startup failures.
- Early return on error. Avoid `else` after an error return.

```go
result, err := doThing()
if err != nil {
    return fmt.Errorf("do thing: %w", err)
}
```

**Concurrency:**

- Use a worker pool pattern for parallel file processing.
- Protect shared state with `sync.Mutex` or channels — never raw global mutation.
- Always pass context (`context.Context`) as the first argument to functions that do I/O.

**Testing:**

- Test files live alongside source: `parser_test.go` next to `parser.go`.
- Use `testing.T` and the standard library. No external test frameworks unless already a dependency.
- Table-driven tests for any function with multiple input/output cases.
- Test function names: `TestFunctionName_scenario` (e.g., `TestRemoveComments_InlineOnly`).

```go
func TestRemoveComments_FullLine(t *testing.T) {
    tests := []struct {
        name  string
        input string
        want  string
    }{
        {"single line comment", "// foo\nfoo()\n", "foo()\n"},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            got := removeComments(tt.input)
            if got != tt.want {
                t.Errorf("got %q, want %q", got, tt.want)
            }
        })
    }
}
```

---

## Language Config (Shared Logic)

The language-to-Tree-sitter-query mapping is the conceptual source of truth for both the plugin and the CLI.

- Lua plugin: `lua/nvim-remove-comments/config.lua`
- Go CLI: `cli/internal/languages/languages.go`

When adding a new language, update **both** files consistently.

The query format is a Tree-sitter S-expression capturing `@comment`:

```
(comment) @comment
(line_comment) @comment
(block_comment) @comment
```

Use the most specific node types available for each language. Fallback default: `(comment) @comment`.

---

## Git Conventions

Commit message format: `type: short description` (lowercase, imperative).

Types: `feat`, `fix`, `refactor`, `test`, `doc`, `chore`.

Examples:

```
feat: add rust block comment support
fix: preserve inline content after comment splice
refactor: extract walker into separate package
```

One logical change per commit. Do not mix refactors with feature additions.

---

## What Not to Do

- Do not add doc comments or explanatory comments to code.
- Do not use `else` after a `return` or `continue`.
- Do not create files unless they are required for the task.
- Do not modify `plugin/` or `lua/` while working on the CLI, and vice versa.
- Do not commit `.env`, secrets, or editor-specific config files.
