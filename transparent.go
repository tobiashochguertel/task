package task

import (
	"context"
	"fmt"
	"os"

	"github.com/go-task/task/v3/internal/transparent"
)

// RunTransparent compiles all requested tasks with tracing enabled
// and renders a diagnostic report instead of executing commands.
func (e *Executor) RunTransparent(ctx context.Context, calls ...*Call) error {
	// Ensure tracer exists
	if e.Compiler.Tracer == nil {
		e.Compiler.Tracer = transparent.NewTracer()
	}

	for _, call := range calls {
		e.Compiler.Tracer.SetCurrentTask(call.Task)

		// Compile the task (resolves variables and templates) without executing
		compiled, err := e.compiledTask(call, true)
		if err != nil {
			return fmt.Errorf("transparent: error compiling task %q: %w", call.Task, err)
		}

		// Record dependencies
		if compiled.Deps != nil {
			for _, dep := range compiled.Deps {
				if dep != nil {
					e.Compiler.Tracer.RecordDep(call.Task, dep.Task)
				}
			}
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

// RunTransparentAll compiles ALL tasks in the Taskfile with tracing enabled.
// Used with --transparent --list-all.
func (e *Executor) RunTransparentAll(ctx context.Context) error {
	var calls []*Call
	for name := range e.Taskfile.Tasks.All(nil) {
		calls = append(calls, &Call{Task: name})
	}
	return e.RunTransparent(ctx, calls...)
}
