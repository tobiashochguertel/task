package transparent

import (
	"bytes"
	"strings"
	"testing"
)

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
	RenderText(&buf, report)
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
