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

func TestIntegrationDotenv(t *testing.T) {
	dir := filepath.Join(examplesDir(), "07-dotenv")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(), &task.Call{Task: "default"})
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	taskTrace := report.Tasks[0]

	// Check dotenv vars are present
	found := map[string]bool{}
	for _, v := range taskTrace.Vars {
		if v.Name == "DB_HOST" || v.Name == "DB_PORT" || v.Name == "DB_NAME" {
			found[v.Name] = true
		}
	}
	for _, name := range []string{"DB_HOST", "DB_PORT", "DB_NAME"} {
		if !found[name] {
			t.Errorf("expected dotenv variable %s in trace", name)
		}
	}
}

func TestIntegrationRefVariables(t *testing.T) {
	dir := filepath.Join(examplesDir(), "10-ref-variables")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(), &task.Call{Task: "default"})
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	taskTrace := report.Tasks[0]

	// Find ALIAS and verify it has IsRef set
	var aliasVar *transparent.VarTrace
	for i := range taskTrace.Vars {
		if taskTrace.Vars[i].Name == "ALIAS" && taskTrace.Vars[i].IsRef {
			aliasVar = &taskTrace.Vars[i]
			break
		}
	}
	if aliasVar == nil {
		t.Fatal("expected ALIAS var with IsRef=true")
	}
	if aliasVar.RefName != ".GREETING" {
		t.Errorf("expected RefName=.GREETING, got %s", aliasVar.RefName)
	}
}

func TestIntegrationNestedIncludes(t *testing.T) {
	dir := filepath.Join(examplesDir(), "13-nested-includes")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(), &task.Call{Task: "level1:level2:greet"})
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	taskTrace := report.Tasks[0]

	// Verify include-vars from multiple levels
	var l1ToL2 *transparent.VarTrace
	var parentVar *transparent.VarTrace
	for i := range taskTrace.Vars {
		switch taskTrace.Vars[i].Name {
		case "L1_TO_L2":
			if taskTrace.Vars[i].Origin == transparent.OriginIncludeVars {
				l1ToL2 = &taskTrace.Vars[i]
			}
		case "PARENT_VAR":
			if taskTrace.Vars[i].Origin == transparent.OriginIncludeVars {
				parentVar = &taskTrace.Vars[i]
			}
		}
	}
	if l1ToL2 == nil {
		t.Fatal("expected L1_TO_L2 include-vars")
	}
	if l1ToL2.Value != "from-level1-include" {
		t.Errorf("expected L1_TO_L2=from-level1-include, got %v", l1ToL2.Value)
	}
	if parentVar == nil {
		t.Fatal("expected PARENT_VAR propagated through levels")
	}
	if parentVar.Value != "from-root-include" {
		t.Errorf("expected PARENT_VAR=from-root-include, got %v", parentVar.Value)
	}
}

func TestIntegrationEnvVarsShadow(t *testing.T) {
	dir := filepath.Join(examplesDir(), "12-env-variables")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(), &task.Call{Task: "default"})
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	taskTrace := report.Tasks[0]

	// APP_ENV is in both env: and vars: — vars wins, should shadow
	var appEnvVar *transparent.VarTrace
	for i := range taskTrace.Vars {
		if taskTrace.Vars[i].Name == "APP_ENV" && taskTrace.Vars[i].Origin == transparent.OriginTaskfileVars {
			appEnvVar = &taskTrace.Vars[i]
			break
		}
	}
	if appEnvVar == nil {
		t.Fatal("expected APP_ENV taskfile-vars")
	}
	if appEnvVar.ShadowsVar == nil {
		t.Fatal("expected APP_ENV to shadow the env: version")
	}
	if appEnvVar.Value != "production" {
		t.Errorf("expected production, got %v", appEnvVar.Value)
	}
}

func TestIntegrationMatrixFor(t *testing.T) {
	dir := filepath.Join(examplesDir(), "14-matrix-for")
	e := setupExecutor(t, dir)

	err := e.RunTransparent(context.Background(), &task.Call{Task: "build-platforms"})
	if err != nil {
		t.Fatalf("RunTransparent failed: %v", err)
	}

	report := e.Compiler.Tracer.Report()
	taskTrace := report.Tasks[0]

	// Should have 3 expanded commands (linux, darwin, windows)
	if len(taskTrace.Cmds) < 3 {
		t.Fatalf("expected at least 3 for-loop expanded cmds, got %d", len(taskTrace.Cmds))
	}
}
