package task

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/Ladicle/tabwriter"
	"golang.org/x/sync/errgroup"

	"github.com/go-task/task/v3/internal/editors"
	"github.com/go-task/task/v3/internal/fingerprint"
	"github.com/go-task/task/v3/internal/logger"
	"github.com/go-task/task/v3/internal/sort"
	"github.com/go-task/task/v3/taskfile/ast"
)

// ListOptions collects list-related options
type ListOptions struct {
	ListOnlyTasksWithDescriptions bool
	ListAllTasks                  bool
	FormatTaskListAsJSON          bool
	NoStatus                      bool
	Nested                        bool
}

// NewListOptions creates a new ListOptions instance
func NewListOptions(list, listAll, listAsJson, noStatus, nested bool) ListOptions {
	return ListOptions{
		ListOnlyTasksWithDescriptions: list,
		ListAllTasks:                  listAll,
		FormatTaskListAsJSON:          listAsJson,
		NoStatus:                      noStatus,
		Nested:                        nested,
	}
}

// ShouldListTasks returns true if one of the options to list tasks has been set to true
func (o ListOptions) ShouldListTasks() bool {
	return o.ListOnlyTasksWithDescriptions || o.ListAllTasks
}

// Filters returns the slice of FilterFunc which filters a list
// of ast.Task according to the given ListOptions
func (o ListOptions) Filters() []FilterFunc {
	filters := []FilterFunc{FilterOutInternal}

	if o.ListOnlyTasksWithDescriptions {
		filters = append(filters, FilterOutNoDesc)
	}

	return filters
}

// ListTasks prints a list of tasks.
// Tasks that match the given filters will be excluded from the list.
// The function returns a boolean indicating whether tasks were found
// and an error if one was encountered while preparing the output.
func (e *Executor) ListTasks(o ListOptions) (bool, error) {
	tasks, err := e.GetTaskList(o.Filters()...)
	if err != nil {
		return false, err
	}
	if o.FormatTaskListAsJSON {
		output, err := e.ToEditorOutput(tasks, o.NoStatus, o.Nested)
		if err != nil {
			return false, err
		}

		encoder := json.NewEncoder(e.Stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(output); err != nil {
			return false, err
		}

		return len(tasks) > 0, nil
	}
	if len(tasks) == 0 {
		if o.ListOnlyTasksWithDescriptions {
			e.Logger.Outf(logger.Yellow, "task: No tasks with description available. Try --list-all to list all tasks\n")
		} else if o.ListAllTasks {
			e.Logger.Outf(logger.Yellow, "task: No tasks available\n")
		}
		return false, nil
	}
	e.Logger.Outf(logger.Default, "task: Available tasks for this project:\n")

	// Format in tab-separated columns with a tab stop of 8.
	w := tabwriter.NewWriter(e.Stdout, 0, 8, 6, ' ', 0)
	for _, task := range tasks {
		e.Logger.FOutf(w, logger.Yellow, "* ")
		e.Logger.FOutf(w, logger.Green, task.Task)
		desc := strings.ReplaceAll(task.Desc, "\n", " ")
		e.Logger.FOutf(w, logger.Default, ": \t%s", desc)
		if len(task.Aliases) > 0 {
			e.Logger.FOutf(w, logger.Cyan, "\t(aliases: %s)", strings.Join(task.Aliases, ", "))
		}
		_, _ = fmt.Fprint(w, "\n")
	}
	if err := w.Flush(); err != nil {
		return false, err
	}
	return true, nil
}

// ListTaskNames prints only the task names in a Taskfile.
// Only tasks with a non-empty description are printed if allTasks is false.
// Otherwise, all task names are printed.
func (e *Executor) ListTaskNames(allTasks bool) error {
	// use stdout if no output defined
	var w io.Writer = os.Stdout
	if e.Stdout != nil {
		w = e.Stdout
	}

	// Sort the tasks
	if e.TaskSorter == nil {
		e.TaskSorter = sort.AlphaNumericWithRootTasksFirst
	}

	// Create a list of task names
	taskNames := make([]string, 0, e.Taskfile.Tasks.Len())
	for task := range e.Taskfile.Tasks.Values(e.TaskSorter) {
		if (allTasks || task.Desc != "") && !task.Internal {
			taskNames = append(taskNames, strings.TrimRight(task.Task, ":"))
			for _, alias := range task.Aliases {
				taskNames = append(taskNames, strings.TrimRight(alias, ":"))
			}
		}
	}
	for _, t := range taskNames {
		fmt.Fprintln(w, t)
	}
	return nil
}

// ListVariables prints a list of global variables with their descriptions
func (e *Executor) ListVariables(formatAsJson bool) error {
	vars := e.Taskfile.Vars

	if formatAsJson {
		type varInfo struct {
			Name  string `json:"name"`
			Desc  string `json:"desc,omitempty"`
			Value string `json:"value,omitempty"`
		}

		output := make([]varInfo, 0, vars.Len())
		for k, v := range vars.All() {
			vi := varInfo{
				Name: k,
				Desc: v.Desc,
			}
			// Include value for static variables only (not dynamic sh/ref)
			if v.Value != nil && v.Sh == nil && v.Ref == "" {
				vi.Value = fmt.Sprintf("%v", v.Value)
			}
			output = append(output, vi)
		}

		encoder := json.NewEncoder(e.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}

	// Format as human-readable table
	if vars.Len() == 0 {
		e.Logger.Outf(logger.Yellow, "task: No variables defined in this Taskfile\n")
		return nil
	}

	e.Logger.Outf(logger.Default, "task: Available variables for this project:\n")
	w := tabwriter.NewWriter(e.Stdout, 0, 8, 6, ' ', 0)
	for k, v := range vars.All() {
		e.Logger.FOutf(w, logger.Yellow, "* ")
		e.Logger.FOutf(w, logger.Green, k)
		if v.Desc != "" {
			e.Logger.FOutf(w, logger.Default, ": \t%s", v.Desc)
		}
		_, _ = fmt.Fprint(w, "\n")
	}
	return w.Flush()
}

func (e *Executor) ToEditorOutput(tasks []*ast.Task, noStatus bool, nested bool) (*editors.Namespace, error) {
	var g errgroup.Group
	editorTasks := make([]editors.Task, len(tasks))

	// Look over each task in parallel and turn it into an editor task
	for i := range tasks {
		g.Go(func() error {
			editorTask := editors.NewTask(tasks[i])

			if noStatus {
				editorTasks[i] = editorTask
				return nil
			}

			// Get the fingerprinting method to use
			method := e.Taskfile.Method
			if tasks[i].Method != "" {
				method = tasks[i].Method
			}
			upToDate, err := fingerprint.IsTaskUpToDate(context.Background(), tasks[i],
				fingerprint.WithMethod(method),
				fingerprint.WithTempDir(e.TempDir.Fingerprint),
				fingerprint.WithDry(e.Dry),
				fingerprint.WithLogger(e.Logger),
			)
			if err != nil {
				return err
			}

			editorTask.UpToDate = &upToDate
			editorTasks[i] = editorTask
			return nil
		})
	}
	if err := g.Wait(); err != nil {
		return nil, err
	}

	// Add global variables to the output
	editorVars := make([]editors.Variable, 0, e.Taskfile.Vars.Len())
	for k, v := range e.Taskfile.Vars.All() {
		editorVar := editors.Variable{
			Name: k,
			Desc: v.Desc,
		}
		// Include value for static variables only (not dynamic sh/ref)
		if v.Value != nil && v.Sh == nil && v.Ref == "" {
			editorVar.Value = fmt.Sprintf("%v", v.Value)
		}
		editorVars = append(editorVars, editorVar)
	}

	// Create the root namespace
	var tasksLen int
	if !nested {
		tasksLen = len(editorTasks)
	}
	rootNamespace := &editors.Namespace{
		Tasks:    make([]editors.Task, tasksLen),
		Vars:     editorVars,
		Location: e.Taskfile.Location,
	}

	// Recursively add namespaces to the root namespace or if nesting is
	// disabled add them all to the root namespace
	for i, task := range editorTasks {
		taskNamespacePath := strings.Split(task.Task, ast.NamespaceSeparator)
		if nested {
			rootNamespace.AddNamespace(taskNamespacePath, task)
		} else {
			rootNamespace.Tasks[i] = task
		}
	}

	return rootNamespace, g.Wait()
}
