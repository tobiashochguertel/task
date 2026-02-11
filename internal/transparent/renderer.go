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
	cReset, cBold, cDim       string
	cRed, cGreen, cYellow     string
	cBlue, cCyan, cWhite      string
	colorOnce                 sync.Once
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
	fmt.Fprintf(w, "\n%s%s══════ Transparent Mode Report ══════%s\n\n", cBold, cCyan, cReset)

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

	fmt.Fprintf(w, "%s%s══════ End Report ══════%s\n", cBold, cCyan, cReset)
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
	fmt.Fprintf(w, "  %s%sVariables:%s\n", cBold, cYellow, cReset)

	// Compute column widths
	maxName := 4 // "Name"
	for _, v := range vars {
		if len(v.Name) > maxName {
			maxName = len(v.Name)
		}
	}

	// Header
	fmt.Fprintf(w, "  %s%-*s  %-14s  %-8s  %-6s  Value%s\n",
		cDim, maxName, "Name", "Origin", "Type", "Ref?", cReset)
	fmt.Fprintf(w, "  %s%s%s\n", cDim, strings.Repeat("─", maxName+14+8+6+10), cReset)

	for _, v := range vars {
		originStr := originLabel(v.Origin)
		typeStr := v.Type
		if typeStr == "" {
			typeStr = "-"
		}

		refStr := "  ·"
		if v.IsRef {
			refStr = fmt.Sprintf("%s ref%s", cRed, cReset)
		}

		valStr := truncate(fmt.Sprintf("%v", v.Value), 60)
		if v.IsDynamic {
			shInfo := ""
			if v.ShCmd != "" {
				shInfo = fmt.Sprintf(" %s(sh: %s)%s", cDim, truncate(v.ShCmd, 40), cReset)
			}
			valStr = fmt.Sprintf("%s(sh)%s %s%s", cBlue, cReset, valStr, shInfo)
			// Feature 2: Warn when dynamic var is empty (likely not evaluated)
			if fmt.Sprintf("%v", v.Value) == "" {
				valStr += fmt.Sprintf(" %s⚠ DYNAMIC — sh: not evaluated (use task run to resolve)%s", cYellow, cReset)
			}
		}

		shadowFlag := ""
		if v.ShadowsVar != nil {
			shadowFlag = fmt.Sprintf(" %s⚠ SHADOWS %s=%q [%s]%s",
				cRed, v.ShadowsVar.Name, truncate(fmt.Sprintf("%v", v.ShadowsVar.Value), 30),
				originLabel(v.ShadowsVar.Origin), cReset)
		}

		fmt.Fprintf(w, "  %-*s  %-14s  %-8s  %-6s  %s%s\n",
			maxName, v.Name, originStr, typeStr, refStr, valStr, shadowFlag)

		if v.ValueID != 0 {
			fmt.Fprintf(w, "  %s%*s  ptr: 0x%x%s\n",
				cDim, maxName, "", v.ValueID, cReset)
		}
		if v.RefName != "" {
			fmt.Fprintf(w, "  %s%*s  → aliases: %s%s\n",
				cDim, maxName, "", v.RefName, cReset)
		}
	}
}

func renderTemplates(w io.Writer, templates []TemplateTrace) {
	fmt.Fprintf(w, "  %s%sTemplate Evaluations:%s\n", cBold, cYellow, cReset)

	for i, t := range templates {
		contextLabel := ""
		if t.Context != "" {
			contextLabel = fmt.Sprintf(" %s(%s)%s", cDim, t.Context, cReset)
		}
		fmt.Fprintf(w, "  %s[%d]%s%s Input:  %s%s%s\n",
			cDim, i+1, cReset, contextLabel, cWhite, t.Input, cReset)
		fmt.Fprintf(w, "       Output: %s%s%s\n",
			cGreen, t.Output, cReset)
		if len(t.VarsUsed) > 0 {
			fmt.Fprintf(w, "       %sVars used: %s%s\n",
				cDim, strings.Join(t.VarsUsed, ", "), cReset)
		}
		if len(t.Steps) > 0 {
			for j, step := range t.Steps {
				fmt.Fprintf(w, "       %s  pipe[%d]: %s(%s) → %q%s\n",
					cDim, j, step.FuncName, step.Args, step.Output, cReset)
			}
		}
		if t.Error != "" {
			fmt.Fprintf(w, "       %s⚠  %s%s\n", cRed, t.Error, cReset)
		}
		for _, tip := range t.Tips {
			fmt.Fprintf(w, "       %s💡 %s%s\n", cCyan, tip, cReset)
		}
	}
}

func renderCmds(w io.Writer, cmds []CmdTrace) {
	fmt.Fprintf(w, "  %s%sCommands:%s\n", cBold, cYellow, cReset)

	for _, c := range cmds {
		fmt.Fprintf(w, "  %s[%d]%s", cDim, c.Index, cReset)
		if c.IterationLabel != "" {
			fmt.Fprintf(w, " %s(%s)%s", cDim, c.IterationLabel, cReset)
		}
		if c.RawCmd == c.ResolvedCmd {
			fmt.Fprintf(w, " %s\n", c.ResolvedCmd)
		} else {
			fmt.Fprintf(w, " %sraw:%s      %s\n", cDim, cReset, c.RawCmd)
			fmt.Fprintf(w, "       %sresolved:%s %s%s%s\n",
				cDim, cReset, cGreen, c.ResolvedCmd, cReset)
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
