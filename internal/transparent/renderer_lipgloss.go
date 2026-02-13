package transparent

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// lipglossMaxValueWidth is the maximum width for the Value column when
// rendering with lipgloss. Lipgloss handles wrapping internally, but we
// still truncate extremely long single-line values to keep output readable.
const lipglossMaxValueWidth = 80

// renderVarsLipgloss renders the variable table using the charmbracelet/lipgloss
// table package, which handles column sizing, borders, and multiline content
// natively.
func renderVarsLipgloss(w io.Writer, vars []VarTrace) {
	fmt.Fprintf(w, "  Variables in scope:\n")

	rows := make([][]string, 0, len(vars))
	for _, v := range vars {
		name := v.Name
		origin := originLabel(v.Origin)

		typeStr := v.Type
		if typeStr == "" {
			typeStr = "-"
		}

		// Value
		valStr := fmt.Sprintf("%v", v.Value)
		if v.IsDynamic {
			shInfo := ""
			if v.ShCmd != "" {
				shInfo = fmt.Sprintf(" (sh: %s)", v.ShCmd)
			}
			valStr = fmt.Sprintf("(sh) %s%s", valStr, shInfo)
			if fmt.Sprintf("%v", v.Value) == "" {
				valStr += " ⚠ DYNAMIC — not evaluated"
			}
		}
		if v.IsRef {
			valStr = fmt.Sprintf("(ref) %s", valStr)
		}

		// Truncate extremely long values (lipgloss wraps, but we still cap)
		valStr = truncateLipglossValue(valStr, lipglossMaxValueWidth)

		// Extra info lines appended to value
		var extras []string
		if v.ValueID != 0 {
			extras = append(extras, fmt.Sprintf("ptr: 0x%x", v.ValueID))
		}
		if v.RefName != "" {
			extras = append(extras, fmt.Sprintf("→ aliases: %s", v.RefName))
		}
		if len(extras) > 0 {
			valStr += "\n" + strings.Join(extras, "\n")
		}

		// Shadow
		shadow := ""
		if v.ShadowsVar != nil {
			shadow = fmt.Sprintf("⚠ SHADOWS %s=%q [%s]",
				v.ShadowsVar.Name, fmt.Sprintf("%v", v.ShadowsVar.Value),
				originLabel(v.ShadowsVar.Origin))
		}

		rows = append(rows, []string{name, origin, typeStr, valStr, shadow})
	}

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Padding(0, 1)

	cellStyle := lipgloss.NewStyle().
		Padding(0, 1)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("240"))).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}
			return cellStyle
		}).
		Headers("Name", "Origin", "Type", "Value", "Shadows?").
		Rows(rows...)

	// Indent the table output by 2 spaces
	rendered := t.Render()
	for _, line := range strings.Split(rendered, "\n") {
		fmt.Fprintf(w, "  %s\n", line)
	}
}

// truncateLipglossValue truncates a value string for lipgloss display.
// Unlike the custom renderer, lipgloss handles multiline wrapping, so we
// only need to cap the total length of each line.
func truncateLipglossValue(s string, maxWidth int) string {
	lines := strings.Split(s, "\n")
	var result []string
	for _, line := range lines {
		r := []rune(line)
		if len(r) > maxWidth {
			line = string(r[:maxWidth-1]) + "…"
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}
