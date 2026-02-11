package transparent

import (
	"fmt"
	"io"
	"strings"
)

// ANSI color codes for terminal output
const (
	colorReset  = "\033[0m"
	colorBold   = "\033[1m"
	colorDim    = "\033[2m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// RenderText writes a human-readable trace report to the given writer.
func RenderText(w io.Writer, report *TraceReport) {
	if report == nil {
		return
	}
	fmt.Fprintf(w, "\n%s%s══════ Transparent Mode Report ══════%s\n\n", colorBold, colorCyan, colorReset)

	for _, task := range report.Tasks {
		renderTask(w, *task)
	}

	fmt.Fprintf(w, "%s%s══════ End Report ══════%s\n", colorBold, colorCyan, colorReset)
}

func renderTask(w io.Writer, task TaskTrace) {
	fmt.Fprintf(w, "%s%s── Task: %s%s\n", colorBold, colorGreen, task.TaskName, colorReset)

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
	fmt.Fprintf(w, "  %s%sVariables:%s\n", colorBold, colorYellow, colorReset)

	// Compute column widths
	maxName := 4 // "Name"
	for _, v := range vars {
		if len(v.Name) > maxName {
			maxName = len(v.Name)
		}
	}

	// Header
	fmt.Fprintf(w, "  %s%-*s  %-14s  %-8s  %-6s  Value%s\n",
		colorDim, maxName, "Name", "Origin", "Type", "Ref?", colorReset)
	fmt.Fprintf(w, "  %s%s%s\n", colorDim, strings.Repeat("─", maxName+14+8+6+10), colorReset)

	for _, v := range vars {
		originStr := originLabel(v.Origin)
		typeStr := v.Type
		if typeStr == "" {
			typeStr = "-"
		}

		refStr := "  ·"
		if v.IsRef {
			refStr = fmt.Sprintf("%s ref%s", colorRed, colorReset)
		}

		valStr := truncate(fmt.Sprintf("%v", v.Value), 60)
		if v.IsDynamic {
			valStr = fmt.Sprintf("%s(sh)%s %s", colorBlue, colorReset, valStr)
		}

		shadowFlag := ""
		if v.ShadowsVar != nil {
			shadowFlag = fmt.Sprintf(" %s⚠ shadows %s%s", colorRed, v.ShadowsVar.Name, colorReset)
		}

		fmt.Fprintf(w, "  %-*s  %-14s  %-8s  %-6s  %s%s\n",
			maxName, v.Name, originStr, typeStr, refStr, valStr, shadowFlag)

		if v.ValueID != 0 {
			fmt.Fprintf(w, "  %s%*s  ptr: 0x%x%s\n",
				colorDim, maxName, "", v.ValueID, colorReset)
		}
		if v.RefName != "" {
			fmt.Fprintf(w, "  %s%*s  → aliases: %s%s\n",
				colorDim, maxName, "", v.RefName, colorReset)
		}
	}
}

func renderTemplates(w io.Writer, templates []TemplateTrace) {
	fmt.Fprintf(w, "  %s%sTemplate Evaluations:%s\n", colorBold, colorYellow, colorReset)

	for i, t := range templates {
		fmt.Fprintf(w, "  %s[%d]%s Input:  %s%s%s\n",
			colorDim, i+1, colorReset, colorWhite, t.Input, colorReset)
		fmt.Fprintf(w, "       Output: %s%s%s\n",
			colorGreen, t.Output, colorReset)
		if len(t.VarsUsed) > 0 {
			fmt.Fprintf(w, "       %sVars used: %s%s\n",
				colorDim, strings.Join(t.VarsUsed, ", "), colorReset)
		}
		if len(t.Steps) > 0 {
			for j, step := range t.Steps {
				fmt.Fprintf(w, "       %s  pipe[%d]: %s(%s) → %q%s\n",
					colorDim, j, step.FuncName, step.Args, step.Output, colorReset)
			}
		}
	}
}

func renderCmds(w io.Writer, cmds []CmdTrace) {
	fmt.Fprintf(w, "  %s%sCommands:%s\n", colorBold, colorYellow, colorReset)

	for _, c := range cmds {
		fmt.Fprintf(w, "  %s[%d]%s", colorDim, c.Index, colorReset)
		if c.RawCmd == c.ResolvedCmd {
			fmt.Fprintf(w, " %s\n", c.ResolvedCmd)
		} else {
			fmt.Fprintf(w, " %sraw:%s      %s\n", colorDim, colorReset, c.RawCmd)
			fmt.Fprintf(w, "       %sresolved:%s %s%s%s\n",
				colorDim, colorReset, colorGreen, c.ResolvedCmd, colorReset)
		}
	}
}

func renderDeps(w io.Writer, deps []string) {
	fmt.Fprintf(w, "  %s%sDependencies:%s %s\n",
		colorBold, colorYellow, colorReset, strings.Join(deps, ", "))
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
