package mtreeapp

import (
	"bufio"
	"fmt"
	"io"
	"maps"
	"slices"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type formatOptions struct {
	color bool
}

func writeMirrorPlan(w io.Writer, plan mirrorPlan, opts formatOptions) error {
	st := style{enabled: opts.color}
	tree := buildTree(plan)
	output := bufio.NewWriter(w)

	fmt.Fprintln(output, st.title("Mirror plan"))
	fmt.Fprintf(output, "  %s %s\n", st.label("source:"), displayName(plan.source))
	fmt.Fprintf(output, "  %s %s\n", st.label("target:"), displayName(plan.target))
	fmt.Fprintln(output)

	if tree.files == 0 {
		fmt.Fprintf(output, "No files match filter %q\n", plan.filter)
	} else {
		fmt.Fprintln(output, st.dir(displayName(tree.root.name)))
		writeTree(output, tree.root, "", st)
	}

	fmt.Fprintln(output)
	fmt.Fprintf(output, "%s, %s\n",
		plural(tree.dirs, "directory", "directories"),
		plural(tree.files, "file", "files"),
	)
	fmt.Fprintln(output, st.warn("Dry run: nothing was created"))

	return output.Flush()
}

// writeTree prints the children of node sorted by name, drawing the branches
// with prefix carrying the indentation of the ancestors.
func writeTree(output io.Writer, node *treeNode, prefix string, st style) {
	names := slices.Sorted(maps.Keys(node.children))
	for i, name := range names {
		child := node.children[name]

		connector, indent := "├── ", "│   "
		if i == len(names)-1 {
			connector, indent = "└── ", "    "
		}

		label := displayName(child.name)
		if child.isDir {
			label = st.dir(label)
		}
		fmt.Fprintf(output, "%s%s%s\n", prefix, connector, label)

		if child.isDir {
			writeTree(output, child, prefix+indent, st)
		}
	}
}

// displayName quotes names with invalid UTF-8 or non-printable characters, so
// they cannot break the tree layout or inject terminal escape sequences.
func displayName(name string) string {
	if !utf8.ValidString(name) || strings.ContainsFunc(name, func(r rune) bool { return !unicode.IsPrint(r) }) {
		return strconv.Quote(name)
	}
	return name
}

func plural(n int, singular, plural string) string {
	if n == 1 {
		return fmt.Sprintf("%d %s", n, singular)
	}
	return fmt.Sprintf("%d %s", n, plural)
}

const (
	ansiReset    = "\x1b[0m"
	ansiBold     = "\x1b[1m"
	ansiDim      = "\x1b[2m"
	ansiBoldBlue = "\x1b[1;34m"
	ansiYellow   = "\x1b[33m"
)

// style wraps text in ANSI codes when enabled, and returns it unchanged
// otherwise.
type style struct {
	enabled bool
}

func (s style) title(text string) string { return s.wrap(ansiBold, text) }
func (s style) label(text string) string { return s.wrap(ansiDim, text) }
func (s style) dir(text string) string   { return s.wrap(ansiBoldBlue, text) }
func (s style) warn(text string) string  { return s.wrap(ansiYellow, text) }

func (s style) wrap(code, text string) string {
	if !s.enabled {
		return text
	}
	return code + text + ansiReset
}
