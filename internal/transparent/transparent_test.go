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
	tracer := NewTracer()
	// Templates in global scope should be silently dropped
	tracer.RecordTemplate(TemplateTrace{Input: "{{.X}}", Output: "y"})
	report := tracer.Report()
	if len(report.Tasks) != 0 {
		t.Error("expected no tasks when recording template without SetCurrentTask")
	}
}

func TestTracerMultipleTasks(t *testing.T) {
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
		"Transparent Mode Report",
		"Task: build",
		"Variables:",
		"NAME",
		"taskfile-vars",
		"World",
		"Template Evaluations:",
		"{{.NAME}}",
		"Commands:",
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
	var buf bytes.Buffer
	RenderText(&buf, nil, nil)
	if buf.Len() != 0 {
		t.Error("expected empty output for nil report")
	}
}

func TestRenderTextWithShadow(t *testing.T) {
	shadowedVar := VarTrace{Name: "X", Value: "old", Origin: OriginTaskfileVars}
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Vars: []VarTrace{
					{Name: "X", Value: "new", Origin: OriginTaskVars, Type: "string",
						ShadowsVar: &shadowedVar},
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
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Vars: []VarTrace{
					{Name: "HOST", Value: "myhost", Origin: OriginTaskfileVars,
						Type: "string", IsDynamic: true},
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
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Vars: []VarTrace{
					{Name: "ALIAS", Value: "val", Origin: OriginIncludeVars,
						Type: "string", IsRef: true, RefName: "ORIGINAL"},
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
	// Unchanged cmd should show inline, not raw/resolved split
	if strings.Contains(output, "raw:") {
		t.Error("unchanged cmd should not show raw/resolved split")
	}
}

func TestRenderTextPipeSteps(t *testing.T) {
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
	if !strings.Contains(output, "pipe[0]") {
		t.Error("expected pipe step output")
	}
	if !strings.Contains(output, "trim") {
		t.Error("expected trim in pipe steps")
	}
}

func TestVarOriginString(t *testing.T) {
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
	o := VarOrigin(999)
	s := o.String()
	if !strings.Contains(s, "unknown") {
		t.Errorf("expected unknown for invalid origin, got %s", s)
	}
}

// ── Pipe Analyzer Tests ──

func TestAnalyzePipesSimple(t *testing.T) {
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
	funcs := template.FuncMap{}
	data := map[string]any{"FOO": "bar"}
	steps := AnalyzePipes("{{.FOO}}", data, funcs)
	if len(steps) != 0 {
		t.Fatalf("expected 0 pipe steps for single-command template, got %d", len(steps))
	}
}

func TestAnalyzePipesPlainText(t *testing.T) {
	funcs := template.FuncMap{}
	data := map[string]any{}
	steps := AnalyzePipes("no templates here", data, funcs)
	if len(steps) != 0 {
		t.Fatalf("expected 0 pipe steps for plain text, got %d", len(steps))
	}
}

func TestAnalyzePipesResolveArgs(t *testing.T) {
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
	warnings := DetectTypeMismatches("just plain text", nil, defaultFuncs())
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for plain text, got: %v", warnings)
	}
}

func TestDetectTypeMismatchesMulWithString(t *testing.T) {
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
	data := map[string]any{
		"NAME": "hello",
	}
	warnings := DetectTypeMismatches("{{upper .NAME}}", data, defaultFuncs())
	if len(warnings) != 0 {
		t.Errorf("expected no warnings for non-numeric func, got: %v", warnings)
	}
}
