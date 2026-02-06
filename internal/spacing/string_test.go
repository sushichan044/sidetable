package spacing_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/sushichan044/sidetable/internal/spacing"
)

func TestFormatter_BasicUsage(t *testing.T) {
	formatter := spacing.NewFormatter(
		spacing.Column(),
		spacing.MinSpacing(2),
		spacing.Column(),
		spacing.MinSpacing(4),
		spacing.Column(),
	)

	err := formatter.AddRows(
		[]string{"init", "i", "Initialize the sidetable for the project."},
		[]string{"add", "a", "Add a new directory to the sidetable."},
		[]string{"remove", "rm", "Remove a directory from the sidetable."},
		[]string{"list", "ls", "List all managed directories."},
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = formatter.Format(&buf)
	require.NoError(t, err)

	expected := strings.TrimSpace(`
init    i     Initialize the sidetable for the project.
add     a     Add a new directory to the sidetable.
remove  rm    Remove a directory from the sidetable.
list    ls    List all managed directories.
`)

	actual := strings.TrimSpace(buf.String())

	assert.Equal(t, expected, actual)
}

func TestFormatter_MultiByteCharacters(t *testing.T) {
	formatter := spacing.NewFormatter(
		spacing.Column(),
		spacing.MinSpacing(2),
		spacing.Column(),
		spacing.MinSpacing(2),
		spacing.Column(),
	)

	err := formatter.AddRows(
		[]string{"あいう", "a", "Description 1"},
		[]string{"x", "えお", "Description 2"},
		[]string{"y", "z", "説明３"},
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = formatter.Format(&buf)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 3)

	// Check that columns align properly
	// "あいう" has display width 6 (3 full-width chars)
	// First column should be 6 chars wide, then 2 spaces, then second column starts
	assert.True(t, strings.HasPrefix(lines[0], "あいう  a"))
	assert.True(t, strings.HasPrefix(lines[1], "x       えお"))
}

func TestFormatter_NoTrailingSpaces(t *testing.T) {
	formatter := spacing.NewFormatter(
		spacing.Column(),
		spacing.MinSpacing(2),
		spacing.Column(),
	)

	err := formatter.AddRows(
		[]string{"short", "text"},
		[]string{"longer", "content"},
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = formatter.Format(&buf)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	for _, line := range lines {
		assert.NotRegexp(t, ` $`, line, "Line should not have trailing spaces")
	}
}

func TestFormatter_NoSpacingBetweenConsecutiveColumns(t *testing.T) {
	formatter := spacing.NewFormatter(
		spacing.Column(),
		spacing.Column(),
		spacing.MinSpacing(2),
		spacing.Column(),
	)

	err := formatter.AddRows(
		[]string{"a", "b", "c"},
		[]string{"xx", "yy", "zz"},
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = formatter.Format(&buf)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 2)

	// First two columns should have no spacing between them (immediately adjacent)
	// But they should be padded as a group to align with the longest row's group width
	// Group 1 (Col 0 + Col 1): max width = 2 + 2 = 4
	// MinSpacing(2): 2 spaces
	// Group 2 (Col 2): starts at position 6
	//
	// Row 1: "a" (1) + "b" (1) = "ab" (2), padded to 4 = "ab  ", + 2 spaces = "ab    ", + "c" = "ab    c"
	// Row 2: "xx" (2) + "yy" (2) = "xxyy" (4), no padding needed, + 2 spaces = "xxyy  ", + "zz" = "xxyy  zz"
	assert.Equal(t, "ab    c", lines[0])
	assert.Equal(t, "xxyy  zz", lines[1])
}

func TestFormatter_SingleColumn(t *testing.T) {
	formatter := spacing.NewFormatter(spacing.Column())

	err := formatter.AddRows(
		[]string{"first"},
		[]string{"second"},
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = formatter.Format(&buf)
	require.NoError(t, err)

	expected := "first\nsecond\n"
	assert.Equal(t, expected, buf.String())
}

func TestFormatter_EmptyRows(t *testing.T) {
	formatter := spacing.NewFormatter(spacing.Column(), spacing.MinSpacing(2), spacing.Column())

	err := formatter.AddRows(
		[]string{"", "content"},
		[]string{"text", ""},
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = formatter.Format(&buf)
	require.NoError(t, err)

	// Don't trim the buffer - we want to preserve leading spaces on each line
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n")
	require.Len(t, lines, 2)

	// Column 0 max width: 4 ("text")
	// MinSpacing(2): 2 spaces
	// Column 1 starts at position 6
	//
	// Row 1: "" (width 0), padded to 4 = "    ", + MinSpacing(2) = "      ", + "content" = "      content"
	// Row 2: "text" (width 4), no padding, + MinSpacing(2) = "text  ", + "" = "text  " → after TrimRight = "text"

	assert.Equal(t, "      content", lines[0])
	assert.Equal(t, "text", lines[1])
}

func TestFormatter_IncorrectColumnCount(t *testing.T) {
	formatter := spacing.NewFormatter(spacing.Column(), spacing.MinSpacing(2), spacing.Column())

	err := formatter.AddRows(
		[]string{"a", "b"},
		[]string{"x"},               // Too few columns
		[]string{"y", "z", "extra"}, // Too many columns
	)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "incorrect column count")

	// Verify both invalid rows are reported
	assert.Contains(t, err.Error(), "[x]")
	assert.Contains(t, err.Error(), "[y z extra]")
}

func TestFormatter_NoRows(t *testing.T) {
	formatter := spacing.NewFormatter(spacing.Column(), spacing.MinSpacing(2), spacing.Column())

	var buf bytes.Buffer
	err := formatter.Format(&buf)
	require.NoError(t, err)

	assert.Empty(t, buf.String())
}

func TestFormatter_ZeroMinSpacing(t *testing.T) {
	formatter := spacing.NewFormatter(
		spacing.Column(),
		spacing.MinSpacing(0),
		spacing.Column(),
	)

	err := formatter.AddRows(
		[]string{"a", "b"},
		[]string{"xx", "yy"},
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = formatter.Format(&buf)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 2)

	// With MinSpacing(0), columns should be immediately adjacent after first column width
	assert.Equal(t, "ab", lines[0])
	assert.Equal(t, "xxyy", lines[1])
}

func TestFormatter_NegativeMinSpacing(t *testing.T) {
	// Negative spacing should be treated as 0
	formatter := spacing.NewFormatter(
		spacing.Column(),
		spacing.MinSpacing(-5),
		spacing.Column(),
	)

	err := formatter.AddRows(
		[]string{"a", "b"},
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = formatter.Format(&buf)
	require.NoError(t, err)

	assert.Equal(t, "ab\n", buf.String())
}

func TestFormatter_MixedWidthCharacters(t *testing.T) {
	formatter := spacing.NewFormatter(
		spacing.Column(),
		spacing.MinSpacing(2),
		spacing.Column(),
		spacing.MinSpacing(2),
		spacing.Column(),
	)

	err := formatter.AddRows(
		[]string{"日本語", "abc", "test"},
		[]string{"eng", "日本語", "テスト"},
		[]string{"x", "y", "ｚ"}, // Half-width katakana
	)
	require.NoError(t, err)

	var buf bytes.Buffer
	err = formatter.Format(&buf)
	require.NoError(t, err)

	lines := strings.Split(strings.TrimSpace(buf.String()), "\n")
	require.Len(t, lines, 3)

	// Verify no trailing spaces
	for _, line := range lines {
		assert.NotRegexp(t, ` $`, line)
	}
}
