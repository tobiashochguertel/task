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

// allVars returns all vars from both global and task scope for searching.
func allVars(report *transparent.TraceReport, taskIdx int) []transparent.VarTrace {
	var all []transparent.VarTrace
	all = append(all, report.GlobalVars...)
	if taskIdx < len(report.Tasks) {
		all = append(all, report.Tasks[taskIdx].Vars...)
	}
	return all
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

	// Check that APP_NAME, VERSION, ENV are present (in global or task vars)
	vars := allVars(report, 0)
	found := map[string]bool{}
	for _, v := range vars {
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

	// Find HOSTNAME in both global and task vars
	vars := allVars(report, 0)
	var hostVar *transparent.VarTrace
	for i := range vars {
		if vars[i].Name == "HOSTNAME" {
			hostVar = &vars[i]
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

	// Check dotenv vars are present (may be in global or task vars)
	vars := allVars(report, 0)
	found := map[string]bool{}
	for _, v := range vars {
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

	// Find ALIAS in both global and task vars
	vars := allVars(report, 0)
	var aliasVar *transparent.VarTrace
	for i := range vars {
		if vars[i].Name == "ALIAS" && vars[i].IsRef {
			aliasVar = &vars[i]
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

	// Verify include-vars from multiple levels (may be in global or task vars)
	vars := allVars(report, 0)
	var l1ToL2 *transparent.VarTrace
	var parentVar *transparent.VarTrace
	for i := range vars {
		switch vars[i].Name {
		case "L1_TO_L2":
			if vars[i].Origin == transparent.OriginIncludeVars {
				l1ToL2 = &vars[i]
			}
		case "PARENT_VAR":
			if vars[i].Origin == transparent.OriginIncludeVars {
				parentVar = &vars[i]
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

	// APP_ENV is in both env: and vars: — vars wins, should shadow
	// After separation, global vars include taskfile-vars
	vars := allVars(report, 0)
	var appEnvVar *transparent.VarTrace
	for i := range vars {
		if vars[i].Name == "APP_ENV" && vars[i].Origin == transparent.OriginTaskfileVars {
			appEnvVar = &vars[i]
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
