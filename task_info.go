package task

import (
	"slices"

	"github.com/go-task/task/v3/taskfile/ast"
)

// taskToMap converts a Task AST node into a template-friendly map.
// The resulting map exposes task properties that can be used in Go templates,
// including iteration with {{range}}.
func taskToMap(t *ast.Task) map[string]any {
	if t == nil {
		return map[string]any{}
	}

	m := map[string]any{
		"Name":        t.Task,
		"Desc":        t.Desc,
		"Summary":     t.Summary,
		"Aliases":     safeStringSlice(t.Aliases),
		"Dir":         t.Dir,
		"Method":      t.Method,
		"Silent":      t.IsSilent(),
		"Interactive": t.Interactive,
		"Internal":    t.Internal,
		"IgnoreError": t.IgnoreError,
		"Run":         t.Run,
		"Watch":       t.Watch,
		"Namespace":   t.Namespace,
		"Label":       t.Label,
		"Prefix":      t.Prefix,
		"If":          t.If,
		"Failfast":    t.Failfast,
		"Sources":     globsToStrings(t.Sources),
		"Generates":   globsToStrings(t.Generates),
		"Status":      safeStringSlice(t.Status),
		"Dotenv":      safeStringSlice(t.Dotenv),
		"Set":         safeStringSlice(t.Set),
		"Shopt":       safeStringSlice(t.Shopt),
	}

	// Commands
	cmds := make([]map[string]any, 0, len(t.Cmds))
	for _, cmd := range t.Cmds {
		if cmd == nil {
			continue
		}
		cmds = append(cmds, cmdToMap(cmd))
	}
	m["Cmds"] = cmds

	// Dependencies
	deps := make([]map[string]any, 0, len(t.Deps))
	for _, dep := range t.Deps {
		if dep == nil {
			continue
		}
		deps = append(deps, depToMap(dep))
	}
	m["Deps"] = deps

	// Vars as a simple name→value map
	m["Vars"] = varsToMap(t.Vars)

	// Env as a simple name→value map
	m["Env"] = varsToMap(t.Env)

	// Location
	if t.Location != nil {
		m["Location"] = map[string]any{
			"Taskfile": t.Location.Taskfile,
			"Line":     t.Location.Line,
			"Column":   t.Location.Column,
		}
	}

	return m
}

// taskfileToMap converts a Taskfile AST into a template-friendly map.
// The resulting map exposes Taskfile properties and all task names/info
// that can be iterated with {{range}}.
func taskfileToMap(tf *ast.Taskfile) map[string]any {
	if tf == nil {
		return map[string]any{}
	}

	m := map[string]any{
		"Location": tf.Location,
		"Method":   tf.Method,
		"Silent":   tf.Silent,
		"Run":      tf.Run,
		"Dotenv":   safeStringSlice(tf.Dotenv),
		"Set":      safeStringSlice(tf.Set),
		"Shopt":    safeStringSlice(tf.Shopt),
	}

	// Version
	if tf.Version != nil {
		m["Version"] = tf.Version.String()
	} else {
		m["Version"] = ""
	}

	// Vars as name→value map
	m["Vars"] = varsToMap(tf.Vars)

	// Env as name→value map
	m["Env"] = varsToMap(tf.Env)

	// Tasks: map of task name → task info map
	taskMap := make(map[string]any)
	taskNames := make([]string, 0)
	if tf.Tasks != nil {
		for name, task := range tf.Tasks.All(nil) {
			taskMap[name] = taskToMap(task)
			taskNames = append(taskNames, name)
		}
	}
	m["Tasks"] = taskMap
	m["TaskNames"] = taskNames

	return m
}

// cmdToMap converts a Cmd to a template-friendly map.
func cmdToMap(cmd *ast.Cmd) map[string]any {
	if cmd == nil {
		return map[string]any{}
	}
	return map[string]any{
		"Cmd":         cmd.Cmd,
		"Task":        cmd.Task,
		"If":          cmd.If,
		"Silent":      cmd.Silent,
		"IgnoreError": cmd.IgnoreError,
		"Defer":       cmd.Defer,
	}
}

// depToMap converts a Dep to a template-friendly map.
func depToMap(dep *ast.Dep) map[string]any {
	if dep == nil {
		return map[string]any{}
	}
	return map[string]any{
		"Task":   dep.Task,
		"Silent": dep.Silent,
	}
}

// varsToMap converts an ast.Vars ordered map to a plain map[string]any.
func varsToMap(vars *ast.Vars) map[string]any {
	if vars == nil {
		return map[string]any{}
	}
	m := make(map[string]any)
	for k, v := range vars.All() {
		if v.Value != nil {
			m[k] = v.Value
		} else if v.Sh != nil {
			m[k] = "(sh: " + *v.Sh + ")"
		}
	}
	return m
}

// globsToStrings converts a slice of Glob pointers to a string slice.
func globsToStrings(globs []*ast.Glob) []string {
	result := make([]string, 0, len(globs))
	for _, g := range globs {
		if g != nil {
			if g.Negate {
				result = append(result, "!"+g.Glob)
			} else {
				result = append(result, g.Glob)
			}
		}
	}
	return result
}

// safeStringSlice returns the input slice or an empty slice if nil,
// ensuring Go templates never get a nil value for range operations.
func safeStringSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return slices.Clone(s)
}
