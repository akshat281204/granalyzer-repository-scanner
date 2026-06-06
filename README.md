# granalyzer

A language/framework-agnostic static analysis command-line utility for scanning and analyzing local repositories. Built purely using Go with zero LLM or external service dependencies. 

`granalyzer` provides a comprehensive view of your codebase's structure, statistics, dependencies, and HTTP endpoints via either a terminal user interface (TUI) or static terminal printouts.

---

## Features

- **Concurrent Repository Walking:** Fast directory scanning with native `.gitignore` and `.granalyzerignore` support.
- **Language Classification:** Automatically classifies files by extension and peeks inside them to detect frameworks.
- **Interactive TUI:** An elegant dashboard built with `charmbracelet/bubbletea` featuring:
  - **File Tree & Preview:** Navigate your codebase structure and preview files with syntax highlighting (`alecthomas/chroma/v2`).
  - **Repository Stats:** View language breakdown, lines of code, sizes, and top largest files.
  - **Dependency Graph:** Visualize structural relationships and internal import trees.
  - **Endpoint Explorer:** Search and look up detected HTTP API routes in your code.
- **Dedicated CLI Subcommands:** Export stats, endpoints, and dependency graphs as raw text, JSON, or Graphviz DOT formats for integration with downstream CI/CD pipelines.

---

## Tech Stack

- **TUI Framework:** [Bubble Tea](https://github.com/charmbracelet/bubbletea), [Lip Gloss](https://github.com/charmbracelet/lipgloss), [Bubbles](https://github.com/charmbracelet/bubbles)
- **CLI Framework:** [Cobra](https://github.com/spf13/cobra)
- **Syntax Highlighting:** [Chroma](https://github.com/alecthomas/chroma)
- **Parsers:** `go/parser` (native AST parser for Go) & Regex engines for generic framework route detections (Express, Flask, Spring, Axum, Actix, Django, FastAPI).

---

## Installation

Ensure you have [Go](https://go.dev/doc/install) installed (v1.22+ recommended).

Clone the repository and run:
```bash
make build
```
This compiles the executable into `./bin/granalyzer`.

To install the binary globally into your `$GOPATH/bin`:
```bash
make install
```

---

## Usage

### 1. Interactive TUI (default or `scan` command)
```bash
# Launches TUI on current directory
granalyzer scan

# Launches TUI on target folder path
granalyzer scan /path/to/project
```

#### TUI Keyboard Shortcuts
- `1` - `4`: Switch between tabs (`Files`, `Stats`, `Deps`, `Endpoints`)
- `Tab`: Switch focus between panels (e.g. File tree and Preview viewport)
- `/`: Activate search/filtering
- `r`: Re-run repository scanner
- `j`/`k` (or `↑`/`↓`): Scroll/navigate lists
- `Enter`: Expand directories in the tree
- `q` or `Ctrl+C`: Quit the application

### 2. Standalone CLI Subcommands
Each subcommand executes static analysis and outputs result directly to standard output.

#### Print Stats
```bash
granalyzer stats [path] [--json]
```

#### Print Dependencies
```bash
granalyzer deps [path] [--json] [--dot]
```
- Use `--dot` to produce output compatible with Graphviz:
  ```bash
  granalyzer deps --dot > graph.dot
  dot -Tpng graph.dot -o graph.png
  ```

#### Print Detected HTTP Endpoints
```bash
granalyzer endpoints [path] [--json]
```

---

## Limitations

- **Best-Effort Endpoint Detection:** Endpoint mapping relies on static AST parsing for Go and regex patterns for other languages. Dynamically built endpoints (e.g. routes constructed inside loops, custom registration helper wrappers, or initialized dynamically at runtime) will not be captured.
- **Dependency Graph Cap:** To prevent terminal performance degradation on large codebases, the dependency tree display inside the TUI defaults to truncation at 200 nodes.
- **File Preview Limit:** Files larger than 500 lines are truncated in the preview panel to keep memory consumption low.

---

## Development

Use the included `Makefile` to speed up common tasks:

- **Build binary:** `make build`
- **Install globally:** `make install`
- **Run tests:** `make test`
- **Lint code:** `make lint`
