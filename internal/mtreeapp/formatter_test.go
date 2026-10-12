package mtreeapp

import (
	"errors"
	"strings"
	"testing"
)

func TestWriteMirrorPlan(t *testing.T) {
	plan := testPlan(t, "sub/deep/c.nc", "a.nc", "sub/b.txt", "database/x.nc")

	var output strings.Builder
	if err := writeMirrorPlan(&output, plan, formatOptions{}); err != nil {
		t.Fatalf("writeMirrorPlan returned an error: %v", err)
	}

	want := "Mirror plan\n" +
		"  source: /data\n" +
		"  target: mirror\n" +
		"\n" +
		"mirror\n" +
		"├── a.nc\n" +
		"├── database\n" +
		"│   └── x.nc\n" +
		"└── sub\n" +
		"    ├── b.txt\n" +
		"    └── deep\n" +
		"        └── c.nc\n" +
		"\n" +
		"3 directories, 4 files\n" +
		"Dry run: nothing was created\n"

	if output.String() != want {
		t.Errorf("output:\n%s\nwant:\n%s", output.String(), want)
	}
}

func TestWriteMirrorPlanEmpty(t *testing.T) {
	plan := testPlan(t)
	plan.filter = "*.nc"

	var output strings.Builder
	if err := writeMirrorPlan(&output, plan, formatOptions{}); err != nil {
		t.Fatalf("writeMirrorPlan returned an error: %v", err)
	}

	want := "Mirror plan\n" +
		"  source: /data\n" +
		"  target: mirror\n" +
		"\n" +
		"No files match filter \"*.nc\"\n" +
		"\n" +
		"0 directories, 0 files\n" +
		"Dry run: nothing was created\n"

	if output.String() != want {
		t.Errorf("output:\n%s\nwant:\n%s", output.String(), want)
	}
}

func TestWriteMirrorPlanSingular(t *testing.T) {
	var output strings.Builder
	if err := writeMirrorPlan(&output, testPlan(t, "sub/a.nc"), formatOptions{}); err != nil {
		t.Fatalf("writeMirrorPlan returned an error: %v", err)
	}

	if !strings.Contains(output.String(), "\n1 directory, 1 file\n") {
		t.Errorf("output does not contain singular summary:\n%s", output.String())
	}
}

func TestWriteMirrorPlanColor(t *testing.T) {
	plan := testPlan(t, "sub/a.nc")

	var colored strings.Builder
	if err := writeMirrorPlan(&colored, plan, formatOptions{color: true}); err != nil {
		t.Fatalf("writeMirrorPlan returned an error: %v", err)
	}

	for _, want := range []string{
		ansiBold + "Mirror plan" + ansiReset,
		ansiDim + "source:" + ansiReset,
		ansiBoldBlue + "mirror" + ansiReset + "\n",
		"└── " + ansiBoldBlue + "sub" + ansiReset + "\n",
		"    └── a.nc\n",
		ansiYellow + "Dry run: nothing was created" + ansiReset,
	} {
		if !strings.Contains(colored.String(), want) {
			t.Errorf("colored output does not contain %q:\n%q", want, colored.String())
		}
	}

	var plain strings.Builder
	if err := writeMirrorPlan(&plain, plan, formatOptions{}); err != nil {
		t.Fatalf("writeMirrorPlan returned an error: %v", err)
	}
	if strings.Contains(plain.String(), "\x1b") {
		t.Errorf("plain output contains ANSI codes: %q", plain.String())
	}
}

func TestWriteMirrorPlanQuotesUnsafeNames(t *testing.T) {
	plan := testPlan(t, "bad\nname.nc", "\x1b[31mred.nc")

	var output strings.Builder
	if err := writeMirrorPlan(&output, plan, formatOptions{}); err != nil {
		t.Fatalf("writeMirrorPlan returned an error: %v", err)
	}

	if strings.Contains(output.String(), "\x1b") || strings.Contains(output.String(), "bad\nname") {
		t.Errorf("output contains raw unsafe names: %q", output.String())
	}
	for _, want := range []string{`"bad\nname.nc"`, `"\x1b[31mred.nc"`} {
		if !strings.Contains(output.String(), want) {
			t.Errorf("output does not contain quoted name %s:\n%s", want, output.String())
		}
	}
}

func TestDisplayName(t *testing.T) {
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "plain", in: "a.nc", want: "a.nc"},
		{name: "space", in: "my file.nc", want: "my file.nc"},
		{name: "non-ascii", in: "café.nc", want: "café.nc"},
		{name: "newline", in: "a\nb", want: `"a\nb"`},
		{name: "tab", in: "a\tb", want: `"a\tb"`},
		{name: "escape sequence", in: "\x1b[31m", want: `"\x1b[31m"`},
		{name: "invalid utf-8", in: "a\x9bb", want: `"a\x9bb"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := displayName(tt.in); got != tt.want {
				t.Errorf("displayName(%q) = %s, want %s", tt.in, got, tt.want)
			}
		})
	}
}

func TestWriteMirrorPlanReturnsWriterError(t *testing.T) {
	err := writeMirrorPlan(failingWriter{}, testPlan(t, "a.nc"), formatOptions{})

	if !errors.Is(err, errWriteFailed) {
		t.Fatalf("writeMirrorPlan error = %v, want %v", err, errWriteFailed)
	}
}

var errWriteFailed = errors.New("write failed")

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) {
	return 0, errWriteFailed
}
