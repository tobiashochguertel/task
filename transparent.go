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

	// Render the report
	report := e.Compiler.Tracer.Report()
	transparent.RenderText(os.Stderr, report)
	return nil
}
