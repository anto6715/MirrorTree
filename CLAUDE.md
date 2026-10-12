# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

MirrorTree (`mtree`) mirrors a source directory tree into a target directory as mock files and directories. It is meant for scenarios where only the tree structure and file names matter, such as presence checks and rolling or archive rules triggered by file name.

Current state: a prototype. It builds the mirror plan and prints it. Nothing is written to disk yet.

## Commands

```bash
go run ./cmd/mtree [-L max-depth] [-f filter] <source> <dest>
go build -o mtree ./cmd/mtree     # .gitignore does not cover extensionless binaries
go test ./...
go test ./internal/mtreeapp -run 'TestMirrorPath/relative_target' -v   # single test / subtest
go vet ./...
golangci-lint run ./...           # installed locally, no repo config; see errcheck note below
```

## Architecture

`cmd/mtree/main.go` parses the CLI with kong and calls `mtreeapp.Run`. It also decides `Config.Color`: colors are on only when stdout is a terminal and neither `--no-color` nor `NO_COLOR` is set. Errors are logged through `slog` (from `internal/logging`) to stderr. The plan goes to stdout.

`internal/mtreeapp` is a single package split by role. It follows the layout of `internal/safermapp` in the sibling `goat` repo. Everything except `Config` and `Run` is unexported. `Run` is a linear pipeline:

1. `normalizeConfig` (config.go):
   - **Source** is canonicalized (`Abs` + `EvalSymlinks`) and must be a directory.
   - **Target** is deliberately *not* canonicalized, so it may be relative. It must be non-empty and must not exist. The check uses `os.Lstat`, so a dangling symlink counts as existing.
   - **Filter** is validated up front with `filepath.Match(filter, "")`.
2. `find.FindFilesWithDirs` (from goat) returns a map of directory → matching files.
3. `buildMirrorPlan` (plan.go) is pure, with no I/O. It sorts directories and files so the output is deterministic. Each file also stores its `rel` path. Source paths map to target paths only through `mirrorPath` (`filepath.Rel` + `filepath.Join`). Never use string replacement: it corrupts paths like `/data/database` when the source is `/data`.
4. `writeMirrorPlan(w, plan, formatOptions)` (formatter.go) prints a `tree`-style view of the target:
   - **Tree building:** `buildTree` (tree.go) builds the tree from each file's `rel`. It creates intermediate directories that hold no files themselves, because the plan only lists directories that contain files. Its counts exclude the root, matching `tree`.
   - **Order:** children are sorted by name in byte order, with files and directories mixed.
   - **Colors:** the `style` helper adds ANSI codes, or returns the text unchanged when color is off. The formatter never looks at the terminal itself.
   - **Unsafe names:** `displayName` quotes names with control characters or invalid UTF-8, so they can't break the layout or inject terminal escape codes.
   - **Write errors:** output goes through `bufio.Writer`. Errors from individual `Fprint*` calls are ignored, and the first one surfaces from `Flush()`. This is why golangci-lint's errcheck reports them. `TestWriteMirrorPlanReturnsWriterError` covers this behavior.

The planned write step is an `execute.go` with `executePlan(mirrorPlan)`, following goat's saferm (print the plan, then apply). Today the whole walk finishes before anything happens, so a target inside the source is harmless. If creation ever moves into the walk callback, guard against target-inside-source.

## goat dependency

`github.com/anto6715/goat/find` comes from the author's own repo, checked out next to this one at `../goat`. The repo has no tags, so the dependency is pinned to a pseudo-version. Walk behavior is defined in goat, not here:

- Only directories that contain at least one matching file appear. Empty directories are not mirrored. This is a known, accepted limitation for now.
- Symlinks are not followed. A symlink to a directory is returned as a file.
- Permission and not-exist errors during the walk are skipped silently.
- The filter is matched against the base name only.
- `MaxDepth`: `0` means only files directly in the source, and `-1` means unlimited.

Changes to walk behavior belong in goat. To develop both repos together, use `go work init . ../goat` (`go.work` is gitignored). Then push goat and run `go get github.com/anto6715/goat@<version>` here.

## Tests

Tests are in-package and use only the standard library, in the same style as goat:
- table-driven cases with `t.Run`
- `t.TempDir()` with the `makeTestDir` / `writeTestFile` helpers in `config_test.go`
- `testPlan(t, rels...)` in `plan_test.go`, which builds a `/data` → `mirror` plan from relative file paths
- golden-string comparisons for formatter output
- a `failingWriter` to check that write errors propagate
