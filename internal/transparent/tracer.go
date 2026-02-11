package transparent

import "sync"

// Tracer collects variable resolution and template evaluation traces.
// All methods are nil-receiver safe — when the tracer is nil (transparent mode off),
// every method is a no-op with zero overhead.
type Tracer struct {
	mu              sync.Mutex
	currentTask     string
	templateContext string // e.g. "cmds[0]", "label", "dir"
	globalVars      []VarTrace
	tasks           map[string]*TaskTrace
	taskOrder       []string // preserves insertion order
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

// SetTemplateContext sets a label (e.g. "cmds[0]", "label", "dir") for
// subsequent RecordTemplate calls. Reset with "".
func (t *Tracer) SetTemplateContext(ctx string) {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()
	t.templateContext = ctx
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
	if tt.Context == "" && t.templateContext != "" {
		tt.Context = t.templateContext
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

// isGlobalOrigin returns true for variable origins that belong to the
// Taskfile/global scope rather than a specific task.
func isGlobalOrigin(o VarOrigin) bool {
	switch o {
	case OriginSpecial, OriginEnvironment, OriginTaskfileEnv,
		OriginTaskfileVars, OriginIncludeVars,
		OriginIncludedTaskfileVars, OriginDotenv:
		return true
	}
	return false
}

// SeparateGlobalVars moves global-scope variables from task traces into the
// globalVars collection. Variables with global origins (special, taskfile:vars,
// taskfile:env, include:vars, dotenv) are extracted from the first task trace
// and stored once in globalVars, then removed from all task traces to avoid
// duplication.
func (t *Tracer) SeparateGlobalVars() {
	if t == nil {
		return
	}
	t.mu.Lock()
	defer t.mu.Unlock()

	if len(t.taskOrder) == 0 {
		return
	}

	// Collect global vars from the first task
	firstTask := t.tasks[t.taskOrder[0]]
	if firstTask == nil {
		return
	}

	globalNames := make(map[string]bool)
	var globals []VarTrace
	var taskOnly []VarTrace

	for _, v := range firstTask.Vars {
		if isGlobalOrigin(v.Origin) {
			globals = append(globals, v)
			globalNames[v.Name] = true
		} else {
			taskOnly = append(taskOnly, v)
		}
	}
	firstTask.Vars = taskOnly
	t.globalVars = append(t.globalVars, globals...)

	// Remove duplicated global vars from remaining tasks
	for i := 1; i < len(t.taskOrder); i++ {
		tt := t.tasks[t.taskOrder[i]]
		if tt == nil {
			continue
		}
		var filtered []VarTrace
		for _, v := range tt.Vars {
			if !globalNames[v.Name] || !isGlobalOrigin(v.Origin) {
				filtered = append(filtered, v)
			}
		}
		tt.Vars = filtered
	}
}
