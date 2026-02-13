package transparent

import (
	"bytes"
	"strings"
	"testing"

	"github.com/go-task/template"
)

// defaultFuncs returns a minimal FuncMap with numeric and string functions for testing.
func defaultFuncs() template.FuncMap {
	return template.FuncMap{
		"add":   func(a, b any) any { return 0 },
		"sub":   func(a, b any) any { return 0 },
		"mul":   func(a, b any) any { return 0 },
		"div":   func(a, b any) any { return 0 },
		"mod":   func(a, b any) any { return 0 },
		"upper": func(s string) string { return s },
	}
}

func TestTracerNilSafe(t *testing.T) {
	t.Parallel()
	var tracer *Tracer
	// All methods should be no-ops on nil receiver
	tracer.SetCurrentTask("test")
	tracer.RecordVar(VarTrace{Name: "FOO", Value: "bar"})
	tracer.RecordTemplate(TemplateTrace{Input: "{{.FOO}}", Output: "bar"})
	tracer.RecordCmd("test", CmdTrace{RawCmd: "echo foo"})
	tracer.RecordDep("test", "dep")
	report := tracer.Report()
	if report == nil {
		t.Fatal("expected non-nil report from nil tracer")
	}
	if len(report.Tasks) != 0 {
		t.Errorf("expected 0 tasks, got %d", len(report.Tasks))
	}
}

func TestTracerRecordVar(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	// Record a global var
	tracer.RecordVar(VarTrace{Name: "FOO", Value: "bar", Origin: OriginTaskfileVars})
	report := tracer.Report()
	if len(report.GlobalVars) != 1 {
		t.Fatalf("expected 1 global var, got %d", len(report.GlobalVars))
	}
	if report.GlobalVars[0].Name != "FOO" {
		t.Errorf("expected FOO, got %s", report.GlobalVars[0].Name)
	}
	if report.GlobalVars[0].Type != "string" {
		t.Errorf("expected type string, got %s", report.GlobalVars[0].Type)
	}
}

func TestTracerRecordVarMultipleOrigins(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	tracer.RecordVar(VarTrace{Name: "A", Value: "env-val", Origin: OriginEnvironment})
	tracer.RecordVar(VarTrace{Name: "B", Value: "special-val", Origin: OriginSpecial})
	tracer.RecordVar(VarTrace{Name: "C", Value: "tf-env-val", Origin: OriginTaskfileEnv})
	tracer.RecordVar(VarTrace{Name: "D", Value: true, Origin: OriginTaskfileVars})

	report := tracer.Report()
	if len(report.GlobalVars) != 4 {
		t.Fatalf("expected 4 global vars, got %d", len(report.GlobalVars))
	}
	// Check type detection
	if report.GlobalVars[3].Type != "bool" {
		t.Errorf("expected type bool, got %s", report.GlobalVars[3].Type)
	}
}

func TestTracerShadowDetection(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	// Record global var
	tracer.RecordVar(VarTrace{Name: "NAME", Value: "global", Origin: OriginTaskfileVars})

	// Switch to task scope and record same var
	tracer.SetCurrentTask("build")
	tracer.RecordVar(VarTrace{Name: "NAME", Value: "task", Origin: OriginTaskVars})

	report := tracer.Report()
	if len(report.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(report.Tasks))
	}
	taskVars := report.Tasks[0].Vars
	if len(taskVars) != 1 {
		t.Fatalf("expected 1 task var, got %d", len(taskVars))
	}
	if taskVars[0].ShadowsVar == nil {
		t.Fatal("expected ShadowsVar to be set")
	}
	if taskVars[0].ShadowsVar.Name != "NAME" {
		t.Errorf("expected shadow name NAME, got %s", taskVars[0].ShadowsVar.Name)
	}
	if taskVars[0].ShadowsVar.Origin != OriginTaskfileVars {
		t.Errorf("expected shadow origin taskfile:vars, got %s", taskVars[0].ShadowsVar.Origin)
	}
}

func TestTracerShadowWithinGlobalScope(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	// Two global vars with same name (e.g., env then taskfile override)
	tracer.RecordVar(VarTrace{Name: "FOO", Value: "env", Origin: OriginEnvironment})
	tracer.RecordVar(VarTrace{Name: "FOO", Value: "taskfile", Origin: OriginTaskfileVars})

	report := tracer.Report()
	if len(report.GlobalVars) != 2 {
		t.Fatalf("expected 2 global vars, got %d", len(report.GlobalVars))
	}
	if report.GlobalVars[1].ShadowsVar == nil {
		t.Fatal("expected second FOO to shadow first")
	}
	if report.GlobalVars[1].ShadowsVar.Origin != OriginEnvironment {
		t.Errorf("expected shadow origin environment, got %s", report.GlobalVars[1].ShadowsVar.Origin)
	}
}

func TestTracerNoShadowForDifferentNames(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	tracer.RecordVar(VarTrace{Name: "FOO", Value: "1", Origin: OriginTaskfileVars})
	tracer.SetCurrentTask("build")
	tracer.RecordVar(VarTrace{Name: "BAR", Value: "2", Origin: OriginTaskVars})

	report := tracer.Report()
	if report.Tasks[0].Vars[0].ShadowsVar != nil {
		t.Error("expected no shadow for different variable names")
	}
}

func TestTracerTemplateAndCmd(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	tracer.SetCurrentTask("test")

	tracer.RecordTemplate(TemplateTrace{
		Input:    "{{.FOO}}",
		Output:   "bar",
		VarsUsed: []string{"FOO"},
	})
	tracer.RecordCmd("test", CmdTrace{
		Index:       0,
		RawCmd:      "echo {{.FOO}}",
		ResolvedCmd: "echo bar",
	})

	report := tracer.Report()
	if len(report.Tasks[0].Templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(report.Tasks[0].Templates))
	}
	if len(report.Tasks[0].Cmds) != 1 {
		t.Fatalf("expected 1 cmd, got %d", len(report.Tasks[0].Cmds))
	}
}

func TestTracerTemplateNotRecordedInGlobalScope(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	// Templates in global scope should be silently dropped
	tracer.RecordTemplate(TemplateTrace{Input: "{{.X}}", Output: "y"})
	report := tracer.Report()
	if len(report.Tasks) != 0 {
		t.Error("expected no tasks when recording template without SetCurrentTask")
	}
}

func TestTracerMultipleTasks(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()

	tracer.SetCurrentTask("build")
	tracer.RecordVar(VarTrace{Name: "A", Value: "1", Origin: OriginTaskVars})

	tracer.SetCurrentTask("test")
	tracer.RecordVar(VarTrace{Name: "B", Value: "2", Origin: OriginTaskVars})

	tracer.SetCurrentTask("deploy")
	tracer.RecordVar(VarTrace{Name: "C", Value: "3", Origin: OriginTaskVars})

	report := tracer.Report()
	if len(report.Tasks) != 3 {
		t.Fatalf("expected 3 tasks, got %d", len(report.Tasks))
	}
	// Verify order preservation
	if report.Tasks[0].TaskName != "build" {
		t.Errorf("expected first task build, got %s", report.Tasks[0].TaskName)
	}
	if report.Tasks[1].TaskName != "test" {
		t.Errorf("expected second task test, got %s", report.Tasks[1].TaskName)
	}
	if report.Tasks[2].TaskName != "deploy" {
		t.Errorf("expected third task deploy, got %s", report.Tasks[2].TaskName)
	}
}

func TestTracerDeps(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	tracer.SetCurrentTask("deploy")
	tracer.RecordDep("deploy", "build")
	tracer.RecordDep("deploy", "test")

	report := tracer.Report()
	deps := report.Tasks[0].Deps
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps, got %d", len(deps))
	}
	if deps[0] != "build" || deps[1] != "test" {
		t.Errorf("expected [build, test], got %v", deps)
	}
}

func TestTracerDynamicVar(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	tracer.RecordVar(VarTrace{
		Name:      "HOST",
		Value:     "myhost",
		Origin:    OriginTaskfileVars,
		IsDynamic: true,
		ShCmd:     "hostname",
	})
	report := tracer.Report()
	v := report.GlobalVars[0]
	if !v.IsDynamic {
		t.Error("expected IsDynamic=true")
	}
	if v.ShCmd != "hostname" {
		t.Errorf("expected ShCmd=hostname, got %s", v.ShCmd)
	}
}

func TestTracerRefTracking(t *testing.T) {
	t.Parallel()
	tracer := NewTracer()
	tracer.RecordVar(VarTrace{
		Name:    "ALIAS",
		Value:   "original",
		Origin:  OriginIncludeVars,
		IsRef:   true,
		RefName: "ORIGINAL_VAR",
	})
	report := tracer.Report()
	v := report.GlobalVars[0]
	if !v.IsRef {
		t.Error("expected IsRef=true")
	}
	if v.RefName != "ORIGINAL_VAR" {
		t.Errorf("expected RefName=ORIGINAL_VAR, got %s", v.RefName)
	}
}

func TestComputeValueID(t *testing.T) {
	t.Parallel()
	slice := []string{"a", "b"}
	vt := VarTrace{Name: "LIST", Value: slice}
	vt.ComputeValueID()
	if vt.ValueID == 0 {
		t.Error("expected non-zero ValueID for slice")
	}

	// Scalar should remain 0
	vt2 := VarTrace{Name: "STR", Value: "hello"}
	vt2.ComputeValueID()
	if vt2.ValueID != 0 {
		t.Error("expected zero ValueID for string scalar")
	}

	// Two vars pointing to same slice should have same ValueID
	vt3 := VarTrace{Name: "LIST2", Value: slice}
	vt3.ComputeValueID()
	if vt.ValueID != vt3.ValueID {
		t.Error("expected same ValueID for same slice")
	}

	// Nil value
	vt4 := VarTrace{Name: "NIL", Value: nil}
	vt4.ComputeValueID()
	if vt4.ValueID != 0 {
		t.Error("expected zero ValueID for nil")
	}

	// Map type
	m := map[string]string{"k": "v"}
	vt5 := VarTrace{Name: "MAP", Value: m}
	vt5.ComputeValueID()
	if vt5.ValueID == 0 {
		t.Error("expected non-zero ValueID for map")
	}
}

func TestTypeString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		value any
		want  string
	}{
		{nil, "nil"},
		{"hello", "string"},
		{42, "int"},
		{true, "bool"},
		{[]string{"a"}, "[]string"},
		{map[string]string{}, "map[string]string"},
	}
	for _, tt := range tests {
		if got := TypeString(tt.value); got != tt.want {
			t.Errorf("TypeString(%v) = %q, want %q", tt.value, got, tt.want)
		}
	}
}

func TestRenderText(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "build",
				Vars: []VarTrace{
					{Name: "NAME", Value: "World", Origin: OriginTaskfileVars, Type: "string"},
				},
				Templates: []TemplateTrace{
					{Input: "{{.NAME}}", Output: "World", VarsUsed: []string{"NAME"}},
				},
				Cmds: []CmdTrace{
					{Index: 0, RawCmd: "echo {{.NAME}}", ResolvedCmd: "echo World"},
				},
			},
		},
	}

	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()

	checks := []string{
		"TRANSPARENT MODE",
		"Task: build",
		"Variables in scope:",
		"NAME",
		"taskfile-vars",
		"World",
		"Template Evaluation",
		"{{.NAME}}",
		"Commands",
		"echo {{.NAME}}",
		"echo World",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("expected output to contain %q", check)
		}
	}
}

func TestRenderTextNilReport(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	RenderText(&buf, nil, nil)
	if buf.Len() != 0 {
		t.Error("expected empty output for nil report")
	}
}

func TestRenderTextWithShadow(t *testing.T) {
	t.Parallel()
	shadowedVar := VarTrace{Name: "X", Value: "old", Origin: OriginTaskfileVars}
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Vars: []VarTrace{
					{
						Name: "X", Value: "new", Origin: OriginTaskVars, Type: "string",
						ShadowsVar: &shadowedVar,
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()
	if !strings.Contains(output, "SHADOWS") {
		t.Error("expected output to contain shadow warning")
	}
}

func TestRenderTextWithDeps(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "deploy",
				Deps:     []string{"build", "test"},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()
	if !strings.Contains(output, "Dependencies:") {
		t.Error("expected output to contain Dependencies section")
	}
	if !strings.Contains(output, "build") || !strings.Contains(output, "test") {
		t.Error("expected output to list dep names")
	}
}

func TestRenderTextDynamicVar(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Vars: []VarTrace{
					{
						Name: "HOST", Value: "myhost", Origin: OriginTaskfileVars,
						Type: "string", IsDynamic: true,
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	if !strings.Contains(buf.String(), "(sh)") {
		t.Error("expected (sh) marker for dynamic var")
	}
}

func TestRenderTextRefVar(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Vars: []VarTrace{
					{
						Name: "ALIAS", Value: "val", Origin: OriginIncludeVars,
						Type: "string", IsRef: true, RefName: "ORIGINAL",
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()
	if !strings.Contains(output, "ref") {
		t.Error("expected ref marker")
	}
	if !strings.Contains(output, "ORIGINAL") {
		t.Error("expected RefName in output")
	}
}

func TestRenderTextUnchangedCmd(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Cmds: []CmdTrace{
					{Index: 0, RawCmd: "echo hello", ResolvedCmd: "echo hello"},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()
	// Unchanged cmd should show single "Command:" box, not "Raw:"/"Resolved:" split
	if strings.Contains(output, "Raw:") {
		t.Error("unchanged cmd should not show Raw:/Resolved: split")
	}
	if !strings.Contains(output, "Command:") {
		t.Error("unchanged cmd should show Command: box")
	}
}

func TestRenderTextPipeSteps(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Templates: []TemplateTrace{
					{
						Input:  "{{.NAME | trim | upper}}",
						Output: "WORLD",
						Steps: []PipeStep{
							{FuncName: "trim", Args: []string{".NAME"}, Output: "world"},
							{FuncName: "upper", Args: []string{}, Output: "WORLD"},
						},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()
	if !strings.Contains(output, "Step 1") {
		t.Error("expected pipe step output")
	}
	if !strings.Contains(output, "trim") {
		t.Error("expected trim in pipe steps")
	}
}

func TestVarOriginString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		origin VarOrigin
		want   string
	}{
		{OriginEnvironment, "environment"},
		{OriginSpecial, "special"},
		{OriginTaskfileEnv, "taskfile:env"},
		{OriginTaskfileVars, "taskfile:vars"},
		{OriginIncludeVars, "include:vars"},
		{OriginIncludedTaskfileVars, "included:taskfile:vars"},
		{OriginCallVars, "call:vars"},
		{OriginTaskVars, "task:vars"},
		{OriginForLoop, "for:loop"},
		{OriginDotenv, "dotenv"},
	}
	for _, tt := range tests {
		if got := tt.origin.String(); got != tt.want {
			t.Errorf("VarOrigin(%d).String() = %q, want %q", tt.origin, got, tt.want)
		}
	}
}

func TestVarOriginStringUnknown(t *testing.T) {
	t.Parallel()
	o := VarOrigin(999)
	s := o.String()
	if !strings.Contains(s, "unknown") {
		t.Errorf("expected unknown for invalid origin, got %s", s)
	}
}

// ── Pipe Analyzer Tests ──

func TestAnalyzePipesSimple(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"upper": strings.ToUpper,
		"trim":  strings.TrimSpace,
	}
	data := map[string]any{"NAME": "  hello  "}
	steps := AnalyzePipes("{{.NAME | trim}}", data, funcs)
	if len(steps) != 2 {
		t.Fatalf("expected 2 pipe steps, got %d", len(steps))
	}
	if steps[0].FuncName != ".NAME" {
		t.Errorf("step[0] func = %q, want .NAME", steps[0].FuncName)
	}
	if steps[0].Output != "  hello  " {
		t.Errorf("step[0] output = %q, want %q", steps[0].Output, "  hello  ")
	}
	if steps[1].FuncName != "trim" {
		t.Errorf("step[1] func = %q, want trim", steps[1].FuncName)
	}
	if steps[1].Output != "hello" {
		t.Errorf("step[1] output = %q, want hello", steps[1].Output)
	}
}

func TestAnalyzePipesThreeSteps(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"upper": strings.ToUpper,
		"trim":  strings.TrimSpace,
	}
	data := map[string]any{"NAME": "  hello  "}
	steps := AnalyzePipes("{{.NAME | trim | upper}}", data, funcs)
	if len(steps) != 3 {
		t.Fatalf("expected 3 pipe steps, got %d", len(steps))
	}
	if steps[2].FuncName != "upper" {
		t.Errorf("step[2] func = %q, want upper", steps[2].FuncName)
	}
	if steps[2].Output != "HELLO" {
		t.Errorf("step[2] output = %q, want HELLO", steps[2].Output)
	}
}

func TestAnalyzePipesNoPipe(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{}
	data := map[string]any{"FOO": "bar"}
	steps := AnalyzePipes("{{.FOO}}", data, funcs)
	if len(steps) != 0 {
		t.Fatalf("expected 0 pipe steps for single-command template, got %d", len(steps))
	}
}

func TestAnalyzePipesPlainText(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{}
	data := map[string]any{}
	steps := AnalyzePipes("no templates here", data, funcs)
	if len(steps) != 0 {
		t.Fatalf("expected 0 pipe steps for plain text, got %d", len(steps))
	}
}

func TestAnalyzePipesResolveArgs(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"trim": strings.TrimSpace,
	}
	data := map[string]any{"NAME": "  world  "}
	steps := AnalyzePipes("{{.NAME | trim}}", data, funcs)
	if len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	// First step (.NAME) has no extra args
	if len(steps[0].Args) != 0 {
		t.Errorf("step[0] args = %v, want empty", steps[0].Args)
	}
	// ArgsValues for .NAME should also be empty
	if len(steps[0].ArgsValues) != 0 {
		t.Errorf("step[0] argsValues = %v, want empty", steps[0].ArgsValues)
	}
}

func TestGeneratePipeTipsMultiArgPipe(t *testing.T) {
	t.Parallel()
	// printf with args piped to trim should generate a tip
	steps := []PipeStep{
		{FuncName: "printf", Args: []string{`"%s : %s"`, `"NAME"`, ".NAME"}},
		{FuncName: "trim", Args: nil},
	}
	tips := GeneratePipeTips(steps)
	if len(tips) == 0 {
		t.Error("expected at least one tip for printf piped to trim")
	}
	if len(tips) > 0 && !strings.Contains(tips[0], "printf") {
		t.Errorf("tip should mention printf, got: %s", tips[0])
	}
}

func TestGeneratePipeTipsNoTipForSingleArgPipe(t *testing.T) {
	t.Parallel()
	// .NAME | trim should NOT generate a tip (no multi-arg function)
	steps := []PipeStep{
		{FuncName: ".NAME", Args: nil},
		{FuncName: "trim", Args: nil},
	}
	tips := GeneratePipeTips(steps)
	if len(tips) != 0 {
		t.Errorf("expected no tips for simple field pipe, got: %v", tips)
	}
}

func TestGeneratePipeTipsEmptySteps(t *testing.T) {
	t.Parallel()
	tips := GeneratePipeTips(nil)
	if len(tips) != 0 {
		t.Errorf("expected no tips for nil steps, got: %v", tips)
	}

	tips = GeneratePipeTips([]PipeStep{})
	if len(tips) != 0 {
		t.Errorf("expected no tips for empty steps, got: %v", tips)
	}
}

// --- Tests for Feature 1: Verbose mode (filterGlobals) ---

func TestFilterGlobalsNonVerbose(t *testing.T) {
	t.Parallel()
	vars := []VarTrace{
		{Name: "TASK", Origin: OriginSpecial, Value: "default"},
		{Name: "MY_VAR", Origin: OriginTaskfileVars, Value: "hello"},
		{Name: "CLI_ARGS", Origin: OriginTaskfileVars, Value: ""},
		{Name: "CLI_FORCE", Origin: OriginTaskfileVars, Value: false},
		{Name: "CLI_SILENT", Origin: OriginTaskfileVars, Value: false},
		{Name: "CLI_VERBOSE", Origin: OriginTaskfileVars, Value: false},
		{Name: "CLI_OFFLINE", Origin: OriginTaskfileVars, Value: false},
		{Name: "CLI_ASSUME_YES", Origin: OriginTaskfileVars, Value: false},
		{Name: "CLI_ARGS_LIST", Origin: OriginTaskfileVars, Value: []string{}},
		{Name: "FROM_ENV", Origin: OriginEnvironment, Value: "val"},
	}
	filtered := filterGlobals(vars, false)
	for _, v := range filtered {
		if v.Origin == OriginEnvironment {
			t.Errorf("non-verbose should hide env vars, got %s", v.Name)
		}
		if isInternalVar(v.Name) {
			t.Errorf("non-verbose should hide internal var %s", v.Name)
		}
	}
	if len(filtered) != 2 { // TASK + MY_VAR
		t.Errorf("expected 2 vars after filter, got %d", len(filtered))
	}
}

func TestFilterGlobalsVerbose(t *testing.T) {
	t.Parallel()
	vars := []VarTrace{
		{Name: "TASK", Origin: OriginSpecial, Value: "default"},
		{Name: "CLI_ARGS", Origin: OriginTaskfileVars, Value: ""},
		{Name: "FROM_ENV", Origin: OriginEnvironment, Value: "val"},
	}
	filtered := filterGlobals(vars, true)
	if len(filtered) != len(vars) {
		t.Errorf("verbose should keep all vars: got %d, want %d", len(filtered), len(vars))
	}
}

func TestRenderTextVerboseHidesInternalVars(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		GlobalVars: []VarTrace{
			{Name: "TASK", Origin: OriginSpecial, Value: "x"},
			{Name: "MY_VAR", Origin: OriginTaskfileVars, Value: "hello"},
			{Name: "CLI_ARGS", Origin: OriginTaskfileVars, Value: ""},
			{Name: "CLI_FORCE", Origin: OriginTaskfileVars, Value: false},
			{Name: "CLI_SILENT", Origin: OriginTaskfileVars, Value: false},
			{Name: "CLI_VERBOSE", Origin: OriginTaskfileVars, Value: false},
			{Name: "CLI_OFFLINE", Origin: OriginTaskfileVars, Value: false},
			{Name: "CLI_ASSUME_YES", Origin: OriginTaskfileVars, Value: false},
			{Name: "CLI_ARGS_LIST", Origin: OriginTaskfileVars, Value: []string{}},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()
	if !strings.Contains(output, "environment variables hidden") {
		t.Error("non-verbose output should contain hidden vars message")
	}
	if strings.Contains(output, "CLI_ARGS ") {
		t.Error("non-verbose output should not show CLI_ARGS")
	}

	buf.Reset()
	RenderText(&buf, report, &RenderOptions{Verbose: true})
	output = buf.String()
	if strings.Contains(output, "hidden") {
		t.Error("verbose output should not contain hidden message")
	}
	if !strings.Contains(output, "CLI_ARGS") {
		t.Error("verbose output should show CLI_ARGS")
	}
}

// --- Tests for Feature 2: Dynamic var warning ---

func TestRenderTextDynamicVarEmptyWarning(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Vars: []VarTrace{
					{Name: "DYN", Origin: OriginTaskfileVars, Value: "", IsDynamic: true, ShCmd: "echo hello"},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()
	if !strings.Contains(output, "DYNAMIC") || !strings.Contains(output, "not evaluated") {
		t.Errorf("expected dynamic var warning, got:\n%s", output)
	}
}

func TestRenderTextDynamicVarNoWarningWhenResolved(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Vars: []VarTrace{
					{Name: "DYN", Origin: OriginTaskfileVars, Value: "resolved-val", IsDynamic: true, ShCmd: "echo hello"},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()
	if strings.Contains(output, "not evaluated") {
		t.Errorf("should not show warning for resolved dynamic var, got:\n%s", output)
	}
}

// --- Tests for Feature 3: Type mismatch detection ---

func TestDetectTypeMismatchesStringInAdd(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"COUNT": 42,
		"NAME":  "hello",
	}
	warnings := DetectTypeMismatches("{{add .COUNT .NAME}}", data, defaultFuncs())
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], "add()") || !strings.Contains(warnings[0], "NAME") {
		t.Errorf("warning should mention add() and NAME, got: %s", warnings[0])
	}
}

func TestDetectTypeMismatchesValidNumeric(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"A": 10,
		"B": 20,
	}
	warnings := DetectTypeMismatches("{{add .A .B}}", data, defaultFuncs())
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for valid numeric, got: %v", warnings)
	}
}

func TestDetectTypeMismatchesNoTemplate(t *testing.T) {
	t.Parallel()
	warnings := DetectTypeMismatches("just plain text", nil, defaultFuncs())
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for plain text, got: %v", warnings)
	}
}

func TestDetectTypeMismatchesMulWithString(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"NUM":   5,
		"LABEL": "abc",
	}
	warnings := DetectTypeMismatches("{{mul .NUM .LABEL}}", data, defaultFuncs())
	if len(warnings) != 1 {
		t.Fatalf("expected 1 warning, got %d: %v", len(warnings), warnings)
	}
	if !strings.Contains(warnings[0], "mul()") {
		t.Errorf("warning should mention mul(), got: %s", warnings[0])
	}
}

func TestDetectTypeMismatchesFloat(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"PRICE": 9.99,
		"QTY":   3,
	}
	warnings := DetectTypeMismatches("{{mul .PRICE .QTY}}", data, defaultFuncs())
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for float*int, got: %v", warnings)
	}
}

func TestDetectTypeMismatchesNonNumericFunc(t *testing.T) {
	t.Parallel()
	data := map[string]any{
		"NAME": "hello",
	}
	warnings := DetectTypeMismatches("{{upper .NAME}}", data, defaultFuncs())
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for non-numeric func, got: %v", warnings)
	}
}

// --- Tests for splitMultiline ---

func TestSplitMultilineSingleLine(t *testing.T) {
	t.Parallel()
	first, extra := splitMultiline("hello")
	if first != "hello" {
		t.Errorf("first = %q, want %q", first, "hello")
	}
	if len(extra) != 0 {
		t.Errorf("extra = %v, want empty", extra)
	}
}

func TestSplitMultilineMultipleLines(t *testing.T) {
	t.Parallel()
	input := "line1\nline2\nline3"
	first, extra := splitMultiline(input)
	if first != "line1" {
		t.Errorf("first = %q, want %q", first, "line1")
	}
	if len(extra) != 2 {
		t.Fatalf("extra length = %d, want 2", len(extra))
	}
	if extra[0] != "line2" || extra[1] != "line3" {
		t.Errorf("extra = %v, want [line2 line3]", extra)
	}
}

func TestSplitMultilineSkipsEmptyLines(t *testing.T) {
	t.Parallel()
	input := "line1\n\n  \nline4"
	first, extra := splitMultiline(input)
	if first != "line1" {
		t.Errorf("first = %q, want %q", first, "line1")
	}
	if len(extra) != 1 || extra[0] != "line4" {
		t.Errorf("extra = %v, want [line4]", extra)
	}
}

func TestSplitMultilinePreservesLongLines(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("a", 200)
	input := "short\n" + long
	first, extra := splitMultiline(input)
	if first != "short" {
		t.Errorf("first = %q, want %q", first, "short")
	}
	if len(extra) != 1 {
		t.Fatalf("extra length = %d, want 1", len(extra))
	}
	// No truncation — full 200-char line preserved
	if extra[0] != long {
		t.Error("long line should be preserved without truncation")
	}
}

// --- Tests for formatVarValue ---

func TestFormatVarValueString(t *testing.T) {
	t.Parallel()
	result := formatVarValue("hello world")
	if result != "hello world" {
		t.Errorf("got %q, want %q", result, "hello world")
	}
}

func TestFormatVarValueNil(t *testing.T) {
	t.Parallel()
	result := formatVarValue(nil)
	if result != "<nil>" {
		t.Errorf("got %q, want %q", result, "<nil>")
	}
}

func TestFormatVarValueMap(t *testing.T) {
	t.Parallel()
	m := map[string]any{"key": "val"}
	result := formatVarValue(m)
	// Should be pretty-printed JSON
	if !strings.Contains(result, "\"key\"") {
		t.Error("should contain JSON key")
	}
	if !strings.Contains(result, "\"val\"") {
		t.Error("should contain JSON value")
	}
	if !strings.Contains(result, "\n") {
		t.Error("should be multiline (indented JSON)")
	}
}

func TestFormatVarValueSlice(t *testing.T) {
	t.Parallel()
	s := []string{"a", "b", "c"}
	result := formatVarValue(s)
	if !strings.Contains(result, "\"a\"") {
		t.Error("should contain slice elements")
	}
}

func TestFormatVarValueBool(t *testing.T) {
	t.Parallel()
	result := formatVarValue(true)
	if result != "true" {
		t.Errorf("got %q, want %q", result, "true")
	}
}

// --- Tests for renderVars dispatching ---

func TestRenderVarsCustomRenderer(t *testing.T) {
	t.Parallel()
	vars := []VarTrace{
		{Name: "FOO", Origin: OriginTaskfileVars, Value: "bar", Type: "string"},
	}
	var buf bytes.Buffer
	renderVars(&buf, vars, &RenderOptions{TableRenderer: "custom"})
	output := buf.String()
	if !strings.Contains(output, "┌") {
		t.Error("custom renderer should use box-drawing characters")
	}
	if !strings.Contains(output, "FOO") {
		t.Error("custom renderer should contain variable name")
	}
	if !strings.Contains(output, "bar") {
		t.Error("custom renderer should contain variable value inline")
	}
}

func TestRenderVarsLipglossRenderer(t *testing.T) {
	t.Parallel()
	vars := []VarTrace{
		{Name: "FOO", Origin: OriginTaskfileVars, Value: "bar", Type: "string"},
	}
	var buf bytes.Buffer
	renderVars(&buf, vars, &RenderOptions{TableRenderer: "lipgloss"})
	output := buf.String()
	if !strings.Contains(output, "FOO") {
		t.Error("lipgloss renderer should contain variable name")
	}
	if !strings.Contains(output, "bar") {
		t.Error("lipgloss renderer should contain variable value")
	}
	if !strings.Contains(output, "Variables in scope") {
		t.Error("lipgloss renderer should contain header")
	}
}

func TestRenderVarsDefaultRendererIsCustom(t *testing.T) {
	t.Parallel()
	vars := []VarTrace{
		{Name: "X", Origin: OriginSpecial, Value: "y", Type: "string"},
	}
	var buf bytes.Buffer
	renderVars(&buf, vars, nil)
	output := buf.String()
	if !strings.Contains(output, "┌") {
		t.Error("nil opts should default to custom renderer with box-drawing chars")
	}
}

// --- Tests for multiline value detail boxes ---

func TestRenderVarsCustomMapValueShowsDetailBox(t *testing.T) {
	t.Parallel()
	mapValue := map[string]any{"key1": "val1", "key2": "val2"}
	vars := []VarTrace{
		{Name: "MAP_VAR", Origin: OriginSpecial, Value: mapValue, Type: "map[string]any"},
	}
	var buf bytes.Buffer
	renderVars(&buf, vars, &RenderOptions{})
	output := buf.String()
	// Table should show summary with "see below"
	if !strings.Contains(output, "see below") {
		t.Error("multiline value should show 'see below' in table")
	}
	// Detail box should show full value
	if !strings.Contains(output, "Value of MAP_VAR") {
		t.Error("should render detail box with variable name")
	}
	if !strings.Contains(output, "\"key1\"") {
		t.Error("detail box should contain full JSON value")
	}
}

func TestRenderVarsLipglossMapValueShowsDetailBox(t *testing.T) {
	t.Parallel()
	mapValue := map[string]any{"name": "test"}
	vars := []VarTrace{
		{Name: "INFO", Origin: OriginSpecial, Value: mapValue, Type: "map[string]any"},
	}
	var buf bytes.Buffer
	renderVarsLipgloss(&buf, vars, &RenderOptions{})
	output := buf.String()
	if !strings.Contains(output, "see below") {
		t.Error("multiline value should show 'see below' in table")
	}
	if !strings.Contains(output, "Value of INFO") {
		t.Error("should render detail box with variable name")
	}
	if !strings.Contains(output, "\"name\"") {
		t.Error("detail box should contain full JSON value")
	}
}

func TestRenderVarsCustomInlineSimpleValue(t *testing.T) {
	t.Parallel()
	vars := []VarTrace{
		{Name: "SIMPLE", Origin: OriginTaskfileVars, Value: "just a string", Type: "string"},
	}
	var buf bytes.Buffer
	renderVars(&buf, vars, &RenderOptions{})
	output := buf.String()
	// Simple string values should be shown inline, no detail box
	if !strings.Contains(output, "just a string") {
		t.Error("simple value should be shown inline")
	}
	if strings.Contains(output, "see below") {
		t.Error("simple value should not have 'see below' indicator")
	}
}

func TestRenderVarsLipglossShadow(t *testing.T) {
	t.Parallel()
	vars := []VarTrace{
		{
			Name: "X", Origin: OriginTaskVars, Value: "new", Type: "string",
			ShadowsVar: &VarTrace{Name: "X", Origin: OriginTaskfileVars, Value: "old"},
		},
	}
	var buf bytes.Buffer
	renderVarsLipgloss(&buf, vars, &RenderOptions{})
	output := buf.String()
	if !strings.Contains(output, "SHADOWS") {
		t.Error("lipgloss renderer should show shadow indicator")
	}
}

// --- Tests for syntax highlighting ---

func TestChromaHighlightReturnsANSIForJSON(t *testing.T) {
	t.Parallel()
	input := `{"key": "value"}`
	result := chromaHighlight(input)
	if result == "" {
		t.Fatal("chromaHighlight should return non-empty result for valid JSON")
	}
	// Should contain ANSI escape codes
	if !strings.Contains(result, "\x1b[") {
		t.Error("highlighted output should contain ANSI escape codes")
	}
	// Should still contain the original content (minus ANSI)
	stripped := ansiRegex.ReplaceAllString(result, "")
	if !strings.Contains(stripped, "key") || !strings.Contains(stripped, "value") {
		t.Error("highlighted output should preserve original content")
	}
}

func TestChromaHighlightEndsWithReset(t *testing.T) {
	t.Parallel()
	result := chromaHighlight(`{"a": 1}`)
	if !strings.HasSuffix(result, ansiReset) {
		t.Error("highlighted output should end with ANSI reset to prevent color bleeding")
	}
}

func TestChromaHighlightMultilineJSON(t *testing.T) {
	t.Parallel()
	input := "{\n  \"key\": \"value\",\n  \"num\": 42\n}"
	result := chromaHighlight(input)
	if result == "" {
		t.Fatal("chromaHighlight should handle multiline JSON")
	}
	stripped := ansiRegex.ReplaceAllString(result, "")
	if !strings.Contains(stripped, "\"key\"") {
		t.Error("multiline JSON should preserve keys")
	}
	if !strings.Contains(stripped, "42") {
		t.Error("multiline JSON should preserve numbers")
	}
}

func TestSyntaxHighlightSkipsNonJSON(t *testing.T) {
	t.Parallel()
	input := "just a plain string"
	result := syntaxHighlight(input, false)
	if result != input {
		t.Errorf("non-JSON content should be returned unchanged, got %q", result)
	}
}

func TestSyntaxHighlightSkipsEmptyContent(t *testing.T) {
	t.Parallel()
	result := syntaxHighlight("", false)
	if result != "" {
		t.Errorf("empty content should be returned unchanged, got %q", result)
	}
}

func TestSyntaxHighlightHandlesJSONObject(t *testing.T) {
	t.Parallel()
	input := "{\n  \"name\": \"test\"\n}"
	result := syntaxHighlight(input, false)
	// When color is enabled (default in tests), result should have ANSI codes
	if !strings.Contains(result, "\x1b[") {
		t.Skip("color is disabled in this environment")
	}
	stripped := ansiRegex.ReplaceAllString(result, "")
	if !strings.Contains(stripped, "\"name\"") {
		t.Error("highlighted JSON should preserve content")
	}
}

func TestSyntaxHighlightHandlesJSONArray(t *testing.T) {
	t.Parallel()
	input := "[\"a\", \"b\", \"c\"]"
	result := syntaxHighlight(input, false)
	if !strings.Contains(result, "\x1b[") {
		t.Skip("color is disabled in this environment")
	}
	stripped := ansiRegex.ReplaceAllString(result, "")
	if !strings.Contains(stripped, "\"a\"") {
		t.Error("highlighted JSON array should preserve content")
	}
}

func TestUndoWhitespaceVisible(t *testing.T) {
	t.Parallel()
	// Simulate makeWhitespaceVisible output
	input := "{\u21b5\n\u00b7\u00b7\"key\":\u00b7\"value\"\u21b5\n}"
	result := undoWhitespaceVisible(input)
	expected := "{\n  \"key\": \"value\"\n}"
	if result != expected {
		t.Errorf("undoWhitespaceVisible:\n  got:  %q\n  want: %q", result, expected)
	}
}

func TestUndoWhitespaceVisibleTabs(t *testing.T) {
	t.Parallel()
	input := "\u2192indented"
	result := undoWhitespaceVisible(input)
	if result != "\tindented" {
		t.Errorf("should convert → to tab, got %q", result)
	}
}

func TestUndoWhitespaceVisibleCRLF(t *testing.T) {
	t.Parallel()
	input := "line\u2190\u21b5\n"
	result := undoWhitespaceVisible(input)
	if result != "line\r\n" {
		t.Errorf("should restore CRLF, got %q", result)
	}
}

func TestUndoWhitespaceVisibleStripsESCMarker(t *testing.T) {
	t.Parallel()
	input := "before[ESC]after"
	result := undoWhitespaceVisible(input)
	if result != "beforeafter" {
		t.Errorf("should strip [ESC] markers, got %q", result)
	}
}

func TestReapplyWhitespaceVisible(t *testing.T) {
	t.Parallel()
	input := "{\n  \"key\": \"value\"\n}"
	result := reapplyWhitespaceVisible(input)
	if !strings.Contains(result, "\u00b7") {
		t.Error("should replace spaces with ·")
	}
	if !strings.Contains(result, "\u21b5") {
		t.Error("should add ↵ before newlines")
	}
	if strings.Contains(result, " ") {
		t.Error("should not have any literal spaces remaining")
	}
}

func TestReapplyWhitespaceVisiblePreservesANSI(t *testing.T) {
	t.Parallel()
	// ANSI codes don't contain spaces or tabs, so they should survive
	input := "\x1b[38;5;197m\"key\"\x1b[0m: \x1b[38;5;186m\"val\"\x1b[0m"
	result := reapplyWhitespaceVisible(input)
	// ANSI codes should be preserved (they don't contain space/tab)
	if !strings.Contains(result, "\x1b[38;5;197m") {
		t.Error("ANSI codes should be preserved after reapplyWhitespaceVisible")
	}
	// Spaces should be replaced
	if strings.Contains(result, " ") {
		t.Error("spaces should be replaced with · even around ANSI codes")
	}
}

func TestSyntaxHighlightRoundtripWithWhitespaces(t *testing.T) {
	t.Parallel()
	// Original JSON
	original := "{\n  \"key\": \"value\"\n}"
	// Apply whitespace visibility (simulating the pipeline)
	wsVisible := makeWhitespaceVisible(original)

	result := syntaxHighlight(wsVisible, true)
	if !strings.Contains(result, "\x1b[") {
		t.Skip("color is disabled in this environment")
	}
	// After highlighting with showWhitespaces=true, whitespace markers should be present
	if !strings.Contains(result, "\u00b7") {
		t.Error("whitespace markers should be re-applied after highlighting")
	}
	// Content should be preserved
	stripped := ansiRegex.ReplaceAllString(result, "")
	if !strings.Contains(stripped, "\"key\"") {
		t.Error("JSON content should be preserved through highlight roundtrip")
	}
}

func TestSyntaxHighlightNoColorEnv(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	input := `{"key": "value"}`
	result := syntaxHighlight(input, false)
	if result != input {
		t.Error("with NO_COLOR=1, syntaxHighlight should return content unchanged")
	}
}

func TestRenderVarsCustomDetailBoxHighlighted(t *testing.T) {
	t.Parallel()
	mapValue := map[string]any{"name": "test"}
	vars := []VarTrace{
		{Name: "INFO", Origin: OriginSpecial, Value: mapValue, Type: "map[string]any"},
	}
	var buf bytes.Buffer
	renderVars(&buf, vars, &RenderOptions{})
	output := buf.String()
	if !strings.Contains(output, "Value of INFO") {
		t.Error("should render detail box for multiline value")
	}
	// When color is active, the detail box should contain ANSI from chroma
	if strings.Contains(output, "\x1b[38;5;") {
		// Chroma codes are present — verify JSON content is preserved
		stripped := ansiRegex.ReplaceAllString(output, "")
		if !strings.Contains(stripped, "\"name\"") {
			t.Error("highlighted detail box should preserve JSON content")
		}
	}
}

func TestRenderVarsLipglossDetailBoxHighlighted(t *testing.T) {
	t.Parallel()
	mapValue := map[string]any{"name": "test"}
	vars := []VarTrace{
		{Name: "INFO", Origin: OriginSpecial, Value: mapValue, Type: "map[string]any"},
	}
	var buf bytes.Buffer
	renderVarsLipgloss(&buf, vars, &RenderOptions{})
	output := buf.String()
	if !strings.Contains(output, "Value of INFO") {
		t.Error("should render detail box for multiline value")
	}
	if strings.Contains(output, "\x1b[38;5;") {
		stripped := ansiRegex.ReplaceAllString(output, "")
		if !strings.Contains(stripped, "\"name\"") {
			t.Error("highlighted detail box should preserve JSON content")
		}
	}
}

func TestRenderVarsDetailBoxWithShowWhitespaces(t *testing.T) {
	t.Parallel()
	mapValue := map[string]any{"key": "hello"}
	vars := []VarTrace{
		{Name: "DATA", Origin: OriginSpecial, Value: mapValue, Type: "map[string]any"},
	}
	// Apply whitespace visibility first (matching RenderText flow)
	wsVars := applyWSToVars(vars)
	var buf bytes.Buffer
	renderVars(&buf, wsVars, &RenderOptions{ShowWhitespaces: true})
	output := buf.String()
	if !strings.Contains(output, "Value of DATA") {
		t.Error("should render detail box")
	}
	// Content should have whitespace markers
	stripped := ansiRegex.ReplaceAllString(output, "")
	if !strings.Contains(stripped, "\u00b7") {
		t.Error("detail box should have whitespace markers when ShowWhitespaces is true")
	}
}

// --- Tests for whitespace visibility on complex values ---

func TestApplyWSToVarsMapValue(t *testing.T) {
	t.Parallel()
	vars := []VarTrace{
		{Name: "MAP", Value: map[string]any{"key": "hello world"}, Type: "map"},
	}
	result := applyWSToVars(vars)
	// The map should be converted to string with whitespace visible
	s, ok := result[0].Value.(string)
	if !ok {
		t.Fatal("map value should be converted to string after applyWSToVars")
	}
	if !strings.Contains(s, "·") {
		t.Error("whitespace should be made visible in map values")
	}
}

func TestApplyWSToVarsStringValue(t *testing.T) {
	t.Parallel()
	vars := []VarTrace{
		{Name: "STR", Value: "hello world", Type: "string"},
	}
	result := applyWSToVars(vars)
	s, ok := result[0].Value.(string)
	if !ok {
		t.Fatal("string value should remain string")
	}
	if !strings.Contains(s, "·") {
		t.Error("whitespace should be made visible")
	}
}
