package transparent

import (
	"fmt"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/fatih/color"
)

// ANSI color code constants
const (
	ansiReset  = "\033[0m"
	ansiBold   = "\033[1m"
	ansiDim    = "\033[2m"
	ansiRed    = "\033[31m"
	ansiGreen  = "\033[32m"
	ansiYellow = "\033[33m"
	ansiBlue   = "\033[34m"
	ansiCyan   = "\033[36m"
	ansiWhite  = "\033[37m"
)

// Active color codes (resolved once at first render)
var (
	cReset, cBold, cDim   string
	cRed, cGreen, cYellow string
	cBlue, cCyan, cWhite  string
	colorOnce             sync.Once
)

func resolveColors() {
	colorOnce.Do(func() {
		if color.NoColor || os.Getenv("NO_COLOR") != "" {
			return // all vars stay empty strings
		}
		cReset = ansiReset
		cBold = ansiBold
		cDim = ansiDim
		cRed = ansiRed
		cGreen = ansiGreen
		cYellow = ansiYellow
		cBlue = ansiBlue
		cCyan = ansiCyan
		cWhite = ansiWhite
	})
}

// RenderOptions controls what the renderers display.
type RenderOptions struct {
	Verbose bool // When false, hide environment-origin global vars for cleaner output
}

// RenderText writes a human-readable trace report to the given writer.
func RenderText(w io.Writer, report *TraceReport, opts *RenderOptions) {
	if report == nil {
		return
	}
	if opts == nil {
		opts = &RenderOptions{}
	}
	resolveColors()
	headerText := "TRANSPARENT MODE — Variable & Template Diagnostics"
	borderLen := len(headerText) + 4
	fmt.Fprintf(w, "\n%s%s╔%s╗%s\n", cBold, cCyan, strings.Repeat("═", borderLen), cReset)
	fmt.Fprintf(w, "%s%s║  %s  ║%s\n", cBold, cCyan, headerText, cReset)
	fmt.Fprintf(w, "%s%s╚%s╝%s\n\n", cBold, cCyan, strings.Repeat("═", borderLen), cReset)

	// Render global variables section (if any)
	globals := filterGlobals(report.GlobalVars, opts.Verbose)
	if len(globals) > 0 {
		fmt.Fprintf(w, "%s%s── Global Variables%s\n", cBold, cGreen, cReset)
		renderVars(w, globals)
		if !opts.Verbose && len(globals) < len(report.GlobalVars) {
			hidden := len(report.GlobalVars) - len(globals)
			fmt.Fprintf(w, "  %s(%d environment variables hidden — use -v to show)%s\n", cDim, hidden, cReset)
		}
		fmt.Fprintln(w)
	}

	for _, task := range report.Tasks {
		renderTask(w, *task)
	}

	fmt.Fprintf(w, "%s%s╚══ End of Transparent Mode Report ══╝%s\n", cBold, cCyan, cReset)
}

// filterGlobals returns global vars, optionally filtering out noisy vars.
// In non-verbose mode, hides environment-origin vars and internal CLI_* vars.
func filterGlobals(vars []VarTrace, verbose bool) []VarTrace {
	if verbose {
		return vars
	}
	filtered := make([]VarTrace, 0, len(vars))
	for _, v := range vars {
		if v.Origin == OriginEnvironment {
			continue
		}
		if isInternalVar(v.Name) {
			continue
		}
		filtered = append(filtered, v)
	}
	return filtered
}

// isInternalVar returns true for CLI_* and other internal variables that
// clutter the default output. Shown with -v.
func isInternalVar(name string) bool {
	switch name {
	case "CLI_ARGS", "CLI_ARGS_LIST", "CLI_FORCE", "CLI_SILENT",
		"CLI_VERBOSE", "CLI_OFFLINE", "CLI_ASSUME_YES":
		return true
	}
	return false
}

func renderTask(w io.Writer, task TaskTrace) {
	fmt.Fprintf(w, "%s%s── Task: %s%s\n", cBold, cGreen, task.TaskName, cReset)

	if len(task.Vars) > 0 {
		renderVars(w, task.Vars)
	}
	if len(task.Templates) > 0 {
		renderTemplates(w, task.Templates)
	}
	if len(task.Cmds) > 0 {
		renderCmds(w, task.Cmds)
	}
	if len(task.Deps) > 0 {
		renderDeps(w, task.Deps)
	}
	fmt.Fprintln(w)
}

func renderVars(w io.Writer, vars []VarTrace) {
	fmt.Fprintf(w, "  %s%sVariables in scope:%s\n", cBold, cYellow, cReset)

	// Compute column widths dynamically
	colName := 4   // min "Name"
	colValue := 5  // min "Value"
	colOrigin := 6 // min "Origin"
	colType := 4   // min "Type"
	colShadow := 8 // min "Shadows?"

	type rowData struct {
		name, value, origin, typeStr, shadow string
		extraLines                           []string // additional lines below the row (ptr, ref alias)
	}

	rows := make([]rowData, 0, len(vars))
	for _, v := range vars {
		rd := rowData{}
		rd.name = v.Name
		rd.origin = originLabel(v.Origin)

		rd.typeStr = v.Type
		if rd.typeStr == "" {
			rd.typeStr = "-"
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
		rd.value = valStr

		// Shadow
		if v.ShadowsVar != nil {
			rd.shadow = fmt.Sprintf("⚠ SHADOWS %s=%q [%s]",
				v.ShadowsVar.Name, fmt.Sprintf("%v", v.ShadowsVar.Value),
				originLabel(v.ShadowsVar.Origin))
		}

		// Extra lines
		if v.ValueID != 0 {
			rd.extraLines = append(rd.extraLines, fmt.Sprintf("ptr: 0x%x", v.ValueID))
		}
		if v.RefName != "" {
			rd.extraLines = append(rd.extraLines, fmt.Sprintf("→ aliases: %s", v.RefName))
		}

		// Update column widths
		if len(rd.name) > colName {
			colName = len(rd.name)
		}
		if len(rd.value) > colValue {
			colValue = len(rd.value)
		}
		if len(rd.origin) > colOrigin {
			colOrigin = len(rd.origin)
		}
		if len(rd.typeStr) > colType {
			colType = len(rd.typeStr)
		}
		if len(rd.shadow) > colShadow {
			colShadow = len(rd.shadow)
		}

		rows = append(rows, rd)
	}

	// Table rendering
	hLine := func(left, mid, right, fill string) {
		fmt.Fprintf(w, "  %s%s%s%s%s%s%s%s%s%s%s%s\n",
			cDim,
			left, strings.Repeat(fill, colName+2),
			mid, strings.Repeat(fill, colOrigin+2),
			mid, strings.Repeat(fill, colType+2),
			mid, strings.Repeat(fill, colValue+2),
			mid, strings.Repeat(fill, colShadow+2),
			right+cReset)
	}

	row := func(name, origin, typeStr, value, shadow string) {
		fmt.Fprintf(w, "  %s│%s %-*s %s│%s %-*s %s│%s %-*s %s│%s %-*s %s│%s %-*s %s│%s\n",
			cDim, cReset, colName, name,
			cDim, cReset, colOrigin, origin,
			cDim, cReset, colType, typeStr,
			cDim, cReset, colValue, value,
			cDim, cReset, colShadow, shadow,
			cDim, cReset)
	}

	hLine("┌", "┬", "┐", "─")
	row("Name", "Origin", "Type", "Value", "Shadows?")
	hLine("├", "┼", "┤", "─")

	for _, rd := range rows {
		shadowStr := rd.shadow
		if shadowStr != "" {
			shadowStr = fmt.Sprintf("%s%s%s", cYellow, rd.shadow, cReset)
		}
		valueStr := rd.value
		if rd.value != "" && strings.HasPrefix(rd.value, "(sh)") {
			valueStr = fmt.Sprintf("%s%s%s", cBlue, rd.value, cReset)
		}
		row(rd.name, rd.origin, rd.typeStr, valueStr, shadowStr)
		for _, extra := range rd.extraLines {
			row("", "", "", fmt.Sprintf("%s%s%s", cDim, extra, cReset), "")
		}
	}

	hLine("└", "┴", "┘", "─")
}

// --- Box-drawing helpers ---

func renderBoxStart(w io.Writer, label string) {
	fmt.Fprintf(w, "  %s┌─ %s:%s\n", cDim, label, cReset)
}

func renderBoxLine(w io.Writer, line string) {
	fmt.Fprintf(w, "  %s│%s %s\n", cDim, cReset, line)
}

func renderBoxEnd(w io.Writer) {
	fmt.Fprintf(w, "  %s└─%s\n", cDim, cReset)
}

func renderBoxContent(w io.Writer, label string, content string) {
	renderBoxStart(w, label)
	for _, line := range strings.Split(content, "\n") {
		renderBoxLine(w, line)
	}
	renderBoxEnd(w)
}

// highlightErrors wraps known Go template error patterns with red ANSI codes.
func highlightErrors(s string) string {
	patterns := []string{
		"%!s(MISSING)", "%!d(MISSING)", "%!v(MISSING)", "%!f(MISSING)",
		"%!q(MISSING)", "%!t(MISSING)", "%!x(MISSING)", "%!b(MISSING)",
		"%!e(MISSING)", "%!g(MISSING)", "%!c(MISSING)",
	}
	for _, p := range patterns {
		if strings.Contains(s, p) {
			s = strings.ReplaceAll(s, p, cRed+cBold+p+cReset)
		}
	}
	return s
}

func renderTemplates(w io.Writer, templates []TemplateTrace) {
	for _, t := range templates {
		contextLabel := t.Context
		if contextLabel == "" {
			contextLabel = "expression"
		}
		fmt.Fprintf(w, "  %s%sTemplate Evaluation — %s:%s\n", cBold, cYellow, contextLabel, cReset)

		// Input box
		renderBoxContent(w, "Input", t.Input)

		// Pipe steps box (if any)
		if len(t.Steps) > 0 {
			renderBoxStart(w, "Pipe Steps")
			for j, step := range t.Steps {
				argsStr := strings.Join(step.Args, ", ")
				if len(step.ArgsValues) > 0 {
					argsStr = strings.Join(step.ArgsValues, ", ")
				}
				renderBoxLine(w, fmt.Sprintf("Step %d: %s(%s) → %s%s%s",
					j+1, step.FuncName, argsStr, cGreen, step.Output, cReset))
			}
			renderBoxEnd(w)
		}

		// Output box
		renderBoxContent(w, "Output", highlightErrors(t.Output))

		// Vars used box
		if len(t.VarsUsed) > 0 {
			renderBoxContent(w, "Vars used", strings.Join(t.VarsUsed, ", "))
		}

		// Error
		if t.Error != "" {
			fmt.Fprintf(w, "  %s⚠ %s%s\n", cRed, t.Error, cReset)
		}

		// Notes / Tips
		for _, tip := range t.Tips {
			fmt.Fprintf(w, "  %sℹ Note: %s%s\n", cCyan, tip, cReset)
		}
	}
}

func renderCmds(w io.Writer, cmds []CmdTrace) {
	for _, c := range cmds {
		header := fmt.Sprintf("cmds[%d]", c.Index)
		if c.IterationLabel != "" {
			header = fmt.Sprintf("cmds[%d] (%s)", c.Index, c.IterationLabel)
		}
		fmt.Fprintf(w, "  %s%sCommands — %s:%s\n", cBold, cYellow, header, cReset)

		if c.RawCmd == c.ResolvedCmd {
			renderBoxContent(w, "Command", c.ResolvedCmd)
		} else {
			renderBoxContent(w, "Raw", c.RawCmd)
			renderBoxContent(w, "Resolved", highlightErrors(c.ResolvedCmd))
		}
	}
}

func renderDeps(w io.Writer, deps []string) {
	fmt.Fprintf(w, "  %s%sDependencies:%s %s\n",
		cBold, cYellow, cReset, strings.Join(deps, ", "))
}

func originLabel(o VarOrigin) string {
	switch o {
	case OriginEnvironment:
		return "env"
	case OriginSpecial:
		return "special"
	case OriginTaskfileEnv:
		return "taskfile-env"
	case OriginTaskfileVars:
		return "taskfile-vars"
	case OriginIncludeVars:
		return "include-vars"
	case OriginIncludedTaskfileVars:
		return "included-tf"
	case OriginCallVars:
		return "call-vars"
	case OriginTaskVars:
		return "task-vars"
	case OriginForLoop:
		return "for-loop"
	case OriginDotenv:
		return "dotenv"
	default:
		return "unknown"
	}
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
