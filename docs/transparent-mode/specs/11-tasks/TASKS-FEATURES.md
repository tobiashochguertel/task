# Evaluation Steps Redesign — Feature Tasks

**Specification**: `docs/transparent-mode/specs/11-EVALUATION-STEPS-REDESIGN.md`

---

### T001 - Add EvalAction Data Model

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
Add the `EvalAction` struct to `internal/transparent/model.go` and update `TemplateTrace` to use `EvalActions []EvalAction` instead of `DetailedSteps []TemplateStep`. Remove the `Expression` field from `TemplateStep` since action-level `Source`/`Result` replaces it.

**Implementation**:

1. Add `EvalAction` struct to `model.go`:
   ```go
   type EvalAction struct {
       ActionIndex int            `json:"action_index"`
       SourceLine  int            `json:"source_line"`
       Source      string         `json:"source"`
       Result      string         `json:"result"`
       Steps       []TemplateStep `json:"steps"`
   }
   ```

2. Remove `Expression` field from `TemplateStep`:
   ```go
   type TemplateStep struct {
       StepNum   int    `json:"step"`
       Operation string `json:"operation"`
       Target    string `json:"target"`
       Input     string `json:"input"`
       Output    string `json:"output"`
   }
   ```

3. Replace `DetailedSteps []TemplateStep` with `EvalActions []EvalAction` in `TemplateTrace`:
   ```go
   type TemplateTrace struct {
       // ...existing fields...
       EvalActions []EvalAction `json:"eval_actions,omitempty"` // replaces DetailedSteps
       // ...existing fields...
   }
   ```

**Acceptance Criteria**:
- [ ] `EvalAction` struct exists in `model.go`
- [ ] `TemplateStep` no longer has `Expression` field
- [ ] `TemplateTrace` uses `EvalActions` instead of `DetailedSteps`
- [ ] Code compiles (downstream references will be temporarily broken — fixed in T002-T005)

**Dependencies**: None

---

### T002 - Implement AnalyzeEvalActions in pipe_analyzer

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
Replace `AnalyzeDetailedSteps` with `AnalyzeEvalActions` in `internal/transparent/pipe_analyzer.go`. The new function groups steps by ActionNode, tracks source line numbers, walks pipe commands recursively (depth-first for sub-pipelines), and computes the result line for each action.

**Implementation**:

1. Add helper `lineNumber(input string, pos parse.Pos) int` — counts newlines before `pos` to get 1-based line number.

2. Add helper `sourceLine(input string, lineNum int) string` — returns the text of a specific line.

3. Add `walkPipeCommands(cmds []*parse.CommandNode, data, funcs, stepCounter) []TemplateStep` — recursively walks commands:
   - For each command's arguments (skipping first if it's a function identifier):
     - `*parse.FieldNode` → "Resolve a Variable" step
     - `*parse.PipeNode` → recursively call `walkPipeCommands` on its `.Cmds`
     - Other nodes (string, number) → skip (literals don't generate steps)
   - If first arg is `*parse.IdentifierNode` (function call):
     - Build input string: `funcName arg1 arg2 ...` (with resolved values)
     - Evaluate the partial pipe to get output
     - Record "Apply a Function" step
   - If first arg is `*parse.FieldNode` (variable at start of pipe):
     - Record "Resolve a Variable" step

4. Implement `AnalyzeEvalActions(input string, data map[string]any, funcs template.FuncMap) []EvalAction`:
   - Parse template to get AST
   - Iterate `root.Nodes` for `*parse.ActionNode`
   - For each ActionNode:
     - Compute line number from `action.Position()`
     - Get source line text
     - Walk the pipe to generate steps
     - Evaluate the action to compute the result value
     - Build the result line (replace `{{...}}` in source line with result)
     - Append `EvalAction`

5. Remove `AnalyzeDetailedSteps` function.

**Acceptance Criteria**:
- [ ] `AnalyzeEvalActions` returns correct `[]EvalAction` for single-line templates
- [ ] Correct action grouping for multi-line templates (each `{{...}}` = one action)
- [ ] Correct source line numbers
- [ ] Correct recursive evaluation order for nested sub-pipelines like `{{func (expr | pipe)}}`
- [ ] Result line correctly shows the resolved value replacing the `{{...}}`
- [ ] Step numbers are sequential and global across all actions
- [ ] `AnalyzeDetailedSteps` is removed

**Testing**:
- [ ] Test single-action template: `{{.NAME | trim}}`
- [ ] Test multi-action template: `{{.A}} and {{.B}}`
- [ ] Test nested sub-pipeline: `{{spew (.ENGINE | trim)}}`
- [ ] Test multi-line template with mixed text and actions
- [ ] Test template with no actions (plain text)

**Dependencies**: T001

---

### T003 - Update Human-Readable Renderer for EvalActions

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
Update the `renderTemplates` function in `internal/transparent/renderer.go` to render the new `EvalActions` structure instead of `DetailedSteps`. The output uses action headers with source/result lines and groups steps per action.

**Implementation**:

1. Replace the `if len(t.DetailedSteps) > 0` block with `if len(t.EvalActions) > 0` block.

2. For each `EvalAction`:
   - Render action header: `── Action N of M — line L` (dim color)
   - Render source line: `S     <source>` (bold/cyan)
   - Render each step using existing step rendering (Step N: Operation — Target + I/O fields)
   - Render result line: `R     <result>` (green)
   - Add blank line between actions

3. Use `renderStepField` for S/R lines (supports multiline alignment).

4. Keep the fallback to `t.Steps` (PipeStep) for backward compatibility.

**Output format per action**:
```
  │
  │ ── Action 1 of 2 — line 3
  │ S     echo "{{spew (.ENGINE | trim)}}"
  │
  │ Step 1: Resolve a Variable — .ENGINE
  │   I     node
  │ Step 2: Apply a Function — trim
  │   I     trim node
  │   O     node
  │ Step 3: Apply a Function — spew
  │   I     spew node
  │   O     (string) (len=4) "node"
  │
  │ R     echo "(string) (len=4) "node""
  │
```

**Acceptance Criteria**:
- [ ] Action headers display with correct N/M/line format
- [ ] S and R lines display with correct labels and alignment
- [ ] Steps within each action are properly indented
- [ ] Multiline content in I/O/S/R aligns correctly
- [ ] No trailing blank lines before `└─`
- [ ] Fallback to PipeSteps still works when EvalActions is empty

**Dependencies**: T001

---

### T004 - Update JSON Renderer for EvalActions

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
Update `internal/transparent/renderer_json.go` to emit the new `eval_actions` JSON structure instead of `detailed_steps`. The JSON tags on the struct handle most of this automatically, but verify the output format matches the golden reference.

**Implementation**:

1. Check if `renderer_json.go` directly references `DetailedSteps` — if so, update to `EvalActions`.
2. Verify JSON output includes `eval_actions` array with nested `steps` arrays.
3. Ensure `action_index`, `source_line`, `source`, `result` fields appear in output.

**Acceptance Criteria**:
- [ ] JSON output contains `eval_actions` instead of `detailed_steps`
- [ ] Each eval_action has `action_index`, `source_line`, `source`, `result`, `steps`
- [ ] Steps within eval_actions have `step`, `operation`, `target`, `input`, `output`
- [ ] JSON golden files pass after regeneration

**Dependencies**: T001

---

### T005 - Update Whitespace Visibility for EvalActions

**Status**: 🔲 TODO
**Priority**: 🟡 Medium

**Description**:
Update `applyWSToTemplates` in `renderer.go` to apply `makeWhitespaceVisible` to the new `EvalAction` fields (`Source`, `Result`, and each step's `Input`/`Output`).

**Implementation**:

Replace the `DetailedSteps` whitespace application with:
```go
if len(tc.EvalActions) > 0 {
    actions := make([]EvalAction, len(tc.EvalActions))
    for j, ea := range tc.EvalActions {
        ac := ea
        ac.Source = makeWhitespaceVisible(ea.Source)
        ac.Result = makeWhitespaceVisible(ea.Result)
        steps := make([]TemplateStep, len(ea.Steps))
        for k, ds := range ea.Steps {
            sc := ds
            sc.Input = makeWhitespaceVisible(ds.Input)
            sc.Output = makeWhitespaceVisible(ds.Output)
            steps[k] = sc
        }
        ac.Steps = steps
        actions[j] = ac
    }
    tc.EvalActions = actions
}
```

**Acceptance Criteria**:
- [ ] With `--show-whitespaces`, Source/Result/Input/Output show `·` for spaces and `→` for tabs
- [ ] Without `--show-whitespaces`, no transformation applied

**Dependencies**: T001

---

### T006 - Wire Up, Update Tests, Regenerate Golden Files

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
Wire `AnalyzeEvalActions` into `templater.go` (replacing `AnalyzeDetailedSteps`), update all tests that reference the old model, and regenerate golden files.

**Implementation**:

1. In `internal/templater/templater.go`, replace:
   ```go
   detailedSteps := transparent.AnalyzeDetailedSteps(v, data, template.FuncMap(templateFuncs))
   ```
   with:
   ```go
   evalActions := transparent.AnalyzeEvalActions(v, data, template.FuncMap(templateFuncs))
   ```
   and update `TemplateTrace` construction to use `EvalActions: evalActions`.

2. Update `internal/transparent/transparent_test.go`:
   - Update `TestRenderText` to check for `Action 1 of` instead of old format
   - Update any tests that reference `DetailedSteps` or `Expression`

3. Build binary and regenerate golden files:
   ```bash
   go build -o ./bin/task ./cmd/task/
   UPDATE_GOLDEN=1 go test ./internal/transparent/ -run TestGoldenText -count=1
   UPDATE_GOLDEN=1 go test ./internal/transparent/ -run TestGoldenJSON -count=1
   ```

4. Run full test suite to verify.

**Acceptance Criteria**:
- [ ] `go build` succeeds
- [ ] `go test ./internal/transparent/` passes
- [ ] Golden files regenerated with new format
- [ ] No references to `DetailedSteps` or `AnalyzeDetailedSteps` remain in codebase

**Dependencies**: T002, T003, T004, T005
