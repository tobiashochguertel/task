package transparent_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"
)

// updateGolden can be set via UPDATE_GOLDEN=1 env to regenerate golden files.
var updateGolden = os.Getenv("UPDATE_GOLDEN") == "1"

// goldenDir returns the path to the golden testdata directory.
func goldenDir() string {
	_, file, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(file), "testdata", "golden")
}

// normalizeText strips machine-specific paths, pointers and versions from text output.
func normalizeText(s string) string {
	s = regexp.MustCompile(`/Users/\S+`).ReplaceAllString(s, "<PATH>")
	s = regexp.MustCompile(`/var/folders/\S+`).ReplaceAllString(s, "<PATH>")
	s = regexp.MustCompile(`/tmp/\S+`).ReplaceAllString(s, "<PATH>")
	s = regexp.MustCompile(`\./bin/task\S*`).ReplaceAllString(s, "<PATH>")
	s = regexp.MustCompile(`ptr: 0x[0-9a-f]+`).ReplaceAllString(s, "ptr: <PTR>")
	s = regexp.MustCompile(`3\.\d+\.\d+`).ReplaceAllString(s, "<TASK_VERSION>")
	return s
}

// normalizeJSON normalizes a JSON byte slice by replacing volatile values.
func normalizeJSON(data []byte) ([]byte, error) {
	var obj any
	if err := json.Unmarshal(data, &obj); err != nil {
		return nil, err
	}
	normalized := normalizeJSONValue(obj)
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(normalized); err != nil {
		return nil, err
	}
	// Encode adds trailing newline; trim it for consistency
	return bytes.TrimRight(buf.Bytes(), "\n"), nil
}

func normalizeJSONValue(v any) any {
	switch val := v.(type) {
	case string:
		s := val
		s = regexp.MustCompile(`/Users/\S+`).ReplaceAllString(s, "<PATH>")
		s = regexp.MustCompile(`/var/folders/\S+`).ReplaceAllString(s, "<PATH>")
		s = regexp.MustCompile(`/tmp/\S+`).ReplaceAllString(s, "<PATH>")
		s = regexp.MustCompile(`\./bin/task\S*`).ReplaceAllString(s, "<PATH>")
		s = regexp.MustCompile(`3\.\d+\.\d+`).ReplaceAllString(s, "<TASK_VERSION>")
		return s
	case map[string]any:
		result := make(map[string]any, len(val))
		for k, v := range val {
			switch k {
			case "value_id":
				result[k] = float64(0)
			case "ptr":
				result[k] = "<PTR>"
			default:
				result[k] = normalizeJSONValue(v)
			}
		}
		return result
	case []any:
		result := make([]any, len(val))
		for i, item := range val {
			result[i] = normalizeJSONValue(item)
		}
		return result
	default:
		return v
	}
}

// goldenTestCases defines all example/task combinations to test.
type goldenTestCase struct {
	Example string
	Task    string
}

func allGoldenTestCases() []goldenTestCase {
	return []goldenTestCase{
		{"01-basic-variables", "default"},
		{"01-basic-variables", "with-task-vars"},
		{"02-variable-shadowing", "default"},
		{"02-variable-shadowing", "override"},
		{"02-variable-shadowing", "double-override"},
		{"03-template-pipes", "trim-pipe"},
		{"03-template-pipes", "upper-lower"},
		{"03-template-pipes", "combined-pipes"},
		{"03-template-pipes", "printf-basic"},
		{"03-template-pipes", "quote-pipe"},
		{"04-dynamic-variables", "default"},
		{"04-dynamic-variables", "task-dynamic"},
		{"05-includes", "default"},
		{"05-includes", "sub:greet"},
		{"06-advanced-combined", "build"},
		{"06-advanced-combined", "conditional"},
		{"06-advanced-combined", "deploy"},
		{"06-advanced-combined", "matrix"},
		{"07-dotenv", "default"},
		{"07-dotenv", "combined"},
		{"08-preconditions", "check-binary"},
		{"08-preconditions", "check-version"},
		{"09-template-fields", "with-label"},
		{"09-template-fields", "with-summary"},
		{"09-template-fields", "with-dir"},
		{"09-template-fields", "with-prefix"},
		{"10-ref-variables", "default"},
		{"10-ref-variables", "with-task-ref"},
		{"11-status-sources", "build"},
		{"12-env-variables", "default"},
		{"12-env-variables", "with-task-env"},
		{"13-nested-includes", "level1:greet"},
		{"13-nested-includes", "level1:level2:greet"},
		{"14-matrix-for", "build-platforms"},
		{"14-matrix-for", "build-inline"},
		{"14-matrix-for", "build-matrix"},
		{"15-undefined-vars", "test-undefined"},
		{"15-undefined-vars", "test-all-defined"},
	}
}

// safeTaskName converts task names like "sub:greet" to "sub_greet" for filenames.
func safeTaskName(task string) string {
	return strings.ReplaceAll(task, ":", "_")
}

// TestGoldenText verifies text output of --transparent matches golden files.
func TestGoldenText(t *testing.T) {
	bin := getTaskBinary(t)

	for _, tc := range allGoldenTestCases() {
		name := fmt.Sprintf("%s/%s", tc.Example, tc.Task)
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(examplesDir(), tc.Example)
			cmd := exec.Command(bin, "--transparent", "-d", dir, tc.Task)
			cmd.Env = append(os.Environ(), "NO_COLOR=1")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("transparent failed: %v\n%s", err, out)
			}

			actual := normalizeText(string(out))
			goldenFile := filepath.Join(goldenDir(), fmt.Sprintf("%s__%s.golden", tc.Example, safeTaskName(tc.Task)))

			if updateGolden {
				if err := os.WriteFile(goldenFile, []byte(actual), 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			expected, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("golden file not found: %v (run with UPDATE_GOLDEN=1 to create)", err)
			}
			if actual != string(expected) {
				t.Errorf("output mismatch for %s\n--- golden ---\n%s\n--- actual ---\n%s",
					name, string(expected), actual)
			}
		})
	}
}

// TestGoldenJSON verifies JSON output of --transparent --json matches golden files.
func TestGoldenJSON(t *testing.T) {
	bin := getTaskBinary(t)

	for _, tc := range allGoldenTestCases() {
		name := fmt.Sprintf("%s/%s", tc.Example, tc.Task)
		t.Run(name, func(t *testing.T) {
			dir := filepath.Join(examplesDir(), tc.Example)
			cmd := exec.Command(bin, "--transparent", "--json", "-d", dir, tc.Task)
			cmd.Env = append(os.Environ(), "NO_COLOR=1")
			out, err := cmd.CombinedOutput()
			if err != nil {
				t.Fatalf("transparent --json failed: %v\n%s", err, out)
			}

			normalizedActual, err := normalizeJSON(out)
			if err != nil {
				t.Fatalf("invalid JSON output: %v\n%s", err, out)
			}
			actual := string(normalizedActual) + "\n"

			goldenFile := filepath.Join(goldenDir(), fmt.Sprintf("%s__%s.golden.json", tc.Example, safeTaskName(tc.Task)))

			if updateGolden {
				if err := os.WriteFile(goldenFile, []byte(actual), 0644); err != nil {
					t.Fatalf("failed to write golden file: %v", err)
				}
				return
			}

			expectedRaw, err := os.ReadFile(goldenFile)
			if err != nil {
				t.Fatalf("golden file not found: %v (run with UPDATE_GOLDEN=1 to create)", err)
			}
			// Re-normalize the golden file through the same pipeline
			normalizedExpected, err := normalizeJSON(expectedRaw)
			if err != nil {
				t.Fatalf("invalid golden JSON: %v", err)
			}
			expected := string(normalizedExpected) + "\n"

			if actual != expected {
				t.Errorf("JSON output mismatch for %s\n--- golden ---\n%s\n--- actual ---\n%s",
					name, expected, actual)
			}
		})
	}
}

// TestGoldenListAll verifies --transparent --list-all output.
func TestGoldenListAll(t *testing.T) {
	bin := getTaskBinary(t)
	dir := filepath.Join(examplesDir(), "01-basic-variables")
	cmd := exec.Command(bin, "--transparent", "--list-all", "-d", dir)
	cmd.Env = append(os.Environ(), "NO_COLOR=1")
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("transparent --list-all failed: %v\n%s", err, out)
	}

	output := normalizeText(string(out))
	// Must contain both tasks
	if !strings.Contains(output, "Task: default") {
		t.Error("missing Task: default in --list-all output")
	}
	if !strings.Contains(output, "Task: with-task-vars") {
		t.Error("missing Task: with-task-vars in --list-all output")
	}
	// Must have global vars section
	if !strings.Contains(output, "Global Variables") {
		t.Error("missing Global Variables section in --list-all output")
	}
}
