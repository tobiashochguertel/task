package transparent_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// getTaskBinary finds the task binary for E2E tests.
// It checks ./bin/task relative to the repo root first, then PATH.
func getTaskBinary(t *testing.T) string {
	t.Helper()
	_, file, _, _ := runtime.Caller(0)
	repoRoot := filepath.Join(filepath.Dir(file), "..", "..")
	binPath := filepath.Join(repoRoot, "bin", "task")
	if _, err := os.Stat(binPath); err == nil {
		return binPath
	}
	if path, err := exec.LookPath("task"); err == nil {
		return path
	}
	t.Skip("task binary not found — run 'go build -o ./bin/task ./cmd/task' first")
	return ""
}

func runTransparent(t *testing.T, dir string, taskName string) string {
	t.Helper()
	bin := getTaskBinary(t)
	args := []string{"--transparent", "-d", dir}
	if taskName != "" {
		args = append(args, taskName)
	}
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("task --transparent failed: %v\nOutput: %s", err, string(out))
	}
	return string(out)
}

func TestE2EBasicVariables(t *testing.T) {
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	output := runTransparent(t, dir, "default")

	// Verify report structure
	assertContains(t, output, "Transparent Mode Report")
	assertContains(t, output, "Task: default")
	assertContains(t, output, "Variables:")

	// Verify variable values are shown
	assertContains(t, output, "APP_NAME")
	assertContains(t, output, "my-app")
	assertContains(t, output, "VERSION")
	assertContains(t, output, "1.0.0")

	// Verify template evaluations
	assertContains(t, output, "Template Evaluations:")
	assertContains(t, output, "{{.APP_NAME}}")
	assertContains(t, output, "echo 'App: my-app'")

	// Verify commands
	assertContains(t, output, "Commands:")
	assertContains(t, output, "resolved:")
}

func TestE2EVariableShadowing(t *testing.T) {
	dir := filepath.Join(examplesDir(), "02-variable-shadowing")

	// Test task that overrides a global variable
	output := runTransparent(t, dir, "override")

	// Should show shadow warning
	assertContains(t, output, "SHADOWS")
	assertContains(t, output, "task-override")

	// Verify the global var is still visible
	assertContains(t, output, "global-value")

	// Should show resolved command with overridden value
	assertContains(t, output, "Hello task-override")
}

func TestE2EVariableShadowingDefault(t *testing.T) {
	dir := filepath.Join(examplesDir(), "02-variable-shadowing")

	// Default task uses global NAME — no shadow expected
	output := runTransparent(t, dir, "default")
	assertContains(t, output, "global-value")
	assertContains(t, output, "Hello global-value")
}

func TestE2ETemplatePipeTrim(t *testing.T) {
	dir := filepath.Join(examplesDir(), "03-template-pipes")
	output := runTransparent(t, dir, "trim-pipe")

	// Verify trimmed output
	assertContains(t, output, "echo 'world'")
	assertContains(t, output, "NAME:world")
}

func TestE2ETemplatePipeUpper(t *testing.T) {
	dir := filepath.Join(examplesDir(), "03-template-pipes")
	output := runTransparent(t, dir, "upper-lower")

	assertContains(t, output, "HELLO WORLD")
	assertContains(t, output, "Hello World")
}

func TestE2ETemplatePipeCombined(t *testing.T) {
	dir := filepath.Join(examplesDir(), "03-template-pipes")
	output := runTransparent(t, dir, "combined-pipes")

	assertContains(t, output, "WORLD")
	assertContains(t, output, "APP_WORLD")
}

func TestE2EDynamicVariables(t *testing.T) {
	dir := filepath.Join(examplesDir(), "04-dynamic-variables")
	output := runTransparent(t, dir, "default")

	// Static var
	assertContains(t, output, "static-value")

	// Dynamic vars resolved via sh:
	assertContains(t, output, "test-host")
	assertContains(t, output, "2026-02-11")

	// Dynamic markers
	assertContains(t, output, "(sh)")
}

func TestE2EDynamicVariablesTaskLevel(t *testing.T) {
	dir := filepath.Join(examplesDir(), "04-dynamic-variables")
	output := runTransparent(t, dir, "task-dynamic")

	assertContains(t, output, "task-123")
}

func TestE2EIncludedTaskfile(t *testing.T) {
	dir := filepath.Join(examplesDir(), "05-includes")
	output := runTransparent(t, dir, "sub:greet")

	// Variables from included taskfile should be visible
	assertContains(t, output, "from-included")
	assertContains(t, output, "from-parent")
}

func TestE2EAdvancedBuild(t *testing.T) {
	dir := filepath.Join(examplesDir(), "06-advanced-combined")
	output := runTransparent(t, dir, "build")

	// Computed variables
	assertContains(t, output, "myproject-2.0.0")
	assertContains(t, output, "myproject/2-0-0")

	// Template evaluation
	assertContains(t, output, "Building myproject-2.0.0")
}

func TestE2EAdvancedConditional(t *testing.T) {
	dir := filepath.Join(examplesDir(), "06-advanced-combined")
	output := runTransparent(t, dir, "conditional")

	assertContains(t, output, "Debug mode ON")
	assertContains(t, output, "myproject v2.0.0 built 2026-02-11")
}

func TestE2EDoesNotExecuteCommands(t *testing.T) {
	// Transparent mode should NOT execute any commands
	// Verify by checking there's no actual command output, only the report
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	output := runTransparent(t, dir, "default")

	// The report should exist
	assertContains(t, output, "Transparent Mode Report")
	assertContains(t, output, "End Report")

	// There should NOT be actual echo output outside the report
	// Count occurrences of "my-app" — should only appear in the report sections
	lines := strings.Split(output, "\n")
	echoCount := 0
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		// If line starts with "App:" that would be actual command output
		if strings.HasPrefix(trimmed, "App: my-app") {
			echoCount++
		}
	}
	if echoCount > 0 {
		t.Error("transparent mode should not execute commands, but found actual echo output")
	}
}

func TestE2EShortFlagT(t *testing.T) {
	// Test -T short flag works the same as --transparent
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	cmd := exec.Command(bin, "-T", "-d", dir)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("task -T failed: %v\nOutput: %s", err, string(out))
	}
	assertContains(t, string(out), "Transparent Mode Report")
}

func TestE2EMultipleTasksOnCLI(t *testing.T) {
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "02-variable-shadowing")
	cmd := exec.Command(bin, "--transparent", "-d", dir, "default", "override")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("task --transparent with multiple tasks failed: %v\nOutput: %s", err, string(out))
	}
	output := string(out)
	// Both tasks should appear in the report
	if strings.Count(output, "Task:") < 2 {
		t.Error("expected at least 2 tasks in report when passing multiple task names")
	}
}

func assertContains(t *testing.T, output, substr string) {
	t.Helper()
	if !strings.Contains(output, substr) {
		t.Errorf("expected output to contain %q\nGot:\n%s", substr, truncateOutput(output, 500))
	}
}

func assertNotContains(t *testing.T, output, substr string) {
	t.Helper()
	if strings.Contains(output, substr) {
		t.Errorf("expected output NOT to contain %q\nGot:\n%s", substr, truncateOutput(output, 500))
	}
}

func truncateOutput(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n... (truncated)"
}

// --- New E2E tests for expanded examples ---

func TestE2EDotenv(t *testing.T) {
	dir := filepath.Join(examplesDir(), "07-dotenv")
	output := runTransparent(t, dir, "default")

	assertContains(t, output, "Transparent Mode Report")
	assertContains(t, output, "Task: default")
	assertContains(t, output, "DB_HOST")
	assertContains(t, output, "localhost")
	assertContains(t, output, "DB_PORT")
	assertContains(t, output, "5432")
	assertContains(t, output, "APP_NAME")
	assertContains(t, output, "my-dotenv-app")
}

func TestE2EDotenvCombined(t *testing.T) {
	dir := filepath.Join(examplesDir(), "07-dotenv")
	output := runTransparent(t, dir, "combined")

	assertContains(t, output, "CONNECTION")
	assertContains(t, output, "localhost:5432/mydb")
}

func TestE2EPreconditions(t *testing.T) {
	dir := filepath.Join(examplesDir(), "08-preconditions")
	output := runTransparent(t, dir, "check-binary")

	assertContains(t, output, "Transparent Mode Report")
	assertContains(t, output, "APP_NAME")
	assertContains(t, output, "my-app")
	assertContains(t, output, "Commands:")
}

func TestE2EPreconditionsTemplate(t *testing.T) {
	dir := filepath.Join(examplesDir(), "08-preconditions")
	output := runTransparent(t, dir, "check-version")

	assertContains(t, output, "VERSION")
	assertContains(t, output, "1.0.0")
}

func TestE2ETemplateFieldsLabel(t *testing.T) {
	dir := filepath.Join(examplesDir(), "09-template-fields")
	output := runTransparent(t, dir, "with-label")

	assertContains(t, output, "APP_NAME")
	assertContains(t, output, "my-app")
	assertContains(t, output, "VERSION")
	assertContains(t, output, "2.0.0")
	assertContains(t, output, "Building my-app")
}

func TestE2ETemplateFieldsSummary(t *testing.T) {
	dir := filepath.Join(examplesDir(), "09-template-fields")
	output := runTransparent(t, dir, "with-summary")

	assertContains(t, output, "Deploying my-app v2.0.0 to production")
}

func TestE2ETemplateFieldsDir(t *testing.T) {
	dir := filepath.Join(examplesDir(), "09-template-fields")
	output := runTransparent(t, dir, "with-dir")

	assertContains(t, output, "APP_NAME")
	assertContains(t, output, "my-app")
}

func TestE2ERefVariables(t *testing.T) {
	dir := filepath.Join(examplesDir(), "10-ref-variables")
	output := runTransparent(t, dir, "default")

	assertContains(t, output, "GREETING")
	assertContains(t, output, "hello-world")
	assertContains(t, output, "ALIAS")
	// Ref marker should appear
	assertContains(t, output, "ref")
	assertContains(t, output, ".GREETING")
}

func TestE2ERefVariablesTaskLevel(t *testing.T) {
	dir := filepath.Join(examplesDir(), "10-ref-variables")
	output := runTransparent(t, dir, "with-task-ref")

	assertContains(t, output, "LOCAL_GREETING")
	assertContains(t, output, "hello-world")
	assertContains(t, output, "ref")
}

func TestE2EStatusSources(t *testing.T) {
	dir := filepath.Join(examplesDir(), "11-status-sources")
	output := runTransparent(t, dir, "build")

	assertContains(t, output, "APP_NAME")
	assertContains(t, output, "my-app")
	assertContains(t, output, "VERSION")
	assertContains(t, output, "BUILD_DIR")
	assertContains(t, output, "Building my-app v1.0.0")
}

func TestE2EEnvVariables(t *testing.T) {
	dir := filepath.Join(examplesDir(), "12-env-variables")
	output := runTransparent(t, dir, "default")

	assertContains(t, output, "APP_NAME")
	assertContains(t, output, "my-app")
	// APP_ENV should show as taskfile-vars (vars override env)
	assertContains(t, output, "APP_ENV")
	assertContains(t, output, "production")
	// LOG_LEVEL from env: block
	assertContains(t, output, "LOG_LEVEL")
}

func TestE2EEnvVariablesShadowing(t *testing.T) {
	dir := filepath.Join(examplesDir(), "12-env-variables")
	output := runTransparent(t, dir, "default")

	// APP_ENV is defined in both env: and vars: — should show shadow
	assertContains(t, output, "SHADOWS")
}

func TestE2ENestedIncludesLevel1(t *testing.T) {
	dir := filepath.Join(examplesDir(), "13-nested-includes")
	output := runTransparent(t, dir, "level1:greet")

	assertContains(t, output, "Task: level1:greet")
	assertContains(t, output, "LEVEL1_VAR")
	assertContains(t, output, "from-level1")
	assertContains(t, output, "PARENT_VAR")
	assertContains(t, output, "from-root-include")
	assertContains(t, output, "include-vars")
}

func TestE2ENestedIncludesLevel2(t *testing.T) {
	dir := filepath.Join(examplesDir(), "13-nested-includes")
	output := runTransparent(t, dir, "level1:level2:greet")

	assertContains(t, output, "Task: level1:level2:greet")
	assertContains(t, output, "LEVEL2_VAR")
	assertContains(t, output, "from-level2")
	assertContains(t, output, "L1_TO_L2")
	assertContains(t, output, "from-level1-include")
	// Parent var propagates through
	assertContains(t, output, "PARENT_VAR")
	assertContains(t, output, "from-root-include")
}

func TestE2EMatrixFor(t *testing.T) {
	dir := filepath.Join(examplesDir(), "14-matrix-for")
	output := runTransparent(t, dir, "build-platforms")

	assertContains(t, output, "PROJECT")
	assertContains(t, output, "my-app")
	// For-loop should expand commands
	assertContains(t, output, "Commands:")
	assertContains(t, output, "Building my-app for linux")
	assertContains(t, output, "Building my-app for darwin")
	assertContains(t, output, "Building my-app for windows")
}

func TestE2EMatrixForInline(t *testing.T) {
	dir := filepath.Join(examplesDir(), "14-matrix-for")
	output := runTransparent(t, dir, "build-inline")

	assertContains(t, output, "Mode: debug for my-app")
	assertContains(t, output, "Mode: release for my-app")
}

func TestE2EMatrixForMatrix(t *testing.T) {
	dir := filepath.Join(examplesDir(), "14-matrix-for")
	output := runTransparent(t, dir, "build-matrix")

	// Matrix should produce platform x arch combinations
	assertContains(t, output, "my-app-linux-amd64")
	assertContains(t, output, "my-app-darwin-arm64")
	assertContains(t, output, "my-app-windows-amd64")
}

func TestE2EMultiLevelShadowChain(t *testing.T) {
	// Test that a variable shadowed through multiple scopes is detected
	dir := filepath.Join(examplesDir(), "12-env-variables")
	output := runTransparent(t, dir, "default")

	// APP_ENV defined in env: then overridden in vars:
	assertContains(t, output, "APP_ENV")
	assertContains(t, output, "SHADOWS")
}

func TestE2EColorFlagFalse(t *testing.T) {
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	cmd := exec.Command(bin, "--transparent", "--color=false", "-d", dir, "default")
	// Don't set NO_COLOR — rely on --color=false flag
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("task --transparent --color=false failed: %v\nOutput: %s", err, string(out))
	}
	output := string(out)
	assertContains(t, output, "Transparent Mode Report")
	// Should not contain raw ANSI escape codes
	assertNotContains(t, output, "\033[")
}

func TestE2EOutputToStderr(t *testing.T) {
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	cmd := exec.Command(bin, "--transparent", "-d", dir, "default")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("task --transparent failed: %v\nStderr: %s", err, stderr.String())
	}

	// Report should be on stderr, not stdout
	if !strings.Contains(stderr.String(), "Transparent Mode Report") {
		t.Error("expected report on stderr")
	}
	if strings.Contains(stdout.String(), "Transparent Mode Report") {
		t.Error("expected report NOT on stdout")
	}
}

// ── Pipe Analyzer E2E Tests ──

func TestE2EPipeStepsTrim(t *testing.T) {
	dir := filepath.Join(examplesDir(), "03-template-pipes")
	output := runTransparent(t, dir, "trim-pipe")

	// Should show pipe step breakdown
	assertContains(t, output, "pipe[0]")
	assertContains(t, output, "pipe[1]")
	assertContains(t, output, ".NAME")
	assertContains(t, output, "trim")
}

func TestE2EPipeStepsCombined(t *testing.T) {
	dir := filepath.Join(examplesDir(), "03-template-pipes")
	output := runTransparent(t, dir, "combined-pipes")

	// Should show 3-step pipe: .NAME | trim | upper
	assertContains(t, output, "pipe[0]")
	assertContains(t, output, "pipe[1]")
	assertContains(t, output, "pipe[2]")
	assertContains(t, output, "upper")
	assertContains(t, output, "WORLD")
}

func TestE2EPipeStepsUpperLower(t *testing.T) {
	dir := filepath.Join(examplesDir(), "03-template-pipes")
	output := runTransparent(t, dir, "upper-lower")

	// Should have pipe steps for upper and lower
	assertContains(t, output, "pipe[0]")
	assertContains(t, output, "pipe[1]")
}

// ── JSON Output E2E Tests ──

func TestE2EJSONOutput(t *testing.T) {
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	cmd := exec.Command(bin, "--transparent", "--json", "-d", dir, "default")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("task --transparent --json failed: %v\n%s", err, out)
	}

	// Should be valid JSON
	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON: %v\nOutput: %s", err, out)
	}
	tasks, ok := result["tasks"].([]any)
	if !ok || len(tasks) == 0 {
		t.Fatal("expected 'tasks' array in JSON output")
	}
	task0, ok := tasks[0].(map[string]any)
	if !ok {
		t.Fatal("expected task to be an object")
	}
	if task0["name"] != "default" {
		t.Errorf("expected task name 'default', got %v", task0["name"])
	}
}

func TestE2EJSONPipeSteps(t *testing.T) {
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "03-template-pipes")
	cmd := exec.Command(bin, "--transparent", "--json", "-d", dir, "trim-pipe")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("task --transparent --json failed: %v\n%s", err, out)
	}

	var result map[string]any
	if err := json.Unmarshal(out, &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	tasks := result["tasks"].([]any)
	task0 := tasks[0].(map[string]any)
	templates := task0["templates"].([]any)
	// First template should have pipe_steps
	tmpl0 := templates[0].(map[string]any)
	steps, ok := tmpl0["pipe_steps"].([]any)
	if !ok || len(steps) < 2 {
		t.Fatalf("expected at least 2 pipe_steps, got %v", tmpl0["pipe_steps"])
	}
	step0 := steps[0].(map[string]any)
	if step0["func"] != ".NAME" {
		t.Errorf("expected pipe step func '.NAME', got %v", step0["func"])
	}
}

// ── Undefined Variable Warning E2E Tests ──

func TestE2EUndefinedVarWarning(t *testing.T) {
	dir := filepath.Join(examplesDir(), "15-undefined-vars")
	output := runTransparent(t, dir, "test-undefined")

	assertContains(t, output, "warning")
	assertContains(t, output, "<no value>")
}

func TestE2ENoWarningForDefinedVars(t *testing.T) {
	dir := filepath.Join(examplesDir(), "15-undefined-vars")
	output := runTransparent(t, dir, "test-all-defined")

	assertNotContains(t, output, "warning")
	assertNotContains(t, output, "<no value>")
}

// --- Round 2 gap-closure tests ---

func TestE2EGlobalVarsSection(t *testing.T) {
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	output := runTransparent(t, dir, "default")

	// Should have a Global Variables section separate from task
	assertContains(t, output, "Global Variables")
	assertContains(t, output, "Task: default")

	// Global vars should contain special + taskfile-level vars
	assertContains(t, output, "TASK_VERSION")
	assertContains(t, output, "special")
	assertContains(t, output, "APP_NAME")
	assertContains(t, output, "taskfile-vars")
}

func TestE2EGlobalVarsSeparation(t *testing.T) {
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	output := runTransparent(t, dir, "default")

	// Split output at "Task: default"
	parts := strings.SplitN(output, "Task: default", 2)
	if len(parts) < 2 {
		t.Fatal("expected Task: default section")
	}
	globalSection := parts[0]
	taskSection := parts[1]

	// APP_NAME should be in global section (taskfile-vars), not task section
	if !strings.Contains(globalSection, "APP_NAME") {
		t.Error("expected APP_NAME in global section")
	}
	// APP_NAME should NOT appear in task-level vars
	// (it may appear in template evaluations though)
	taskVarsEnd := strings.Index(taskSection, "Template Evaluations:")
	if taskVarsEnd > 0 {
		taskVarsSection := taskSection[:taskVarsEnd]
		lines := strings.Split(taskVarsSection, "\n")
		for _, line := range lines {
			if strings.Contains(line, "APP_NAME") && strings.Contains(line, "taskfile-vars") {
				t.Error("APP_NAME should not be in task-level vars section")
			}
		}
	}
}

func TestE2EShadowWarningFormat(t *testing.T) {
	dir := filepath.Join(examplesDir(), "02-variable-shadowing")
	output := runTransparent(t, dir, "override")

	// New format: ⚠ SHADOWS NAME="value" [origin]
	assertContains(t, output, "SHADOWS")
	assertContains(t, output, "global-value")
	assertContains(t, output, "taskfile-vars")
}

func TestE2EDynamicVarShellCommand(t *testing.T) {
	dir := filepath.Join(examplesDir(), "04-dynamic-variables")
	output := runTransparent(t, dir, "default")

	// Should show the shell command for dynamic vars
	assertContains(t, output, "(sh)")
	assertContains(t, output, "echo")
}

func TestE2EListAllTransparent(t *testing.T) {
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	cmd := exec.Command(bin, "--transparent", "--list-all", "-d", dir)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--transparent --list-all failed: %v\nOutput: %s", err, string(out))
	}
	output := string(out)

	// Should show ALL tasks, not just default
	assertContains(t, output, "Task: default")
	assertContains(t, output, "Task: with-task-vars")
}

func TestE2ETemplateContext(t *testing.T) {
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	output := runTransparent(t, dir, "default")

	// Template evaluations should show context labels
	assertContains(t, output, "cmds[0]")
	assertContains(t, output, "cmds[1]")
}

func TestE2EJSONGlobalVars(t *testing.T) {
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	cmd := exec.Command(bin, "--transparent", "--json", "-d", dir, "default")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--transparent --json failed: %v\nOutput: %s", err, string(out))
	}

	var report map[string]any
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check version field
	if v, ok := report["version"]; !ok || v != "1.0" {
		t.Errorf("expected version 1.0, got %v", v)
	}

	// Check global_vars array exists and has entries
	gv, ok := report["global_vars"]
	if !ok {
		t.Fatal("expected global_vars in JSON output")
	}
	gvArr, ok := gv.([]any)
	if !ok || len(gvArr) == 0 {
		t.Fatal("expected non-empty global_vars array")
	}
}

func TestE2EJSONTemplateContext(t *testing.T) {
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	cmd := exec.Command(bin, "--transparent", "--json", "-d", dir, "default")
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("--transparent --json failed: %v\nOutput: %s", err, string(out))
	}

	var report map[string]any
	if err := json.Unmarshal(out, &report); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	// Check templates have context field
	tasks := report["tasks"].([]any)
	task := tasks[0].(map[string]any)
	templates := task["templates"].([]any)
	if len(templates) == 0 {
		t.Fatal("expected templates")
	}
	tmpl := templates[0].(map[string]any)
	ctx, ok := tmpl["context"]
	if !ok || ctx == "" {
		t.Error("expected non-empty context field in template trace")
	}
}
