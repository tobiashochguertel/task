package transparent

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// lipgloss color palette
var (
	lgPurple    = lipgloss.Color("99")
	lgGray      = lipgloss.Color("245")
	lgLightGray = lipgloss.Color("241")
	lgYellow    = lipgloss.Color("220")
	lgBlue      = lipgloss.Color("75")
	lgGreen     = lipgloss.Color("78")
	lgCyan      = lipgloss.Color("80")
	lgDimGray   = lipgloss.Color("240")
)

// renderVarsLipgloss renders the variable table using the charmbracelet/lipgloss
// table package with advanced coloring and styling. Multiline values (like
// pretty-printed JSON maps) get a compact summary in the table with the full
// content rendered in a detail box below.
func renderVarsLipgloss(w io.Writer, vars []VarTrace, opts *RenderOptions) {
	fmt.Fprintf(w, "  Variables in scope:\n")

	// Track per-row metadata for styling decisions
	type rowMeta struct {
		isDynamic   bool
		hasShadow   bool
		isMultiline bool
	}
	metas := make([]rowMeta, 0, len(vars))

	// Collect multiline details to render below the table
	type multilineDetail struct {
		name  string
		value string
	}
	var details []multilineDetail

	rows := make([][]string, 0, len(vars))
	for _, v := range vars {
		name := v.Name
		origin := originLabel(v.Origin)

		typeStr := v.Type
		if typeStr == "" {
			typeStr = "-"
		}

		// Value — use formatVarValue for pretty-printed complex types
		valStr := formatVarValue(v.Value)
		meta := rowMeta{}
		if v.IsDynamic {
			meta.isDynamic = true
			shInfo := ""
			if v.ShCmd != "" {
				shInfo = fmt.Sprintf(" (sh: %s)", v.ShCmd)
			}
			rawStr := formatVarValue(v.Value)
			valStr = fmt.Sprintf("(sh) %s%s", rawStr, shInfo)
			if rawStr == "" {
				valStr += " ⚠ DYNAMIC — not evaluated"
			}
		}
		if v.IsRef {
			valStr = fmt.Sprintf("(ref) %s", valStr)
		}

		// Extra info appended to value
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

		// For multiline values, show compact summary in table, full below
		displayVal := valStr
		if strings.Contains(valStr, "\n") {
			meta.isMultiline = true
			lineCount := strings.Count(valStr, "\n") + 1
			firstLine, _ := splitMultiline(valStr)
			if len([]rune(firstLine)) > 60 {
				firstLine = string([]rune(firstLine)[:57]) + "..."
			}
			displayVal = fmt.Sprintf("%s (%d lines — see below)", firstLine, lineCount)
			details = append(details, multilineDetail{name: name, value: valStr})
		}

		// Shadow
		shadow := ""
		if v.ShadowsVar != nil {
			meta.hasShadow = true
			shadow = fmt.Sprintf("⚠ SHADOWS %s=%q [%s]",
				v.ShadowsVar.Name, fmt.Sprintf("%v", v.ShadowsVar.Value),
				originLabel(v.ShadowsVar.Origin))
		}

		rows = append(rows, []string{name, origin, typeStr, displayVal, shadow})
		metas = append(metas, meta)
	}

	// Styles
	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lgPurple).
		Padding(0, 1)

	t := table.New().
		Border(lipgloss.RoundedBorder()).
		BorderStyle(lipgloss.NewStyle().Foreground(lgDimGray)).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return headerStyle
			}

			base := lipgloss.NewStyle().Padding(0, 1)

			// Alternating row foreground color
			if row%2 == 0 {
				base = base.Foreground(lgGray)
			} else {
				base = base.Foreground(lgLightGray)
			}

			// Per-column styling
			idx := row
			if idx >= 0 && idx < len(metas) {
				switch col {
				case 0: // Name — bold cyan
					base = base.Foreground(lgCyan).Bold(true)
				case 1: // Origin — green
					base = base.Foreground(lgGreen)
				case 3: // Value — blue for dynamic, dim for multiline ref
					if metas[idx].isDynamic {
						base = base.Foreground(lgBlue)
					} else if metas[idx].isMultiline {
						base = base.Foreground(lgDimGray)
					}
				case 4: // Shadows — yellow warning
					if metas[idx].hasShadow {
						base = base.Foreground(lgYellow).Bold(true)
					}
				}
			}

			return base
		}).
		Headers("Name", "Origin", "Type", "Value", "Shadows?").
		Rows(rows...)

	// Indent the table output by 2 spaces
	rendered := t.Render()
	for _, line := range strings.Split(rendered, "\n") {
		fmt.Fprintf(w, "  %s\n", line)
	}

	// Render full values for multiline variables below the table
	for _, d := range details {
		fmt.Fprintln(w)
		showWS := opts != nil && opts.ShowWhitespaces
		renderBoxContent(w, fmt.Sprintf("Value of %s", d.name), syntaxHighlight(d.value, showWS))
	}
}
