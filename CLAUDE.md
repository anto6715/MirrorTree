# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project

MirrorTree (`mtree`) mirrors a source directory tree into a target directory as mock files and directories. It is meant for scenarios where only the tree structure and file names matter, such as presence checks and rolling or archive rules triggered by file name.

By default it is a dry run that only prints the mirror plan. `--apply` also creates the mirror: directories plus empty (0-byte) files.

## Commands

```bash
go run ./cmd/mtree [-L max-depth] [-f filter] [--apply] <source> <dest>
go build -o mtree ./cmd/mtree     # only /main is gitignored, not other binary names
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
   - **Target** is deliberately *not* canonicalized, so it may be relative. It is only `filepath.Clean`ed: with a trailing slash, `filepath.Dir(target)` would return the target itself, and creating the parents would create the target. It must be non-empty and must not exist. The check uses `os.Lstat`, so a dangling symlink counts as existing.
   - **Filter** is validated up front with `filepath.Match(filter, "")`.
2. `find.FindFilesWithDirs` (from goat) returns a map of directory → matching files.
3. `buildMirrorPlan` (plan.go) is pure, with no I/O. It sorts directories and files so the output is deterministic. Each file also stores its `rel` path. Source paths map to target paths only through `mirrorPath` (`filepath.Rel` + `filepath.Join`). Never use string replacement: it corrupts paths like `/data/database` when the source is `/data`.
4. `writeMirrorPlan(w, plan, formatOptions)` (formatter.go) prints a `tree`-style view of the target:
   - **Tree building:** `buildTree` (tree.go) builds the tree from each file's `rel`. It creates intermediate directories that hold no files themselves, because the plan only lists directories that contain files. Its counts exclude the root, matching `tree`.
   - **Order:** children are sorted by name in byte order, with files and directories mixed.
   - **Colors:** the `style` helper adds ANSI codes, or returns the text unchanged when color is off. The formatter never looks at the terminal itself.
   - **Unsafe names:** `displayName` quotes names with control characters or invalid UTF-8, so they can't break the layout or inject terminal escape codes.
   - **Write errors:** output goes through `bufio.Writer`. Errors from individual `Fprint*` calls are ignored, and the first one surfaces from `Flush()`. This is why golangci-lint's errcheck reports them. `TestWriteMirrorPlanReturnsWriterError` covers this behavior.
   - **Footer:** without `--apply`, the plan ends with a dry-run notice.
5. `executePlan` (execute.go) runs only with `--apply` and when at least one file matched. It follows goat's saferm: print the plan, then apply it.
   - **Parents:** missing parents of the target are created with `MkdirAll`.
   - **Target:** created with `os.Mkdir`, so the run fails if the target appeared after validation.
   - **Contents:** created by walking `buildTree` in the same sorted order the printer uses. Directories are created with `os.Mkdir` (0755) and files with `O_CREATE|O_EXCL` (0644, empty), so nothing existing is ever overwritten.
   - **Errors:** it stops at the first error and leaves the partial mirror in place. This is a user decision: never add automatic cleanup. It returns how many directories and files it created, and `Run` puts those counts in the error.
   - **Success:** `writeMirrorCreated` prints the closing line.

The whole walk finishes before anything is created, so a target inside the source is harmless. If creation ever moves into the walk callback, guard against target-inside-source.

## goat dependency

`github.com/anto6715/goat/find` comes from the author's own repo, checked out next to this one at `../goat`. The repo has no tags, so the dependency is pinned to a pseudo-version. Walk behavior is defined in goat, not here:

- Only directories that contain at least one matching file appear. Empty directories are not mirrored. This is a known, accepted limitation for now.
- Symlinks are not followed. Any symlink, including one to a directory, is returned as a file, so it is mirrored as an empty regular file.
- Permission and not-exist errors during the walk are skipped silently.
- The filter is matched against the base name only.
- `MaxDepth`: `0` means only files directly in the source, and `-1` means unlimited.

Changes to walk behavior belong in goat. To develop both repos together, use `go work init . ../goat` (`go.work` is gitignored). Then push goat and run `go get github.com/anto6715/goat@<version>` here.

## Tests

Tests are in-package and use only the standard library, in the same style as goat:
- table-driven cases with `t.Run`
- `t.TempDir()` with the `makeTestDir` / `writeTestFile` helpers in `config_test.go`
- `testPlan(t, rels...)` in `plan_test.go`, which builds a `/data` → `mirror` plan from relative file paths. `testPlanAt(t, target, rels...)` does the same with a custom target; execute tests use it with a target under `t.TempDir()`.
- creation failures are forced with a NUL byte in a file name (`"b\x00.nc"`) instead of file permissions, so the tests also work when run as root
- golden-string comparisons for formatter output
- a `failingWriter` to check that write errors propagate
