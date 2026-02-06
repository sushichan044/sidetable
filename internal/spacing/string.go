package spacing

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/mattn/go-runewidth"
)

// elementType represents the type of element in the formatter specification.
type elementType int

const (
	elementColumn elementType = iota
	elementMinSpacing
)

// Element represents a single Element in the formatter specification.
type Element struct {
	typ     elementType
	spacing int // Only used for elementMinSpacing
}

// Column creates a column element for the formatter specification.
func Column() Element {
	return Element{typ: elementColumn}
}

// MinSpacing creates a minimum spacing element between columns.
// Negative values are treated as 0.
func MinSpacing(n int) Element {
	if n < 0 {
		n = 0
	}
	return Element{typ: elementMinSpacing, spacing: n}
}

// Formatter builds and formats columnar text with configurable spacing.
type Formatter struct {
	elements     []Element
	columnCount  int
	rows         [][]string
	maxColWidths []int // Maximum display width for each column
}

// NewFormatter creates a new Formatter with the specified column and spacing configuration.
// Use Column() and MinSpacing(n) to define the layout.
// Example: NewFormatter(Column(), MinSpacing(2), Column(), MinSpacing(4), Column()).
func NewFormatter(elements ...Element) *Formatter {
	columnCount := 0
	for _, elem := range elements {
		if elem.typ == elementColumn {
			columnCount++
		}
	}

	return &Formatter{
		elements:    elements,
		columnCount: columnCount,
	}
}

// AddRows adds multiple rows to the formatter.
// Each row must have exactly the same number of columns as specified by Column() calls.
// Returns an error if any row has an incorrect number of columns.
func (f *Formatter) AddRows(rows ...[]string) error {
	var errs []error

	for _, row := range rows {
		if len(row) != f.columnCount {
			errs = append(
				errs,
				fmt.Errorf("row %v has incorrect column count, expected %d but got %d", row, f.columnCount, len(row)),
			)
			continue
		}

		f.rows = append(f.rows, row)

		// Update maximum column widths
		if len(f.maxColWidths) < len(row) {
			f.maxColWidths = make([]int, len(row))
		}
		for i, cell := range row {
			width := runewidth.StringWidth(cell)
			if width > f.maxColWidths[i] {
				f.maxColWidths[i] = width
			}
		}
	}

	if len(errs) > 0 {
		return errors.Join(errs...)
	}
	return nil
}

// Println writes the formatted output to the provided writer.
// Each row is aligned according to the column specifications and spacing requirements.
func (f *Formatter) Println(output io.Writer) error {
	if len(f.rows) == 0 {
		return nil
	}

	// Format and write each row
	for _, row := range f.rows {
		line := f.formatRow(row)
		if _, err := fmt.Fprintln(output, line); err != nil {
			return err
		}
	}

	return nil
}

// formatRow formats a single row based on element specifications.
// Columns without MinSpacing between them are placed immediately after each other (no spacing).
// When MinSpacing is encountered, it aligns all rows by padding the previous column group
// to its maximum width, then adds the specified minimum spacing.
func (f *Formatter) formatRow(row []string) string {
	if len(row) == 0 {
		return ""
	}

	var builder strings.Builder
	currentDisplayWidth := 0
	colIdx := 0
	pendingMinSpacing := 0
	groupStartCol := 0
	groupStartPos := 0

	for _, elem := range f.elements {
		switch elem.typ {
		case elementColumn:
			if colIdx >= len(row) {
				break
			}

			cell := row[colIdx]
			cellWidth := runewidth.StringWidth(cell)

			if pendingMinSpacing > 0 {
				// Calculate the total maximum width of the previous column group
				groupMaxWidth := 0
				for i := groupStartCol; i < colIdx; i++ {
					groupMaxWidth += f.maxColWidths[i]
				}

				// Pad to the group's maximum width (from group start position)
				targetPos := groupStartPos + groupMaxWidth
				if targetPos > currentDisplayWidth {
					builder.WriteString(strings.Repeat(" ", targetPos-currentDisplayWidth))
					currentDisplayWidth = targetPos
				}

				// Add the minimum spacing
				builder.WriteString(strings.Repeat(" ", pendingMinSpacing))
				currentDisplayWidth += pendingMinSpacing

				// Start new group
				groupStartCol = colIdx
				groupStartPos = currentDisplayWidth
				pendingMinSpacing = 0
			}

			// Within a group, columns are placed immediately after each other
			builder.WriteString(cell)
			currentDisplayWidth += cellWidth
			colIdx++

		case elementMinSpacing:
			// Store the spacing to be applied before the next column
			pendingMinSpacing = elem.spacing
		}
	}

	return strings.TrimRight(builder.String(), " ")
}
