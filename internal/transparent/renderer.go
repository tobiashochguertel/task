package transparent

import (
	"fmt"
	"io"
	"os"
	"regexp"
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
)

// Active color codes (resolved once at first render)
var (
	cReset, cBold, cDim   string
	cRed, cGreen, cYellow string
	cBlue, cCyan          string
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
	})
}

// RenderOptions controls what the renderers display.
type RenderOptions struct {
	Verbose         bool // When false, hide environment-origin global vars for cleaner output
	ShowWhitespaces bool // When true, replace spaces with · and tabs with → in values
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

	// ShowWhitespaces: apply transformation to report values
	if opts.ShowWhitespaces {
		report = applyWhitespaceVisibility(report)
	}

	headerText := "TRANSPARENT MODE \u2014 Variable & Template Diagnostics"
	borderLen := len(headerText) + 4
	fmt.Fprintf(w, "\n%s%s╔%s╗%s\n", cBold, cCyan, strings.Repeat("═", borderLen), cReset)
	fmt.Fprintf(w, "%s%s║  %s  ║%s\n", cBold, cCyan, headerText, cReset)
	fmt.Fprintf(w, "%s%s\u255a%s\u255d%s\n", cBold, cCyan, strings.Repeat("\u2550", borderLen), cReset)

	if opts.ShowWhitespaces {
		fmt.Fprintf(w, "%sLegend: \u00b7 = space, \u2192 = tab%s\n", cDim, cReset)
	}
	fmt.Fprintln(w)

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
		"CLI_VERBOSE", "CLI_OFFLINE", "CLI_ASSUME_YES",
		"TASK_INFO", "TASKFILE_INFO":
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
	if len(task.SubtaskCalls) > 0 {
		renderSubtaskCalls(w, task.SubtaskCalls)
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
		// Include extra lines (ptr, ref alias) in value column width
		for _, extra := range rd.extraLines {
			if len(extra) > colValue {
				colValue = len(extra)
			}
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
		// Adjust padding widths to account for invisible ANSI escape sequences
		padName := colName + (len(name) - visibleLen(name))
		padOrigin := colOrigin + (len(origin) - visibleLen(origin))
		padType := colType + (len(typeStr) - visibleLen(typeStr))
		padValue := colValue + (len(value) - visibleLen(value))
		padShadow := colShadow + (len(shadow) - visibleLen(shadow))
		fmt.Fprintf(w, "  %s│%s %-*s %s│%s %-*s %s│%s %-*s %s│%s %-*s %s│%s %-*s %s│%s\n",
			cDim, cReset, padName, name,
			cDim, cReset, padOrigin, origin,
			cDim, cReset, padType, typeStr,
			cDim, cReset, padValue, value,
			cDim, cReset, padShadow, shadow,
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
	lines := strings.Split(content, "\n")
	// Remove trailing empty lines to avoid blank line before └─
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	for _, line := range lines {
		renderBoxLine(w, line)
	}
	renderBoxEnd(w)
}

// ansiRegex matches ANSI escape sequences for stripping when computing visible width.
var ansiRegex = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]`)

// visibleLen returns the visible length of a string, excluding ANSI escape sequences.
func visibleLen(s string) int {
	return len(ansiRegex.ReplaceAllString(s, ""))
}

// stepFieldPad is the indentation for continuation lines in step fields,
// matching the column where content starts after the single-char label.
const stepFieldPad = "        " // 8 spaces: 2 indent + 1 label + 5 spacing

// renderStepField renders a step field (I/O/E) with proper multiline alignment.
// label is a single character. content may contain newlines. colorStart/colorEnd
// wrap each content line in ANSI color codes (pass empty strings for no color).
func renderStepField(w io.Writer, label string, content string, colorStart, colorEnd string) {
	lines := strings.Split(content, "\n")
	// Remove trailing empty lines
	for len(lines) > 0 && strings.TrimSpace(lines[len(lines)-1]) == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) == 0 {
		return
	}
	// First line with label: "  I     content"
	renderBoxLine(w, fmt.Sprintf("  %s     %s%s%s", label, colorStart, lines[0], colorEnd))
	// Continuation lines aligned to same column
	for _, line := range lines[1:] {
		renderBoxLine(w, fmt.Sprintf("%s%s%s%s", stepFieldPad, colorStart, line, colorEnd))
	}
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

		// Action-grouped step-by-step evaluation (if available)
		if len(t.EvalActions) > 0 {
			totalActions := len(t.EvalActions)
			renderBoxStart(w, "Evaluation Steps")
			for _, ea := range t.EvalActions {
				// Action header
				renderBoxLine(w, "")
				renderBoxLine(w, fmt.Sprintf("%s── Action %d of %d — line %d%s",
					cDim, ea.ActionIndex+1, totalActions, ea.SourceLine, cReset))
				// Source line
				renderStepField(w, "S", ea.Source, cCyan, cReset)
				renderBoxLine(w, "")

				// Steps within this action
				for _, ds := range ea.Steps {
					opColor := cCyan
					if ds.Operation == "Apply a Function" {
						opColor = cYellow
					}
					renderBoxLine(w, fmt.Sprintf("%sStep %d:%s %s%s%s — %s%s%s",
						cBold, ds.StepNum, cReset,
						opColor, ds.Operation, cReset,
						cDim, ds.Target, cReset))
					if ds.Input != "" {
						renderStepField(w, "I", ds.Input, "", "")
					}
					if ds.Output != "" {
						renderStepField(w, "O", ds.Output, cGreen, cReset)
					}
				}

				// Result line
				renderBoxLine(w, "")
				renderStepField(w, "R", ea.Result, cGreen, cReset)
			}
			renderBoxLine(w, "")
			renderBoxEnd(w)
		} else if len(t.Steps) > 0 {
			// Fallback: show pipe steps if detailed steps not available
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

		// Structured diagnostics
		for _, d := range t.Diagnostics {
			renderDiagnostic(w, d)
		}

		// Notes / Tips (legacy hints, shown only if no structured diagnostics)
		if len(t.Diagnostics) == 0 {
			for _, tip := range t.Tips {
				fmt.Fprintf(w, "  %sℹ Note: %s%s\n", cCyan, tip, cReset)
			}
		}
	}
}

// renderDiagnostic renders a single FuncDiagnostic with structured formatting.
func renderDiagnostic(w io.Writer, d FuncDiagnostic) {
	// Diagnostic header with type icon
	icon := "⚠"
	typeLabel := "Output Anomaly"
	headerColor := cYellow
	if d.DiagType == "exec_error" {
		icon = "✖"
		typeLabel = "Execution Error"
		headerColor = cRed
	}

	fmt.Fprintf(w, "  %s%s %s — %s (Step %d)%s\n",
		headerColor, icon, typeLabel, d.FuncName, d.StepNum, cReset)

	// Expression context
	if d.Expression != "" {
		fmt.Fprintf(w, "      %sExpression%s  %s{{%s}}%s\n",
			cDim, cReset, cCyan, d.Expression, cReset)
	}

	// Error message
	if d.ErrorMsg != "" {
		fmt.Fprintf(w, "      %sError%s       %s%s%s\n",
			cDim, cReset, cRed, d.ErrorMsg, cReset)
	}

	// Output produced
	if d.Output != "" {
		fmt.Fprintf(w, "      %sOutput%s      %s%s%s\n",
			cDim, cReset, cRed, d.Output, cReset)
	}

	// Signature & Example as compact reference
	if d.Signature != "" {
		fmt.Fprintf(w, "      %s┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈%s\n", cDim, cReset)
		fmt.Fprintf(w, "      %sSignature%s   %s\n", cDim, cReset, d.Signature)
		if d.Example != "" {
			fmt.Fprintf(w, "      %sExample%s     %s%s%s\n", cDim, cReset, cCyan, d.Example, cReset)
		}
		fmt.Fprintf(w, "      %s┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈┈%s\n", cDim, cReset)
	}

	// Call and parameter mapping
	if d.Call != "" {
		fmt.Fprintf(w, "      %sCall%s        %s%s%s\n",
			cDim, cReset, cBold, d.Call, cReset)
	}

	if len(d.Params) > 0 {
		fmt.Fprintf(w, "      %sParams%s\n", cDim, cReset)
		varIdx := 0
		for _, p := range d.Params {
			label := p.Name
			if p.Variadic {
				label = fmt.Sprintf("%s[%d]", p.Name, varIdx)
				varIdx++
			} else {
				varIdx = 0
			}
			if p.Missing {
				fmt.Fprintf(w, "        %s%-12s%s %s⚠ MISSING%s  %s(%s)%s\n",
					cDim, label, cReset, cRed, cReset, cDim, p.Type, cReset)
			} else {
				fmt.Fprintf(w, "        %s%-12s%s %s  %s(%s)%s\n",
					cDim, label, cReset, p.Value, cDim, p.Type, cReset)
			}
		}
	}

	fmt.Fprintln(w)
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

func renderSubtaskCalls(w io.Writer, calls []SubtaskCall) {
	fmt.Fprintf(w, "  %s%sSubtask calls:%s\n", cBold, cYellow, cReset)
	for _, sc := range calls {
		fmt.Fprintf(w, "    %scmds[%d]%s → %s%s%s\n",
			cDim, sc.CmdIndex, cReset,
			cCyan, sc.TaskName, cReset)
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

// makeWhitespaceVisible replaces spaces with · and tabs with → to make
// whitespace visible in the output.
func makeWhitespaceVisible(s string) string {
	// Strip ANSI escape sequences and replace with visible marker
	s = stripANSI(s)
	// Order matters: replace \r\n as a unit before individual \r and \n
	s = strings.ReplaceAll(s, "\r\n", "\u2190\u21b5\n")
	s = strings.ReplaceAll(s, "\n", "\u21b5\n")
	s = strings.ReplaceAll(s, "\r", "\u2190")
	s = strings.ReplaceAll(s, " ", "\u00b7")
	s = strings.ReplaceAll(s, "\t", "\u2192")
	return s
}

// ansiPattern matches ANSI escape sequences (CSI sequences and OSC sequences).
var ansiPattern = regexp.MustCompile(`\x1b\[[0-9;]*[a-zA-Z]|\x1b\][^\x1b]*\x1b\\`)

// stripANSI replaces ANSI escape sequences with a visible [ESC] marker.
func stripANSI(s string) string {
	return ansiPattern.ReplaceAllString(s, "[ESC]")
}

// applyWhitespaceVisibility returns a copy of the report with whitespace
// made visible in all value fields.
func applyWhitespaceVisibility(report *TraceReport) *TraceReport {
	copy := *report

	copy.GlobalVars = applyWSToVars(report.GlobalVars)

	copy.Tasks = make([]*TaskTrace, len(report.Tasks))
	for i, t := range report.Tasks {
		tc := *t
		tc.Vars = applyWSToVars(t.Vars)
		tc.Templates = applyWSToTemplates(t.Templates)
		tc.Cmds = applyWSToCmds(t.Cmds)
		copy.Tasks[i] = &tc
	}
	return &copy
}

func applyWSToVars(vars []VarTrace) []VarTrace {
	out := make([]VarTrace, len(vars))
	for i, v := range vars {
		vc := v
		if s, ok := v.Value.(string); ok {
			vc.Value = makeWhitespaceVisible(s)
		}
		if v.ShCmd != "" {
			vc.ShCmd = makeWhitespaceVisible(v.ShCmd)
		}
		if v.ShadowsVar != nil {
			sc := *v.ShadowsVar
			if s, ok := sc.Value.(string); ok {
				sc.Value = makeWhitespaceVisible(s)
			}
			vc.ShadowsVar = &sc
		}
		out[i] = vc
	}
	return out
}

func applyWSToTemplates(templates []TemplateTrace) []TemplateTrace {
	out := make([]TemplateTrace, len(templates))
	for i, t := range templates {
		tc := t
		tc.Output = makeWhitespaceVisible(t.Output)
		// Apply whitespace visibility to eval actions
		if len(tc.EvalActions) > 0 {
			actions := make([]EvalAction, len(tc.EvalActions))
			for j, ea := range tc.EvalActions {
				ac := ea
				ac.Source = makeWhitespaceVisible(ea.Source)
				ac.Result = makeWhitespaceVisible(ea.Result)
				steps := make([]TemplateStep, len(ea.Steps))
				for k, ds := range ea.Steps {
					sc := ds
					sc.Input = makeWhitespaceVisible(ds.Input)
					sc.Output = makeWhitespaceVisible(ds.Output)
					steps[k] = sc
				}
				ac.Steps = steps
				actions[j] = ac
			}
			tc.EvalActions = actions
		}
		// Apply whitespace visibility to diagnostics
		if len(tc.Diagnostics) > 0 {
			diags := make([]FuncDiagnostic, len(tc.Diagnostics))
			for j, d := range tc.Diagnostics {
				dc := d
				dc.Expression = makeWhitespaceVisible(d.Expression)
				dc.Call = makeWhitespaceVisible(d.Call)
				dc.Output = makeWhitespaceVisible(d.Output)
				dc.ErrorMsg = makeWhitespaceVisible(d.ErrorMsg)
				if len(d.Params) > 0 {
					params := make([]ParamMapping, len(d.Params))
					for k, p := range d.Params {
						pc := p
						pc.Value = makeWhitespaceVisible(p.Value)
						params[k] = pc
					}
					dc.Params = params
				}
				diags[j] = dc
			}
			tc.Diagnostics = diags
		}
		out[i] = tc
	}
	return out
}

func applyWSToCmds(cmds []CmdTrace) []CmdTrace {
	out := make([]CmdTrace, len(cmds))
	for i, c := range cmds {
		cc := c
		cc.ResolvedCmd = makeWhitespaceVisible(c.ResolvedCmd)
		out[i] = cc
	}
	return out
}
