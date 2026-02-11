package transparent

import "sync"

// Tracer collects variable resolution and template evaluation traces.
// All methods are nil-receiver safe — when the tracer is nil (transparent mode off),
// every method is a no-op with zero overhead.
type Tracer struct {
	mu          sync.Mutex
	currentTask string
	globalVars  []VarTrace
	tasks       map[string]*TaskTrace
	taskOrder   []string // preserves insertion order
}

// NewTracer creates a new Tracer instance.
func NewTracer() *Tracer {
	return &Tracer{
		tasks: make(map[string]*TaskTrace),
	}
}

// SetCurrentTask sets the task context for subsequent RecordVar/RecordTemplate calls.
// Pass "" to switch back to global scope.
func (t *Tracer) SetCurrentTask(taskName string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.currentTask = taskName
	if taskName != "" {
		if _, exists := t.tasks[taskName]; !exists {
			t.tasks[taskName] = &TaskTrace{TaskName: taskName}
			t.taskOrder = append(t.taskOrder, taskName)
		}
	}
}

// RecordVar records a variable resolution event.
func (t *Tracer) RecordVar(vt VarTrace) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	vt.Type = TypeString(vt.Value)
	vt.ComputeValueID()
	vt.TaskName = t.currentTask

	if t.currentTask == "" {
		// Check for shadow in global scope
		for i := range t.globalVars {
			if t.globalVars[i].Name == vt.Name {
				prev := t.globalVars[i]
				vt.ShadowsVar = &prev
				break
			}
		}
		t.globalVars = append(t.globalVars, vt)
	} else {
		tt := t.getOrCreateTask(t.currentTask)
		// Check for shadow against task vars and global vars
		for i := range tt.Vars {
			if tt.Vars[i].Name == vt.Name {
				prev := tt.Vars[i]
				vt.ShadowsVar = &prev
				break
			}
		}
		if vt.ShadowsVar == nil {
			for i := range t.globalVars {
				if t.globalVars[i].Name == vt.Name {
					prev := t.globalVars[i]
					vt.ShadowsVar = &prev
					break
				}
			}
		}
		tt.Vars = append(tt.Vars, vt)
	}
}

// RecordTemplate records a template evaluation event.
func (t *Tracer) RecordTemplate(tt TemplateTrace) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	if t.currentTask == "" {
		return
	}
	task := t.getOrCreateTask(t.currentTask)
	task.Templates = append(task.Templates, tt)
}

// RecordCmd records a command trace.
func (t *Tracer) RecordCmd(taskName string, ct CmdTrace) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	task := t.getOrCreateTask(taskName)
	task.Cmds = append(task.Cmds, ct)
}

// RecordDep records a dependency for a task.
func (t *Tracer) RecordDep(taskName string, depName string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	task := t.getOrCreateTask(taskName)
	task.Deps = append(task.Deps, depName)
}

// Report generates the final TraceReport.
func (t *Tracer) Report() *TraceReport {
	if t == nil {
		return &TraceReport{}
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	report := &TraceReport{
		GlobalVars: t.globalVars,
	}
	for _, name := range t.taskOrder {
		if tt, ok := t.tasks[name]; ok {
			report.Tasks = append(report.Tasks, tt)
		}
	}
	return report
}

func (t *Tracer) getOrCreateTask(name string) *TaskTrace {
	if tt, ok := t.tasks[name]; ok {
		return tt
	}
	tt := &TaskTrace{TaskName: name}
	t.tasks[name] = tt
	t.taskOrder = append(t.taskOrder, name)
	return tt
}
