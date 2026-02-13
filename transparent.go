package task

import (
	"context"
	"fmt"
	"os"

	"github.com/go-task/task/v3/internal/logger"
	"github.com/go-task/task/v3/internal/transparent"
)

// RunTransparent compiles all requested tasks with tracing enabled
// and renders a diagnostic report instead of executing commands.
// It recursively traces subtask calls (e.g. `- task: clean`) and
// dependencies so the report covers the full execution tree.
func (e *Executor) RunTransparent(ctx context.Context, calls ...*Call) error {
	// Ensure tracer exists
	if e.Compiler.Tracer == nil {
		e.Compiler.Tracer = transparent.NewTracer()
	}

	visited := make(map[string]bool)
	for _, call := range calls {
		if err := e.compileTaskRecursive(call, visited); err != nil {
			return err
		}
	}

	// Post-process: separate global vars from task-specific vars
	e.Compiler.Tracer.SeparateGlobalVars()

	// Render the report
	report := e.Compiler.Tracer.Report()
	opts := &transparent.RenderOptions{
		Verbose:         e.Verbose,
		ShowWhitespaces: e.ShowWhitespaces,
	}
	if e.TransparentJSON {
		return transparent.RenderJSON(os.Stderr, report, opts)
	}
	transparent.RenderText(os.Stderr, report, opts)
	return nil
}

// compileTaskRecursive compiles a task and recursively compiles any subtask
// calls and dependencies found in its commands. The visited map prevents
// infinite loops from cyclic task references.
func (e *Executor) compileTaskRecursive(call *Call, visited map[string]bool) error {
	if visited[call.Task] {
		return nil
	}
	visited[call.Task] = true

	e.Compiler.Tracer.SetCurrentTask(call.Task)

	// Compile the task (resolves variables and templates) without executing
	compiled, err := e.compiledTask(call, true)
	if err != nil {
		return fmt.Errorf("transparent: error compiling task %q: %w", call.Task, err)
	}

	// Record dependencies and recursively compile them
	if compiled.Deps != nil {
		for _, dep := range compiled.Deps {
			if dep != nil {
				e.Compiler.Tracer.RecordDep(call.Task, dep.Task)
				depCall := &Call{Task: dep.Task, Vars: dep.Vars, Indirect: true}
				if err := e.compileTaskRecursive(depCall, visited); err != nil {
					// Non-fatal: log but continue with other deps
					e.Logger.Errf(logger.Red, "transparent: warning: could not compile dep %q: %v\n", dep.Task, err)
				}
			}
		}
	}

	// Find subtask calls in commands and recursively compile them
	if compiled.Cmds != nil {
		for i, cmd := range compiled.Cmds {
			if cmd != nil && cmd.Task != "" {
				e.Compiler.Tracer.RecordSubtaskCall(call.Task, i, cmd.Task)
				subtaskCall := &Call{Task: cmd.Task, Vars: cmd.Vars, Indirect: true}
				if err := e.compileTaskRecursive(subtaskCall, visited); err != nil {
					// Non-fatal: log but continue with other commands
					e.Logger.Errf(logger.Red, "transparent: warning: could not compile subtask %q: %v\n", cmd.Task, err)
				}
			}
		}
	}

	return nil
}

// RunTransparentAll compiles ALL tasks in the Taskfile with tracing enabled.
// Used with --transparent --list-all.
func (e *Executor) RunTransparentAll(ctx context.Context) error {
	var calls []*Call
	for name := range e.Taskfile.Tasks.All(nil) {
		calls = append(calls, &Call{Task: name})
	}
	return e.RunTransparent(ctx, calls...)
}
