# Mirror Tree

GoLang tool capable to mirror a given root dir and its contents to a target dir,
creating mock files and directories.

## Why Mock mirroring?

There are some scenario where only tree structure and filename matters, for example:

- check on file presence
- check on rolling rules triggered by filename
- archive rules based on filename
- ecc..

On all this scenario, having a Tree Mockup can help simulating
the target dir structure without actually creating files.

## Status

Prototype: `mtree` builds the mirror plan and prints it (dry run).
Nothing is written to disk yet.

## Install

```bash
go install github.com/anto6715/MirrorTree/cmd/mtree@latest
```

or build from a checkout:

```bash
go build -o mtree ./cmd/mtree
```

## Usage

```
mtree <path> <dest> [flags]

Arguments:
  <path>    Root directory to mirror
  <dest>    Destination directory

Flags:
  -h, --help            Show context-sensitive help.
  -L, --max-depth=-1    Maximum depth to mirror (-1 for unlimited)
  -f, --filter="*"      Filter pattern to match files
      --no-color        Disable colored output
```

Example:

```
$ mtree -f '*.nc' /data mirror
Mirror plan
  source: /data
  target: mirror

mirror
├── a.nc
├── database
│   └── x.nc
└── sub
    └── deep
        └── c.nc

3 directories, 3 files
Dry run: nothing was created
```

### Behaviour

- `<dest>` must not exist yet. `<path>` must be a directory; symlinks in it are resolved.
- `--filter` is a shell glob (`filepath.Match` syntax) matched against file names only,
  not their directories. Quote it so the shell does not expand it.
- `--max-depth 0` mirrors only the files directly inside `<path>`.
- Only directories containing at least one matching file are mirrored: empty directories are skipped.
- Symlinks are not followed, and unreadable directories are skipped.
- Colors are used only when the output is a terminal. Disable them with `--no-color`
  or by setting `NO_COLOR`.

## Development

```bash
go test ./...
go vet ./...
```

Directory walking comes from the `find` package of
[goat](https://github.com/anto6715/goat). To work on both repositories at the
same time, create a local, untracked workspace:

```bash
go work init . ../goat
```
