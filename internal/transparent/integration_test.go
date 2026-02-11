package transparent_test

import (
	"context"
	"path/filepath"
	"runtime"
	"testing"

	task "github.com/go-task/task/v3"
	"github.com/go-task/task/v3/internal/transparent"
)

func examplesDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "..", "..", "docs", "transparent-mode", "examples")
}

func setupExecutor(t *testing.T, dir string) *task.Executor {
	t.Helper()
	e := task.NewExecutor(
		task.WithDir(dir),
		task.WithTransparent(true),
	)
	if err := e.Setup(); err != nil {
		t.Fatalf("setup failed: %v", err)
	}
	return e
}

func TestIntegrationBasicVariables(t *testing.T) {
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(), &task.Call{Task: "default"})
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	if len(report.Tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(report.Tasks))
	}

	task := report.Tasks[0]
	if task.TaskName != "default" {
		t.Errorf("expected task name default, got %s", task.TaskName)
	}

	// Check that APP_NAME, VERSION, ENV are present
	found := map[string]bool{}
	for _, v := range task.Vars {
		if v.Name == "APP_NAME" || v.Name == "VERSION" || v.Name == "ENV" {
			found[v.Name] = true
		}
	}
	for _, name := range []string{"APP_NAME", "VERSION", "ENV"} {
		if !found[name] {
			t.Errorf("expected variable %s in trace", name)
		}
	}

	// Check commands were traced
	if len(task.Cmds) < 3 {
		t.Errorf("expected at least 3 commands, got %d", len(task.Cmds))
	}

	// Check template traces
	if len(task.Templates) < 3 {
		t.Errorf("expected at least 3 template evaluations, got %d", len(task.Templates))
	}
}

func TestIntegrationVariableShadowing(t *testing.T) {
	dir := filepath.Join(examplesDir(), "02-variable-shadowing")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(), &task.Call{Task: "override"})
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	taskTrace := report.Tasks[0]

	// Find the NAME var and verify it shadows the global
	var nameVar *transparent.VarTrace
	for i := range taskTrace.Vars {
		if taskTrace.Vars[i].Name == "NAME" && taskTrace.Vars[i].Origin == transparent.OriginTaskVars {
			nameVar = &taskTrace.Vars[i]
			break
		}
	}
	if nameVar == nil {
		t.Fatal("expected NAME task var")
	}
	if nameVar.ShadowsVar == nil {
		t.Fatal("expected NAME to shadow a global var")
	}
	if nameVar.Value != "task-override" {
		t.Errorf("expected value task-override, got %v", nameVar.Value)
	}
	if nameVar.ShadowsVar.Value != "global-value" {
		t.Errorf("expected shadowed value global-value, got %v", nameVar.ShadowsVar.Value)
	}
}

func TestIntegrationTemplatePipes(t *testing.T) {
	dir := filepath.Join(examplesDir(), "03-template-pipes")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(), &task.Call{Task: "trim-pipe"})
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	taskTrace := report.Tasks[0]

	if len(taskTrace.Templates) < 2 {
		t.Fatalf("expected at least 2 template traces, got %d", len(taskTrace.Templates))
	}

	// First template: .NAME | trim => "world"
	tmpl0 := taskTrace.Templates[0]
	if tmpl0.Output != "echo 'world'" {
		t.Errorf("expected trimmed output 'echo 'world'', got %q", tmpl0.Output)
	}

	// Commands should show raw vs resolved
	if len(taskTrace.Cmds) < 2 {
		t.Fatalf("expected at least 2 cmds, got %d", len(taskTrace.Cmds))
	}
	if taskTrace.Cmds[0].RawCmd == taskTrace.Cmds[0].ResolvedCmd {
		t.Error("expected raw != resolved for template command")
	}
}

func TestIntegrationDynamicVariables(t *testing.T) {
	dir := filepath.Join(examplesDir(), "04-dynamic-variables")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(), &task.Call{Task: "default"})
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	taskTrace := report.Tasks[0]

	// Find HOSTNAME and verify it's marked dynamic
	var hostVar *transparent.VarTrace
	for i := range taskTrace.Vars {
		if taskTrace.Vars[i].Name == "HOSTNAME" {
			hostVar = &taskTrace.Vars[i]
			break
		}
	}
	if hostVar == nil {
		t.Fatal("expected HOSTNAME var in trace")
	}
	if !hostVar.IsDynamic {
		t.Error("expected HOSTNAME to be marked as dynamic (sh:)")
	}
	if hostVar.Value != "test-host" {
		t.Errorf("expected value test-host, got %v", hostVar.Value)
	}
}

func TestIntegrationMultipleTasks(t *testing.T) {
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(),
		&task.Call{Task: "default"},
		&task.Call{Task: "with-task-vars"},
	)
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	if len(report.Tasks) < 2 {
		t.Fatalf("expected at least 2 tasks, got %d", len(report.Tasks))
	}
}
