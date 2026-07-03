package ui

import (
	"strings"
	"testing"
)

func TestWrapLineWidthZero(t *testing.T) {
	long := strings.Repeat("hello world ", 20)
	got := WrapLine(long, 0, "")
	if !strings.Contains(got, "\n") {
		t.Fatalf("width 0 should clamp to minWrapWidth and wrap: %q", got)
	}
	for _, line := range strings.Split(got, "\n") {
		if len(line) > minWrapWidth {
			t.Fatalf("width 0 line exceeds minWrapWidth %d: %q (%d)", minWrapWidth, line, len(line))
		}
	}
}

func TestWrapLineNegativeWidthDoesNotOverflow(t *testing.T) {
	long := strings.Repeat("hello world ", 20)
	got := WrapLine(long, -50, IndentBody)
	for _, line := range strings.Split(got, "\n") {
		if len(line) > minWrapWidth {
			t.Fatalf("negative width line exceeds minWrapWidth %d: %q (%d)", minWrapWidth, line, len(line))
		}
	}
}

func TestWrapLineShorterThanWidth(t *testing.T) {
	got := WrapLine("short", 80, IndentBody)
	if got != "short" {
		t.Fatalf("got %q, want %q", got, "short")
	}
}

func TestWrapLineWrapsAtWordBoundary(t *testing.T) {
	got := WrapLine("hello world this is", 12, IndentBody)
	want := "hello world" + "\n" + IndentBody + "this is"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWrapLineSingleLongWordNotSplit(t *testing.T) {
	got := WrapLine("ABCDEFGHIJKLMNOPQRSTUVWXYZ", 10, IndentBody)
	if strings.Contains(got, "\n") {
		t.Fatalf("long word should not be split: %q", got)
	}
}

func TestWrapLineEmpty(t *testing.T) {
	got := WrapLine("", 80, IndentBody)
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestWrapLineWhitespaceOnly(t *testing.T) {
	got := WrapLine("   ", 80, IndentBody)
	if got != "" {
		t.Fatalf("got %q, want empty", got)
	}
}

func TestWrapLineMultipleWraps(t *testing.T) {
	got := WrapLine("one two three four five six seven eight nine ten", 14, ">>")
	want := "one two three\n>>four five six\n>>seven eight\n>>nine ten"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWrapLineIndentPreserved(t *testing.T) {
	got := WrapLine("a bb ccc dddd eeeee ffffff ggggggg", 10, IndentBlock)
	lines := strings.Split(got, "\n")
	for i := 1; i < len(lines); i++ {
		if !strings.HasPrefix(lines[i], IndentBlock) {
			t.Fatalf("continuation line %d missing indent: %q", i, lines[i])
		}
		if len(strings.TrimPrefix(lines[i], IndentBlock)) > 10 {
			t.Fatalf("continuation line %d exceeds width: %q", i, lines[i])
		}
	}
}

func TestWrapLineLongWordThenShort(t *testing.T) {
	got := WrapLine("ABCDEFGHIJKLMNOPQRSTUVWXYZ short", 10, IndentBody)
	want := "ABCDEFGHIJKLMNOPQRSTUVWXYZ\n" + IndentBody + "short"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWrapLineExactWidth(t *testing.T) {
	got := WrapLine("1234567890", 10, IndentBody)
	if got != "1234567890" {
		t.Fatalf("got %q, want exact width no wrap", got)
	}
}

func TestWrapLineJustOverWidth(t *testing.T) {
	got := WrapLine("1234567890 a", 10, IndentBody)
	want := "1234567890\n" + IndentBody + "a"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestWrapLineWidth80(t *testing.T) {
	s := "A PKGBUILD source entry labelled 'patches' points at a personal GitHub " +
		"repo unrelated to the upstream project and is executed during build."
	got := WrapLine(s, 80, "         ")
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapping at width 80, got %d lines: %q", len(lines), got)
	}
	for i, line := range lines {
		content := line
		if i > 0 {
			if !strings.HasPrefix(line, "         ") {
				t.Fatalf("continuation line %d missing indent: %q", i, line)
			}
			content = strings.TrimPrefix(line, "         ")
		}
		if len(content) > 80 {
			t.Fatalf("line %d exceeds width 80: %q (%d chars)", i, line, len(content))
		}
	}
}

func TestWrapLineWidth120(t *testing.T) {
	s := "The build script downloads and executes a binary from a CDN URL that was " +
		"registered less than 24 hours ago, uses a self-signed certificate, and " +
		"obfuscates the download command with base64 decoding before piping to bash."
	got := WrapLine(s, 120, "         ")
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapping at width 120, got %d lines: %q", len(lines), got)
	}
	for i, line := range lines {
		content := line
		if i > 0 {
			if !strings.HasPrefix(line, "         ") {
				t.Fatalf("continuation line %d missing indent: %q", i, line)
			}
			content = strings.TrimPrefix(line, "         ")
		}
		if len(content) > 120 {
			t.Fatalf("line %d exceeds width 120: %q (%d chars)", i, line, len(content))
		}
	}
}

func TestWrapQuotePrefix(t *testing.T) {
	s := "the file downloads and executes a remote binary from a URL that was registered less than 24 hours ago"
	const indent = "             > "
	got := WrapLine(s, 40, indent)
	lines := strings.Split(got, "\n")
	for i, line := range lines {
		if i > 0 && !strings.HasPrefix(line, indent) {
			t.Fatalf("continuation line %d missing quote indent: %q", i, line)
		}
		content := line
		if i > 0 {
			content = strings.TrimPrefix(line, indent)
		}
		if len(content) > 40 {
			t.Fatalf("line %d content exceeds width 40: %q (%d)", i, content, len(content))
		}
	}
}

func TestWrapUsageLine(t *testing.T) {
	s := "in: 500 · out: 320 · cost: $0.05  in: 600 · out: 400 · cost: $0.08"
	got := WrapLine(s, 30, "         ↳ ")
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %q", got)
	}
	for i, line := range lines {
		if i > 0 {
			if !strings.HasPrefix(line, "         ↳ ") {
				t.Fatalf("continuation line %d missing usage indent: %q", i, line)
			}
		}
	}
}

func TestWrapBlockMessage(t *testing.T) {
	msg := "5 package(s) flagged MALICIOUS with multiple findings across several PKGBUILDs and install scripts"
	got := WrapLine(msg, 50, IndentBlock)
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %q", got)
	}
	for i, line := range lines {
		if i > 0 {
			if !strings.HasPrefix(line, IndentBlock) {
				t.Fatalf("continuation line %d missing block indent: %q", i, line)
			}
			content := strings.TrimPrefix(line, IndentBlock)
			if len(content) > 50 {
				t.Fatalf("continuation line %d exceeds width 50: %q (%d chars)", i, line, len(content))
			}
		}
	}
}

func TestWrapHookMessage(t *testing.T) {
	s := "Plain `yay` (v14) will now scan AUR packages after download, before build, with automatic malware detection enabled by default for all future runs"
	got := WrapLine(s, 70, "      ")
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %q", got)
	}
	for i, line := range lines {
		if i > 0 {
			content := strings.TrimPrefix(line, "      ")
			if len(content) > 70 {
				t.Fatalf("continuation line %d exceeds width 70: %q (%d chars)", i, line, len(content))
			}
		}
	}
}

func TestWrapScannerUsage(t *testing.T) {
	s := "scanner usage: 5 call(s) · in: 2048 · out: 1536 · cost: $0.10  in: 4096 · out: 3072 · cost: $0.20"
	got := WrapLine(s, 50, "scanner usage: ")
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %q", got)
	}
	for i, line := range lines {
		if i > 0 {
			if !strings.HasPrefix(line, "scanner usage: ") {
				t.Fatalf("continuation line %d missing scanner usage indent: %q", i, line)
			}
		}
	}
}

func TestWrapReportInstructions(t *testing.T) {
	s := "Review it carefully with your security team, then email it to aurscan@manticore-projects.com for triage and upstream disclosure coordination"
	got := WrapLine(s, 50, "   ")
	lines := strings.Split(got, "\n")
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %q", got)
	}
	for i, line := range lines {
		if i > 0 {
			if !strings.HasPrefix(line, "   ") {
				t.Fatalf("continuation line %d missing report indent: %q", i, line)
			}
		}
	}
}

// TestWrapLineNoTrailingSpacesMultiSpaceRun verifies that when the input
// contains a run of consecutive spaces (an LLM artifact), WrapLine does not
// emit trailing whitespace on any line. LastIndexByte may land inside the
// space run; the emitted line must end at the last non-space character.
// Reproduces the user-reported symptom: a finding line ending in ~47 spaces
// followed by a short orphan line.
func TestWrapLineNoTrailingSpacesMultiSpaceRun(t *testing.T) {
	s := "The patch replaces a checksummed registry crate with a Git branch " +
		"dependency. Cargo.lock pins commit 98aedbd, but the supplied files " +
		"cannot independently establish that this repository is" +
		strings.Repeat(" ", 47) +
		"the legitimate upstream."
	got := WrapLine(s, 117, IndentBody)
	lines := strings.Split(got, "\n")
	for i, line := range lines {
		if strings.HasSuffix(line, " ") {
			t.Fatalf("line %d has trailing spaces (len %d): %q", i, len(line), line)
		}
	}
}

// TestWrapLineNoTrailingSpacesSingleSpace verifies that normal single-space
// input never produces trailing whitespace — a regression guard for any
// future change to the cut logic.
func TestWrapLineNoTrailingSpacesSingleSpace(t *testing.T) {
	s := "The build script downloads and executes a binary from a CDN URL that " +
		"was registered less than 24 hours ago, uses a self-signed certificate, " +
		"and obfuscates the download command with base64 decoding before piping " +
		"to bash."
	for _, w := range []int{40, 60, 80, 93, 100, 120, 142} {
		got := WrapLine(s, w, IndentBody)
		for i, line := range strings.Split(got, "\n") {
			if strings.HasSuffix(line, " ") {
				t.Fatalf("width %d line %d has trailing spaces: %q", w, i, line)
			}
		}
	}
}

// TestWrapLineNoOrphanAfterSpaceRun verifies that a multi-space run does
// not produce a ragged last line (e.g. "legitimate upstream." alone) while
// the previous line holds a stretched-out space run. When a space run is
// collapsed, content should flow to fill the line naturally.
func TestWrapLineNoOrphanAfterSpaceRun(t *testing.T) {
	s := "The patch replaces a checksummed registry crate with a Git branch " +
		"dependency. Cargo.lock pins commit 98aedbd, but the supplied files " +
		"cannot independently establish that this repository is" +
		strings.Repeat(" ", 47) +
		"the legitimate upstream."
	for _, w := range []int{119, 120, 121, 122, 142} {
		got := WrapLine(s, w, IndentBody)
		lines := strings.Split(got, "\n")
		for i, line := range lines {
			content := strings.TrimPrefix(line, IndentBody)
			// No line should contain an internal run of 2+ spaces when
			// the input has a space run that should have been collapsed.
			if strings.Contains(content, "  ") {
				t.Fatalf("width %d line %d contains an internal space run (%d chars): %q", w, i, len(content), content)
			}
		}
	}
}

// TestWrapLineCollapsesInternalSpaceRun verifies that an internal run of
// spaces in the input is treated as a single word boundary for wrapping
// purposes — no trailing spaces on the line before the run, and no internal
// space run in any output line's content (excluding the caller's indent).
func TestWrapLineCollapsesInternalSpaceRun(t *testing.T) {
	s := "word1 word2" + strings.Repeat(" ", 10) + "word3 word4"
	got := WrapLine(s, 20, IndentBody)
	lines := strings.Split(got, "\n")
	for i, line := range lines {
		if strings.HasSuffix(line, " ") {
			t.Fatalf("line %d has trailing spaces: %q", i, line)
		}
		// Check the content after the indent, not the full line: the indent
		// itself is a legitimate run of spaces.
		content := strings.TrimPrefix(line, IndentBody)
		if strings.Contains(content, "  ") {
			t.Fatalf("line %d content contains an internal space run: %q", i, content)
		}
	}
}

// TestWrapLineTabIsWordBoundary verifies that a tab in the input is treated
// as a word boundary (equivalent to a single space), not emitted literally
// or allowed to create trailing whitespace.
func TestWrapLineTabIsWordBoundary(t *testing.T) {
	s := "repository is\tthe legitimate upstream."
	got := WrapLine(s, 30, IndentBody)
	lines := strings.Split(got, "\n")
	for i, line := range lines {
		if strings.HasSuffix(line, " ") || strings.HasSuffix(line, "\t") {
			t.Fatalf("line %d has trailing whitespace: %q", i, line)
		}
		if strings.Contains(line, "\t") {
			t.Fatalf("line %d contains a literal tab: %q", i, line)
		}
	}
}

// TestWrapLineNewlineIsWordBoundary verifies that an embedded newline in the
// input (an LLM artifact) is treated as a word boundary, not emitted as a
// literal line break that creates a degenerate short line.
func TestWrapLineNewlineIsWordBoundary(t *testing.T) {
	s := "The patch replaces a checksummed registry crate with a Git branch " +
		"dependency. Cargo.lock pins commit 98aedbd, but the supplied files " +
		"cannot independently establish that this repository is\nthe legitimate upstream."
	got := WrapLine(s, 93, IndentBody)
	lines := strings.Split(got, "\n")
	for i, line := range lines {
		content := strings.TrimPrefix(line, IndentBody)
		if i > 0 && len(content) > 0 && len(content) <= 3 {
			t.Fatalf("line %d is a short orphan (%d chars) from embedded newline: %q", i, len(content), content)
		}
	}
}

// TestWrapLineANSIDoesNotInflateWidth verifies that if an ANSI escape sequence
// leaks into f.Why, it does not corrupt the visible width budget. WrapLine
// operates on bytes; an embedded \x1b[0m (4 bytes) must not consume 4 chars of
// the visible width or be emitted as trailing/leading invisible bytes.
func TestWrapLineANSIDoesNotInflateWidth(t *testing.T) {
	s := "The patch replaces a checksummed registry crate with a Git branch " +
		"dependency. Cargo.lock pins commit 98aedbd, but the supplied files " +
		"cannot independently establish that this repository is\x1b[0m the legitimate upstream."
	got := WrapLine(s, 93, IndentBody)
	for i, line := range strings.Split(got, "\n") {
		if strings.Contains(line, "\x1b") {
			t.Fatalf("line %d contains an ANSI escape byte: %q", i, line)
		}
	}
}

// TestWrapLineNormalizedHasNoLeadingSpace verifies that normalizeWrapInput
// strips leading whitespace, so lastSpaceBefore never sees a space at byte
// offset 0 (which would make cut==0 and bypass the long-word fallback).
func TestWrapLineNormalizedHasNoLeadingSpace(t *testing.T) {
	for _, s := range []string{" lead", "  lead", "\tlead", "\nlead"} {
		got := WrapLine(s, 80, IndentBody)
		if strings.HasPrefix(got, " ") || strings.HasPrefix(got, "\t") {
			t.Fatalf("output has leading whitespace for input %q: %q", s, got)
		}
	}
}

// TestWrapLineNonASCIIDoesNotInflateWidth verifies that a non-ASCII character
// (e.g. an em dash, 3 bytes in UTF-8) is measured by its visible rune width
// (1), not its byte length (3), so it does not under-fill a line.
func TestWrapLineNonASCIIDoesNotInflateWidth(t *testing.T) {
	// "word — word" is 11 runes (10 visible + 1 em dash), 13 bytes. With
	// width=11 it should fit on one line without wrapping.
	s := "word — word"
	got := WrapLine(s, 11, IndentBody)
	if strings.Contains(got, "\n") {
		t.Fatalf("em dash line wrapped unexpectedly (byte vs rune width mismatch): %q", got)
	}
}
