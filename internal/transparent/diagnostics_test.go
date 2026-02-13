package transparent

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/go-task/template"
)

// ── AnalyzeEvalActions Tests ──

func TestAnalyzeEvalActionsSingleAction(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"trim": strings.TrimSpace,
	}
	data := map[string]any{"NAME": "  world  "}
	actions := AnalyzeEvalActions("{{.NAME | trim}}", data, funcs)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	ea := actions[0]
	if ea.ActionIndex != 0 {
		t.Errorf("ActionIndex = %d, want 0", ea.ActionIndex)
	}
	if ea.SourceLine != 1 {
		t.Errorf("SourceLine = %d, want 1", ea.SourceLine)
	}
	if ea.Source != "{{.NAME | trim}}" {
		t.Errorf("Source = %q, want %q", ea.Source, "{{.NAME | trim}}")
	}
	if ea.Result != "world" {
		t.Errorf("Result = %q, want %q", ea.Result, "world")
	}
	if len(ea.Steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(ea.Steps))
	}
	// Step 1: Resolve .NAME
	if ea.Steps[0].Operation != "Resolve a Variable" {
		t.Errorf("step 0 operation = %q, want Resolve a Variable", ea.Steps[0].Operation)
	}
	if ea.Steps[0].Target != ".NAME" {
		t.Errorf("step 0 target = %q, want .NAME", ea.Steps[0].Target)
	}
	// Step 2: Apply trim
	if ea.Steps[1].Operation != "Apply a Function" {
		t.Errorf("step 1 operation = %q, want Apply a Function", ea.Steps[1].Operation)
	}
	if ea.Steps[1].Target != "trim" {
		t.Errorf("step 1 target = %q, want trim", ea.Steps[1].Target)
	}
	if ea.Steps[1].Output != "world" {
		t.Errorf("step 1 output = %q, want world", ea.Steps[1].Output)
	}
}

func TestAnalyzeEvalActionsMultiAction(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"upper": strings.ToUpper,
	}
	data := map[string]any{"A": "hello", "B": "world"}
	actions := AnalyzeEvalActions("{{.A}} and {{.B | upper}}", data, funcs)
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	// First action: {{.A}}
	if actions[0].ActionIndex != 0 {
		t.Errorf("action 0 index = %d, want 0", actions[0].ActionIndex)
	}
	if len(actions[0].Steps) != 1 {
		t.Errorf("action 0 steps = %d, want 1", len(actions[0].Steps))
	}
	// Second action: {{.B | upper}}
	if actions[1].ActionIndex != 1 {
		t.Errorf("action 1 index = %d, want 1", actions[1].ActionIndex)
	}
	if len(actions[1].Steps) != 2 {
		t.Errorf("action 1 steps = %d, want 2", len(actions[1].Steps))
	}
}

func TestAnalyzeEvalActionsMultiLine(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"trim": strings.TrimSpace,
	}
	data := map[string]any{"A": "hello", "B": "  world  "}
	input := "line1 {{.A}}\nline2 {{.B | trim}}"
	actions := AnalyzeEvalActions(input, data, funcs)
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	if actions[0].SourceLine != 1 {
		t.Errorf("action 0 source line = %d, want 1", actions[0].SourceLine)
	}
	if actions[1].SourceLine != 2 {
		t.Errorf("action 1 source line = %d, want 2", actions[1].SourceLine)
	}
	if actions[0].Source != "line1 {{.A}}" {
		t.Errorf("action 0 source = %q, want %q", actions[0].Source, "line1 {{.A}}")
	}
	if actions[1].Source != "line2 {{.B | trim}}" {
		t.Errorf("action 1 source = %q, want %q", actions[1].Source, "line2 {{.B | trim}}")
	}
}

func TestAnalyzeEvalActionsSubPipeline(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"trim":  strings.TrimSpace,
		"upper": strings.ToUpper,
		"printf": func(format string, args ...any) string {
			return "formatted"
		},
	}
	data := map[string]any{"NAME": "  hello  "}
	actions := AnalyzeEvalActions(`{{printf "%s" (.NAME | trim)}}`, data, funcs)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	ea := actions[0]
	// Steps should be depth-first: .NAME → trim → printf
	if len(ea.Steps) < 3 {
		t.Fatalf("expected at least 3 steps, got %d", len(ea.Steps))
	}
	// Verify depth-first order: variable → trim → printf
	if ea.Steps[0].Target != ".NAME" {
		t.Errorf("step 0 target = %q, want .NAME", ea.Steps[0].Target)
	}
	if ea.Steps[1].Target != "trim" {
		t.Errorf("step 1 target = %q, want trim", ea.Steps[1].Target)
	}
	if ea.Steps[2].Target != "printf" {
		t.Errorf("step 2 target = %q, want printf", ea.Steps[2].Target)
	}
}

func TestAnalyzeEvalActionsNoActions(t *testing.T) {
	t.Parallel()
	actions := AnalyzeEvalActions("plain text no templates", nil, nil)
	if len(actions) != 0 {
		t.Errorf("expected 0 actions for plain text, got %d", len(actions))
	}
}

func TestAnalyzeEvalActionsInvalidTemplate(t *testing.T) {
	t.Parallel()
	actions := AnalyzeEvalActions("{{invalid...syntax", nil, nil)
	if actions != nil {
		t.Errorf("expected nil for invalid template, got %v", actions)
	}
}

func TestAnalyzeEvalActionsGlobalStepNumbers(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"upper": strings.ToUpper,
	}
	data := map[string]any{"A": "hello", "B": "world"}
	actions := AnalyzeEvalActions("{{.A | upper}} {{.B | upper}}", data, funcs)
	if len(actions) != 2 {
		t.Fatalf("expected 2 actions, got %d", len(actions))
	}
	// Step numbers should be global (1,2) for action 0, (3,4) for action 1
	if actions[0].Steps[0].StepNum != 1 {
		t.Errorf("action 0, step 0 num = %d, want 1", actions[0].Steps[0].StepNum)
	}
	if actions[0].Steps[1].StepNum != 2 {
		t.Errorf("action 0, step 1 num = %d, want 2", actions[0].Steps[1].StepNum)
	}
	if actions[1].Steps[0].StepNum != 3 {
		t.Errorf("action 1, step 0 num = %d, want 3", actions[1].Steps[0].StepNum)
	}
	if actions[1].Steps[1].StepNum != 4 {
		t.Errorf("action 1, step 1 num = %d, want 4", actions[1].Steps[1].StepNum)
	}
}

func TestAnalyzeEvalActionsPipedInputShown(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"trim":  strings.TrimSpace,
		"upper": strings.ToUpper,
	}
	data := map[string]any{"NAME": "  hello  "}
	actions := AnalyzeEvalActions("{{.NAME | trim | upper}}", data, funcs)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	ea := actions[0]
	if len(ea.Steps) != 3 {
		t.Fatalf("expected 3 steps, got %d", len(ea.Steps))
	}
	// trim's input should show the piped value
	if !strings.Contains(ea.Steps[1].Input, `"  hello  "`) {
		t.Errorf("trim input should contain piped value, got: %q", ea.Steps[1].Input)
	}
	// upper's input should show the piped value
	if !strings.Contains(ea.Steps[2].Input, `"hello"`) {
		t.Errorf("upper input should contain piped value, got: %q", ea.Steps[2].Input)
	}
}

func TestAnalyzeEvalActionsResultLine(t *testing.T) {
	t.Parallel()
	funcs := template.FuncMap{
		"upper": strings.ToUpper,
	}
	data := map[string]any{"NAME": "world"}
	actions := AnalyzeEvalActions("echo '{{.NAME | upper}}'", data, funcs)
	if len(actions) != 1 {
		t.Fatalf("expected 1 action, got %d", len(actions))
	}
	if actions[0].Result != "echo 'WORLD'" {
		t.Errorf("Result = %q, want %q", actions[0].Result, "echo 'WORLD'")
	}
}

// ── CollectDiagnostics Tests ──

func TestCollectDiagnosticsOutputAnomaly(t *testing.T) {
	t.Parallel()
	evalActions := []EvalAction{
		{
			ActionIndex: 0,
			Steps: []TemplateStep{
				{StepNum: 1, Operation: "Resolve a Variable", Target: ".NAME", Input: "hello"},
				{StepNum: 2, Operation: "Apply a Function", Target: "trim", Input: `trim "hello"`, Output: "hello"},
				{StepNum: 3, Operation: "Apply a Function", Target: "printf", Input: `printf "%s %s" hello`, Output: `hello %!s(MISSING)`},
			},
		},
	}
	pipeSteps := []PipeStep{
		{FuncName: "printf", Args: []string{`"%s %s"`, ".NAME"}, ArgsValues: []string{`"%s %s"`, `"hello"`}},
	}
	diags := CollectDiagnostics(evalActions, pipeSteps)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	d := diags[0]
	if d.DiagType != "output_anomaly" {
		t.Errorf("DiagType = %q, want output_anomaly", d.DiagType)
	}
	if d.FuncName != "printf" {
		t.Errorf("FuncName = %q, want printf", d.FuncName)
	}
	if d.StepNum != 3 {
		t.Errorf("StepNum = %d, want 3", d.StepNum)
	}
	if d.Signature == "" {
		t.Error("expected non-empty Signature")
	}
	if d.Example == "" {
		t.Error("expected non-empty Example")
	}
}

func TestCollectDiagnosticsExecError(t *testing.T) {
	t.Parallel()
	evalActions := []EvalAction{
		{
			ActionIndex: 0,
			Steps: []TemplateStep{
				{
					StepNum: 1, Operation: "Apply a Function", Target: "trim",
					Input:  "trim",
					Output: `<exec error: template: :1:2: executing "" at <trim>: wrong number of args for trim: want 1 got 0>`,
				},
			},
		},
	}
	diags := CollectDiagnostics(evalActions, nil)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	d := diags[0]
	if d.DiagType != "exec_error" {
		t.Errorf("DiagType = %q, want exec_error", d.DiagType)
	}
	if d.FuncName != "trim" {
		t.Errorf("FuncName = %q, want trim", d.FuncName)
	}
	if !strings.Contains(d.ErrorMsg, "wrong number of args") {
		t.Errorf("ErrorMsg should contain 'wrong number of args', got: %q", d.ErrorMsg)
	}
}

func TestCollectDiagnosticsCorrectFunctionBlamed(t *testing.T) {
	t.Parallel()
	// This is the key regression test: trim should NOT be blamed for printf's error.
	// trim executes successfully (Output: "node"), but printf produces %!s(MISSING).
	evalActions := []EvalAction{
		{
			ActionIndex: 0,
			Result:      `ENGINE:                 node %!s(MISSING)`,
			Steps: []TemplateStep{
				{StepNum: 1, Operation: "Resolve a Variable", Target: ".ENGINE", Input: " node "},
				{StepNum: 2, Operation: "Apply a Function", Target: "trim", Input: `trim " node "`, Output: "node"},
				{StepNum: 3, Operation: "Resolve a Variable", Target: ".SPACE", Input: "20"},
				{
					StepNum: 4, Operation: "Apply a Function", Target: "printf",
					Input:  `printf "%s: %*s %s" "ENGINE" 20 node`,
					Output: `ENGINE:                 node %!s(MISSING)`,
				},
			},
		},
	}
	diags := CollectDiagnostics(evalActions, nil)

	// Only printf should produce a diagnostic, NOT trim
	if len(diags) != 1 {
		t.Fatalf("expected exactly 1 diagnostic, got %d", len(diags))
	}
	if diags[0].FuncName != "printf" {
		t.Errorf("blamed function = %q, want printf (trim should NOT be blamed)", diags[0].FuncName)
	}
	if diags[0].DiagType != "output_anomaly" {
		t.Errorf("DiagType = %q, want output_anomaly", diags[0].DiagType)
	}
}

func TestCollectDiagnosticsNoDiagForCleanOutput(t *testing.T) {
	t.Parallel()
	evalActions := []EvalAction{
		{
			ActionIndex: 0,
			Steps: []TemplateStep{
				{StepNum: 1, Operation: "Resolve a Variable", Target: ".NAME", Input: "hello"},
				{StepNum: 2, Operation: "Apply a Function", Target: "trim", Input: `trim "hello"`, Output: "hello"},
			},
		},
	}
	diags := CollectDiagnostics(evalActions, nil)
	if len(diags) != 0 {
		t.Errorf("expected 0 diagnostics for clean output, got %d", len(diags))
	}
}

func TestCollectDiagnosticsMultipleIssues(t *testing.T) {
	t.Parallel()
	// Two actions, both with errors
	evalActions := []EvalAction{
		{
			ActionIndex: 0,
			Steps: []TemplateStep{
				{
					StepNum: 1, Operation: "Apply a Function", Target: "trim",
					Output: `<exec error: wrong number of args>`,
				},
			},
		},
		{
			ActionIndex: 1,
			Steps: []TemplateStep{
				{
					StepNum: 2, Operation: "Apply a Function", Target: "printf",
					Output: `hello %!s(MISSING)`,
				},
			},
		},
	}
	diags := CollectDiagnostics(evalActions, nil)
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diags))
	}
	if diags[0].DiagType != "exec_error" {
		t.Errorf("diag 0 type = %q, want exec_error", diags[0].DiagType)
	}
	if diags[1].DiagType != "output_anomaly" {
		t.Errorf("diag 1 type = %q, want output_anomaly", diags[1].DiagType)
	}
}

func TestCollectDiagnosticsTrimFalsePositiveFiltered(t *testing.T) {
	t.Parallel()
	// trim receives printf's error output as input and passes it through.
	// This should NOT produce a diagnostic for trim — the error was in the input.
	evalActions := []EvalAction{
		{
			ActionIndex: 0,
			Steps: []TemplateStep{
				{StepNum: 1, Operation: "Resolve a Variable", Target: ".SPACE", Input: "20"},
				{StepNum: 2, Operation: "Resolve a Variable", Target: ".ENGINE", Input: " node "},
				{
					StepNum: 3, Operation: "Apply a Function", Target: "printf",
					Input:  `printf "%s: %*s %s" "ENGINE" 20  node `,
					Output: `ENGINE:                node  %!s(MISSING)`,
				},
				{
					StepNum: 4, Operation: "Apply a Function", Target: "trim",
					Input:  `trim "ENGINE:                node  %!s(MISSING)"`,
					Output: `ENGINE:                node  %!s(MISSING)`,
				},
			},
		},
	}
	diags := CollectDiagnostics(evalActions, nil)

	// Only printf should produce a diagnostic, NOT trim
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic (printf only), got %d", len(diags))
	}
	if diags[0].FuncName != "printf" {
		t.Errorf("blamed = %q, want printf (trim is a false positive)", diags[0].FuncName)
	}
}

func TestCollectDiagnosticsBadWidthDetected(t *testing.T) {
	t.Parallel()
	// %*s with a non-integer width produces %!(BADWIDTH)
	evalActions := []EvalAction{
		{
			ActionIndex: 0,
			Steps: []TemplateStep{
				{
					StepNum: 1, Operation: "Apply a Function", Target: "printf",
					Input:  `printf "%s: %*s" "ENGINE" ".SPACE"`,
					Output: `ENGINE: %!(BADWIDTH)%!s(MISSING)`,
				},
			},
		},
	}
	diags := CollectDiagnostics(evalActions, nil)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	if !strings.Contains(diags[0].ErrorMsg, "width specifier") || !strings.Contains(diags[0].ErrorMsg, "non-integer") {
		t.Errorf("error message should explain BADWIDTH, got: %q", diags[0].ErrorMsg)
	}
}

func TestCollectDiagnosticsSpecificMissingCount(t *testing.T) {
	t.Parallel()
	// printf "%s: %*s" expects 3 args (1 for %s, 2 for %*s) but only gets 2
	evalActions := []EvalAction{
		{
			ActionIndex: 0,
			Steps: []TemplateStep{
				{
					StepNum: 1, Operation: "Apply a Function", Target: "printf",
					Input:  `printf "%s: %*s" "ENGINE" 20`,
					Output: `ENGINE: %!s(MISSING)`,
				},
			},
		},
	}
	diags := CollectDiagnostics(evalActions, nil)
	if len(diags) != 1 {
		t.Fatalf("expected 1 diagnostic, got %d", len(diags))
	}
	// Error message should mention the specific count: expects 3, got 2, 1 missing
	if !strings.Contains(diags[0].ErrorMsg, "expects 3") {
		t.Errorf("error should say 'expects 3', got: %q", diags[0].ErrorMsg)
	}
	if !strings.Contains(diags[0].ErrorMsg, "only 2 provided") {
		t.Errorf("error should say 'only 2 provided', got: %q", diags[0].ErrorMsg)
	}
}

func TestCollectDiagnosticsPerActionCallParams(t *testing.T) {
	t.Parallel()
	// Two different printf calls with different args should produce
	// different Call/Params for each diagnostic
	evalActions := []EvalAction{
		{
			ActionIndex: 0,
			Steps: []TemplateStep{
				{
					StepNum: 1, Operation: "Apply a Function", Target: "printf",
					Input:  `printf "%s: %*s" "ENGINE" 20`,
					Output: `ENGINE: %!s(MISSING)`,
				},
			},
		},
		{
			ActionIndex: 1,
			Steps: []TemplateStep{
				{
					StepNum: 2, Operation: "Apply a Function", Target: "printf",
					Input:  `printf "%s: %*s %s" "ENGINE" 20`,
					Output: `ENGINE: %!s(MISSING) %!s(MISSING)`,
				},
			},
		},
	}
	diags := CollectDiagnostics(evalActions, nil)
	if len(diags) != 2 {
		t.Fatalf("expected 2 diagnostics, got %d", len(diags))
	}
	// First diagnostic: format + 2 variadic args = 3 params
	if !strings.Contains(diags[0].Call, `"ENGINE", 20`) {
		t.Errorf("diag 0 Call should match its specific args, got: %q", diags[0].Call)
	}
	if len(diags[0].Params) != 3 {
		t.Errorf("diag 0 should have 3 params (format + 2 args), got %d", len(diags[0].Params))
	}
	// Second diagnostic: different format string
	if !strings.Contains(diags[1].Call, `"%s: %*s %s"`) {
		t.Errorf("diag 1 Call should have its own format string, got: %q", diags[1].Call)
	}
	// Verify per-action isolation: diag 1 has format + 2 variadic args = 3
	if len(diags[1].Params) != 3 {
		t.Errorf("diag 1 should have 3 params (format + 2 args), got %d", len(diags[1].Params))
	}
	// But the format strings themselves differ
	if diags[0].Params[0].Value == diags[1].Params[0].Value {
		t.Error("the two diagnostics should have different format string values")
	}
}

// ── countFormatVerbs Tests ──

func TestCountFormatVerbsSimple(t *testing.T) {
	t.Parallel()
	verbs, slots := countFormatVerbs(`"%s: %s"`)
	if verbs != 2 {
		t.Errorf("verbs = %d, want 2", verbs)
	}
	if slots != 2 {
		t.Errorf("slots = %d, want 2", slots)
	}
}

func TestCountFormatVerbsStarWidth(t *testing.T) {
	t.Parallel()
	// %*s requires 2 args: int width + string value
	verbs, slots := countFormatVerbs(`"%s: %*s"`)
	if verbs != 2 {
		t.Errorf("verbs = %d, want 2", verbs)
	}
	if slots != 3 {
		t.Errorf("slots = %d, want 3 (one extra for *)", slots)
	}
}

func TestCountFormatVerbsMixed(t *testing.T) {
	t.Parallel()
	// %s, %*s, %s = 3 verbs, 4 slots
	verbs, slots := countFormatVerbs(`"%s: %*s %s"`)
	if verbs != 3 {
		t.Errorf("verbs = %d, want 3", verbs)
	}
	if slots != 4 {
		t.Errorf("slots = %d, want 4", slots)
	}
}

func TestCountFormatVerbsNoVerbs(t *testing.T) {
	t.Parallel()
	verbs, slots := countFormatVerbs(`"hello world"`)
	if verbs != 0 || slots != 0 {
		t.Errorf("verbs=%d, slots=%d, want 0,0", verbs, slots)
	}
}

func TestCountFormatVerbsIntVerb(t *testing.T) {
	t.Parallel()
	verbs, slots := countFormatVerbs(`"%s: %d"`)
	if verbs != 2 || slots != 2 {
		t.Errorf("verbs=%d, slots=%d, want 2,2", verbs, slots)
	}
}

// ── analyzeFormatError Tests ──

func TestAnalyzeFormatErrorMissingArgs(t *testing.T) {
	t.Parallel()
	msg := analyzeFormatError(
		`ENGINE: %!s(MISSING)`,
		`printf "%s: %*s" "ENGINE" 20`,
		"printf",
	)
	if !strings.Contains(msg, "expects 3") {
		t.Errorf("should mention 'expects 3', got: %q", msg)
	}
	if !strings.Contains(msg, "only 2 provided") {
		t.Errorf("should mention 'only 2 provided', got: %q", msg)
	}
}

func TestAnalyzeFormatErrorBadWidth(t *testing.T) {
	t.Parallel()
	msg := analyzeFormatError(
		`ENGINE: %!(BADWIDTH)%!s(MISSING)`,
		`printf "%s: %*s" "ENGINE" ".SPACE"`,
		"printf",
	)
	if !strings.Contains(msg, "width specifier") {
		t.Errorf("should mention width specifier, got: %q", msg)
	}
	if !strings.Contains(msg, "non-integer") {
		t.Errorf("should explain non-integer value, got: %q", msg)
	}
}

func TestAnalyzeFormatErrorNonPrintf(t *testing.T) {
	t.Parallel()
	msg := analyzeFormatError(
		`%!s(MISSING)`,
		`trim something`,
		"trim",
	)
	if msg != "Output contains format error pattern(s)" {
		t.Errorf("non-printf should get generic message, got: %q", msg)
	}
}

// ── buildParamMappings Tests ──

func TestBuildParamMappingsSimple(t *testing.T) {
	t.Parallel()
	// trim(s string) string — 1 param
	mappings := buildParamMappings("trim", []string{`"hello"`})
	if len(mappings) != 1 {
		t.Fatalf("expected 1 mapping, got %d", len(mappings))
	}
	if mappings[0].Name != "s" {
		t.Errorf("param name = %q, want s", mappings[0].Name)
	}
	if mappings[0].Value != `"hello"` {
		t.Errorf("param value = %q, want %q", mappings[0].Value, `"hello"`)
	}
	if mappings[0].Missing {
		t.Error("param should not be missing")
	}
}

func TestBuildParamMappingsVariadic(t *testing.T) {
	t.Parallel()
	// printf(format string, args ...any) string — variadic
	mappings := buildParamMappings("printf", []string{`"%s: %s"`, `"ENGINE"`, `"node"`})
	if len(mappings) != 3 {
		t.Fatalf("expected 3 mappings, got %d", len(mappings))
	}
	// First param: format
	if mappings[0].Name != "format" {
		t.Errorf("param 0 name = %q, want format", mappings[0].Name)
	}
	if mappings[0].Variadic {
		t.Error("param 0 should not be variadic")
	}
	// Second param: args[0]
	if mappings[1].Name != "args" {
		t.Errorf("param 1 name = %q, want args", mappings[1].Name)
	}
	if !mappings[1].Variadic {
		t.Error("param 1 should be variadic")
	}
	if mappings[1].Value != `"ENGINE"` {
		t.Errorf("param 1 value = %q, want %q", mappings[1].Value, `"ENGINE"`)
	}
	// Third param: args[1]
	if !mappings[2].Variadic {
		t.Error("param 2 should be variadic")
	}
}

func TestBuildParamMappingsMissing(t *testing.T) {
	t.Parallel()
	// printf needs format + args, but we only provide format
	mappings := buildParamMappings("printf", []string{`"%s: %s"`})
	if len(mappings) != 2 {
		t.Fatalf("expected 2 mappings (format + missing args), got %d", len(mappings))
	}
	if !mappings[1].Missing {
		t.Error("args param should be marked as missing")
	}
}

func TestBuildParamMappingsUnknownFunc(t *testing.T) {
	t.Parallel()
	mappings := buildParamMappings("unknownFunc", []string{"arg1"})
	if mappings != nil {
		t.Errorf("expected nil for unknown function, got %v", mappings)
	}
}

// ── parseSigParams Tests ──

func TestParseSigParams(t *testing.T) {
	t.Parallel()
	params := parseSigParams("printf(format string, args ...any) string")
	if len(params) != 2 {
		t.Fatalf("expected 2 params, got %d", len(params))
	}
	if params[0].Name != "format" || params[0].Type != "string" || params[0].Variadic {
		t.Errorf("param 0 = %+v, want {format string false}", params[0])
	}
	if params[1].Name != "args" || params[1].Type != "...any" || !params[1].Variadic {
		t.Errorf("param 1 = %+v, want {args ...any true}", params[1])
	}
}

func TestParseSigParamsEmpty(t *testing.T) {
	t.Parallel()
	params := parseSigParams("nosig")
	if params != nil {
		t.Errorf("expected nil for no-parens signature, got %v", params)
	}
}

// ── parseInputArgs Tests ──

func TestParseInputArgs(t *testing.T) {
	t.Parallel()
	args := parseInputArgs(`printf "%s: %s" "hello" world`, "printf")
	if len(args) != 3 {
		t.Fatalf("expected 3 args, got %d: %v", len(args), args)
	}
	if args[0] != `"%s: %s"` {
		t.Errorf("arg 0 = %q, want %q", args[0], `"%s: %s"`)
	}
	if args[1] != `"hello"` {
		t.Errorf("arg 1 = %q, want %q", args[1], `"hello"`)
	}
	if args[2] != "world" {
		t.Errorf("arg 2 = %q, want %q", args[2], "world")
	}
}

func TestParseInputArgsEmpty(t *testing.T) {
	t.Parallel()
	args := parseInputArgs("trim", "trim")
	if args != nil {
		t.Errorf("expected nil for func-name-only input, got %v", args)
	}
}

// ── containsFormatError Tests ──

func TestContainsFormatError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  bool
	}{
		{`hello %!s(MISSING) world`, true},
		{`count: %!d(MISSING)`, true},
		{`clean output`, false},
		{``, false},
		{`%!v(MISSING)`, true},
	}
	for _, tt := range tests {
		if got := containsFormatError(tt.input); got != tt.want {
			t.Errorf("containsFormatError(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// ── GenerateErrorHints Tests ──

func TestGenerateErrorHintsWithEvalActions(t *testing.T) {
	t.Parallel()
	evalActions := []EvalAction{
		{
			Steps: []TemplateStep{
				{
					StepNum: 1, Operation: "Apply a Function", Target: "printf",
					Input: `printf "%s %s" hello`, Output: `hello %!s(MISSING)`,
				},
			},
		},
	}
	hints := GenerateErrorHints(`hello %!s(MISSING)`, nil, evalActions)
	if len(hints) == 0 {
		t.Fatal("expected at least 1 hint")
	}
	if !strings.Contains(hints[0], "printf") {
		t.Errorf("hint should mention printf, got: %q", hints[0])
	}
}

func TestGenerateErrorHintsNoErrorNoHints(t *testing.T) {
	t.Parallel()
	hints := GenerateErrorHints("clean output", nil, nil)
	if len(hints) != 0 {
		t.Errorf("expected 0 hints for clean output, got %d", len(hints))
	}
}

func TestGenerateErrorHintsFallbackToPipeSteps(t *testing.T) {
	t.Parallel()
	steps := []PipeStep{
		{FuncName: "printf", Args: []string{`"%s"`}, Output: `%!s(MISSING)`},
	}
	hints := GenerateErrorHints(`%!s(MISSING)`, steps, nil)
	if len(hints) == 0 {
		t.Fatal("expected fallback hint from PipeSteps")
	}
	if !strings.Contains(hints[0], "printf") {
		t.Errorf("hint should mention printf, got: %q", hints[0])
	}
}

func TestGenerateErrorHintsGenericFallback(t *testing.T) {
	t.Parallel()
	// No steps, no eval actions — should still produce a generic printf hint
	hints := GenerateErrorHints(`%!s(MISSING)`, nil, nil)
	if len(hints) == 0 {
		t.Fatal("expected generic fallback hint")
	}
	if !strings.Contains(hints[0], "printf") {
		t.Errorf("generic hint should mention printf, got: %q", hints[0])
	}
}

// ── makeWhitespaceVisible Tests ──

func TestMakeWhitespaceVisibleSpaces(t *testing.T) {
	t.Parallel()
	got := makeWhitespaceVisible("hello world")
	if got != "hello\u00b7world" {
		t.Errorf("spaces not replaced: got %q", got)
	}
}

func TestMakeWhitespaceVisibleTabs(t *testing.T) {
	t.Parallel()
	got := makeWhitespaceVisible("hello\tworld")
	if got != "hello\u2192world" {
		t.Errorf("tabs not replaced: got %q", got)
	}
}

func TestMakeWhitespaceVisibleNewlines(t *testing.T) {
	t.Parallel()
	got := makeWhitespaceVisible("hello\nworld")
	if !strings.Contains(got, "\u21b5") {
		t.Errorf("newlines should show ↵ symbol, got %q", got)
	}
	// The actual newline should still be present (for line breaking)
	if !strings.Contains(got, "\n") {
		t.Errorf("actual newline should be preserved, got %q", got)
	}
}

func TestMakeWhitespaceVisibleCR(t *testing.T) {
	t.Parallel()
	got := makeWhitespaceVisible("hello\rworld")
	if !strings.Contains(got, "\u2190") {
		t.Errorf("carriage return should show ← symbol, got %q", got)
	}
}

func TestMakeWhitespaceVisibleCRLF(t *testing.T) {
	t.Parallel()
	got := makeWhitespaceVisible("hello\r\nworld")
	// Should show ←↵ for \r\n
	if !strings.Contains(got, "\u2190\u21b5") {
		t.Errorf("CRLF should show ←↵, got %q", got)
	}
}

func TestMakeWhitespaceVisibleANSI(t *testing.T) {
	t.Parallel()
	got := makeWhitespaceVisible("hello\033[31mred\033[0mworld")
	if !strings.Contains(got, "[ESC]") {
		t.Errorf("ANSI escape should be replaced with [ESC], got %q", got)
	}
	// Original ANSI codes should be gone
	if strings.Contains(got, "\033") {
		t.Errorf("raw ANSI codes should be stripped, got %q", got)
	}
}

func TestMakeWhitespaceVisibleTrailingNewline(t *testing.T) {
	t.Parallel()
	// This is the key case: " node\n" should show the trailing newline
	got := makeWhitespaceVisible(" node\n")
	if !strings.Contains(got, "\u21b5") {
		t.Errorf("trailing newline should show ↵ symbol, got %q", got)
	}
	if !strings.Contains(got, "\u00b7") {
		t.Errorf("leading space should show · symbol, got %q", got)
	}
}

func TestStripANSI(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input string
		want  string
	}{
		{"\033[31mred\033[0m", "[ESC]red[ESC]"},
		{"\033[38;5;87maqua\033[0m", "[ESC]aqua[ESC]"},
		{"no escape", "no escape"},
		{"\033[1m\033[2m", "[ESC][ESC]"},
	}
	for _, tt := range tests {
		if got := stripANSI(tt.input); got != tt.want {
			t.Errorf("stripANSI(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// ── Render Diagnostics (Human-Readable) Tests ──

func TestRenderTextDiagnosticsOutputAnomaly(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{
						Input:  `echo '{{printf "%s %s" .NAME}}'`,
						Output: `echo 'hello %!s(MISSING)'`,
						Diagnostics: []FuncDiagnostic{
							{
								DiagType:   "output_anomaly",
								FuncName:   "printf",
								StepNum:    2,
								Expression: `printf "%s %s" hello`,
								Signature:  `printf(format string, args ...any) string`,
								Example:    `{{printf "%s: %s" .KEY .VALUE}}`,
								Call:       `printf("%s %s", "hello")`,
								Params: []ParamMapping{
									{Name: "format", Type: "string", Value: `"%s %s"`},
									{Name: "args", Type: "...any", Value: `"hello"`, Variadic: true},
									{Name: "args", Type: "...any", Variadic: true, Missing: true},
								},
								ErrorMsg: "Output contains missing format verb argument(s)",
								Output:   `hello %!s(MISSING)`,
							},
						},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()

	checks := []string{
		"Output Anomaly",
		"printf",
		"Step 2",
		"Expression",
		"Error",
		"Output contains missing format verb",
		"Signature",
		"Example",
		"Call",
		"Params",
		"format",
		"MISSING",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("diagnostic output should contain %q, got:\n%s", check, output)
		}
	}
}

func TestRenderTextDiagnosticsExecError(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{
						Input:  `{{trim}}`,
						Output: ``,
						Diagnostics: []FuncDiagnostic{
							{
								DiagType:  "exec_error",
								FuncName:  "trim",
								StepNum:   1,
								ErrorMsg:  "wrong number of args for trim: want 1 got 0",
								Signature: "trim(s string) string",
								Example:   "{{.VAR | trim}}",
								Output:    `<exec error: wrong number of args>`,
							},
						},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()

	if !strings.Contains(output, "Execution Error") {
		t.Error("exec_error diagnostic should show 'Execution Error'")
	}
	if !strings.Contains(output, "trim") {
		t.Error("diagnostic should mention trim")
	}
	if !strings.Contains(output, "wrong number of args") {
		t.Error("diagnostic should show error message")
	}
}

func TestRenderTextDiagnosticsSuppressTips(t *testing.T) {
	t.Parallel()
	// When diagnostics are present, legacy tips should be suppressed
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{
						Input:  `{{printf "%s" .X}}`,
						Output: `%!s(MISSING)`,
						Diagnostics: []FuncDiagnostic{
							{
								DiagType: "output_anomaly", FuncName: "printf", StepNum: 1,
								ErrorMsg: "anomaly detected",
							},
						},
						Tips: []string{"Hint: printf signature: printf(format string, args ...any) string"},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()

	if strings.Contains(output, "ℹ Note:") {
		t.Error("legacy tips should be suppressed when structured diagnostics exist")
	}
}

func TestRenderTextNoDiagnosticsShowsTips(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{
						Input:  `{{printf "%s" .X}}`,
						Output: `%!s(MISSING)`,
						Tips:   []string{"Hint: printf format error"},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()

	if !strings.Contains(output, "ℹ Note:") {
		t.Error("legacy tips should be shown when no structured diagnostics exist")
	}
}

// ── JSON Renderer Diagnostics Tests ──

func TestRenderJSONDiagnosticsStructure(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{
						Input:   `echo '{{printf "%s %s" .NAME}}'`,
						Output:  `echo 'hello %!s(MISSING)'`,
						Context: "cmds[0]",
						Diagnostics: []FuncDiagnostic{
							{
								DiagType:   "output_anomaly",
								FuncName:   "printf",
								StepNum:    2,
								Expression: `printf "%s %s" hello`,
								Signature:  `printf(format string, args ...any) string`,
								Example:    `{{printf "%s: %s" .KEY .VALUE}}`,
								Call:       `printf("%s %s", "hello")`,
								Params: []ParamMapping{
									{Name: "format", Type: "string", Value: `"%s %s"`},
									{Name: "args", Type: "...any", Value: `"hello"`, Variadic: true},
								},
								ErrorMsg: "Output contains missing format verb argument(s)",
								Output:   `hello %!s(MISSING)`,
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := RenderJSON(&buf, report, nil)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	// Parse the JSON output
	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}

	// Navigate to diagnostics
	tasks, ok := result["tasks"].([]any)
	if !ok || len(tasks) == 0 {
		t.Fatal("expected tasks array in JSON")
	}
	task := tasks[0].(map[string]any)
	templates, ok := task["templates"].([]any)
	if !ok || len(templates) == 0 {
		t.Fatal("expected templates array in JSON")
	}
	tmpl := templates[0].(map[string]any)
	diagnostics, ok := tmpl["diagnostics"].([]any)
	if !ok || len(diagnostics) == 0 {
		t.Fatal("expected diagnostics array in JSON output")
	}
	diag := diagnostics[0].(map[string]any)

	// Verify all expected fields are present
	expectedFields := map[string]any{
		"diag_type": "output_anomaly",
		"func_name": "printf",
		"step_num":  float64(2), // JSON numbers are float64
		"error_msg": "Output contains missing format verb argument(s)",
	}
	for key, want := range expectedFields {
		got, exists := diag[key]
		if !exists {
			t.Errorf("JSON diagnostic missing field %q", key)
			continue
		}
		if got != want {
			t.Errorf("JSON diagnostic %q = %v, want %v", key, got, want)
		}
	}

	// Verify string fields are present and non-empty
	for _, field := range []string{"expression", "signature", "example", "call", "output"} {
		v, exists := diag[field]
		if !exists {
			t.Errorf("JSON diagnostic missing field %q", field)
			continue
		}
		if v == "" {
			t.Errorf("JSON diagnostic field %q should not be empty", field)
		}
	}

	// Verify params array
	params, ok := diag["params"].([]any)
	if !ok || len(params) == 0 {
		t.Fatal("expected params array in JSON diagnostic")
	}
	param0 := params[0].(map[string]any)
	if param0["name"] != "format" {
		t.Errorf("param 0 name = %v, want format", param0["name"])
	}
	if param0["type"] != "string" {
		t.Errorf("param 0 type = %v, want string", param0["type"])
	}
	param1 := params[1].(map[string]any)
	if param1["variadic"] != true {
		t.Errorf("param 1 variadic = %v, want true", param1["variadic"])
	}
}

func TestRenderJSONDiagnosticsExecError(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{
						Input:  "{{trim}}",
						Output: "",
						Diagnostics: []FuncDiagnostic{
							{
								DiagType: "exec_error",
								FuncName: "trim",
								StepNum:  1,
								ErrorMsg: "wrong number of args",
								Output:   "<exec error: wrong number of args>",
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := RenderJSON(&buf, report, nil)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	tasks := result["tasks"].([]any)
	task := tasks[0].(map[string]any)
	templates := task["templates"].([]any)
	tmpl := templates[0].(map[string]any)
	diagnostics := tmpl["diagnostics"].([]any)
	diag := diagnostics[0].(map[string]any)

	if diag["diag_type"] != "exec_error" {
		t.Errorf("diag_type = %v, want exec_error", diag["diag_type"])
	}
	if diag["func_name"] != "trim" {
		t.Errorf("func_name = %v, want trim", diag["func_name"])
	}
}

func TestRenderJSONNoDiagnosticsOmitted(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{Input: "{{.NAME}}", Output: "hello"},
				},
			},
		},
	}

	var buf bytes.Buffer
	err := RenderJSON(&buf, report, nil)
	if err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	tasks := result["tasks"].([]any)
	task := tasks[0].(map[string]any)
	templates := task["templates"].([]any)
	tmpl := templates[0].(map[string]any)

	if _, exists := tmpl["diagnostics"]; exists {
		t.Error("diagnostics should be omitted from JSON when empty")
	}
}

// ── JSON EvalActions Structure Tests ──

func TestRenderJSONEvalActionsStructure(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{
						Input:   "echo '{{.NAME | upper}}'",
						Output:  "echo 'WORLD'",
						Context: "cmds[0]",
						EvalActions: []EvalAction{
							{
								ActionIndex: 0,
								SourceLine:  1,
								Source:      "echo '{{.NAME | upper}}'",
								Result:      "echo 'WORLD'",
								Steps: []TemplateStep{
									{StepNum: 1, Operation: "Resolve a Variable", Target: ".NAME", Input: "world"},
									{StepNum: 2, Operation: "Apply a Function", Target: "upper", Input: `upper "world"`, Output: "WORLD"},
								},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	if err := RenderJSON(&buf, report, nil); err != nil {
		t.Fatalf("RenderJSON failed: %v", err)
	}

	var result map[string]any
	if err := json.Unmarshal(buf.Bytes(), &result); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}

	tasks := result["tasks"].([]any)
	task := tasks[0].(map[string]any)
	templates := task["templates"].([]any)
	tmpl := templates[0].(map[string]any)
	evalActions, ok := tmpl["eval_actions"].([]any)
	if !ok || len(evalActions) == 0 {
		t.Fatal("expected eval_actions array in JSON")
	}
	ea := evalActions[0].(map[string]any)

	if ea["action_index"] != float64(0) {
		t.Errorf("action_index = %v, want 0", ea["action_index"])
	}
	if ea["source_line"] != float64(1) {
		t.Errorf("source_line = %v, want 1", ea["source_line"])
	}
	if ea["source"] != "echo '{{.NAME | upper}}'" {
		t.Errorf("source = %v", ea["source"])
	}
	if ea["result"] != "echo 'WORLD'" {
		t.Errorf("result = %v", ea["result"])
	}

	steps, ok := ea["steps"].([]any)
	if !ok || len(steps) != 2 {
		t.Fatalf("expected 2 steps, got %d", len(steps))
	}
	step0 := steps[0].(map[string]any)
	if step0["step"] != float64(1) {
		t.Errorf("step 0 num = %v, want 1", step0["step"])
	}
	if step0["operation"] != "Resolve a Variable" {
		t.Errorf("step 0 operation = %v", step0["operation"])
	}
	if step0["target"] != ".NAME" {
		t.Errorf("step 0 target = %v", step0["target"])
	}

	step1 := steps[1].(map[string]any)
	if step1["step"] != float64(2) {
		t.Errorf("step 1 num = %v, want 2", step1["step"])
	}
	if step1["output"] != "WORLD" {
		t.Errorf("step 1 output = %v", step1["output"])
	}
}

// ── Whitespace Visibility Applied to Diagnostics Tests ──

func TestWhitespaceVisibilityAppliedToDiagnostics(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{
						Input:  "{{.X}}",
						Output: "val",
						Diagnostics: []FuncDiagnostic{
							{
								DiagType:   "output_anomaly",
								FuncName:   "printf",
								StepNum:    1,
								Expression: `printf "%s" "hello world"`,
								Call:       `printf("%s", "hello world")`,
								Output:     `hello world`,
								ErrorMsg:   `missing arg`,
								Params: []ParamMapping{
									{Name: "format", Type: "string", Value: `"%s %s"`},
								},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	RenderText(&buf, report, &RenderOptions{ShowWhitespaces: true})
	output := buf.String()

	// Whitespace in diagnostic fields should be visible
	if !strings.Contains(output, "\u00b7") {
		t.Error("spaces in diagnostic output should be made visible with ·")
	}
}

func TestWhitespaceVisibilityAppliedToEvalActions(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "test",
				Templates: []TemplateTrace{
					{
						Input:  "{{.X}}",
						Output: "hello world",
						EvalActions: []EvalAction{
							{
								ActionIndex: 0,
								SourceLine:  1,
								Source:      "echo {{.X}}",
								Result:      "echo hello world",
								Steps: []TemplateStep{
									{StepNum: 1, Operation: "Resolve a Variable", Target: ".X", Input: "hello world"},
								},
							},
						},
					},
				},
			},
		},
	}

	var buf bytes.Buffer
	RenderText(&buf, report, &RenderOptions{ShowWhitespaces: true})
	output := buf.String()

	// All spaces in Source, Result, and step Input should be visible
	if !strings.Contains(output, "hello\u00b7world") {
		t.Error("spaces in EvalAction fields should be made visible with ·")
	}
}

// ── Render EvalActions Human-Readable Tests ──

func TestRenderTextEvalActionsShown(t *testing.T) {
	t.Parallel()
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Templates: []TemplateTrace{
					{
						Input:  "{{.NAME | upper}}",
						Output: "WORLD",
						EvalActions: []EvalAction{
							{
								ActionIndex: 0,
								SourceLine:  1,
								Source:      "{{.NAME | upper}}",
								Result:      "WORLD",
								Steps: []TemplateStep{
									{StepNum: 1, Operation: "Resolve a Variable", Target: ".NAME", Input: "world"},
									{StepNum: 2, Operation: "Apply a Function", Target: "upper", Input: `upper "world"`, Output: "WORLD"},
								},
							},
						},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()

	checks := []string{
		"Evaluation Steps",
		"Action 1 of 1",
		"line 1",
		"Step 1:",
		"Resolve a Variable",
		".NAME",
		"Step 2:",
		"Apply a Function",
		"upper",
		"WORLD",
	}
	for _, check := range checks {
		if !strings.Contains(output, check) {
			t.Errorf("EvalActions output should contain %q", check)
		}
	}
	// S and R labels
	if !strings.Contains(output, "S ") {
		t.Error("should contain Source label 'S'")
	}
	if !strings.Contains(output, "R ") {
		t.Error("should contain Result label 'R'")
	}
}

func TestRenderTextEvalActionsFallbackToPipeSteps(t *testing.T) {
	t.Parallel()
	// When EvalActions is empty, should fall back to PipeSteps
	report := &TraceReport{
		Tasks: []*TaskTrace{
			{
				TaskName: "t",
				Templates: []TemplateTrace{
					{
						Input:  "{{.NAME | trim}}",
						Output: "world",
						Steps: []PipeStep{
							{FuncName: ".NAME", Output: "  world  "},
							{FuncName: "trim", Output: "world"},
						},
					},
				},
			},
		},
	}
	var buf bytes.Buffer
	RenderText(&buf, report, nil)
	output := buf.String()

	if !strings.Contains(output, "Pipe Steps") {
		t.Error("should fall back to Pipe Steps when EvalActions is empty")
	}
}
