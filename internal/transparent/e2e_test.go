package transparent_test

import (
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
	assertContains(t, output, "shadows")
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

func truncateOutput(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "\n... (truncated)"
}
