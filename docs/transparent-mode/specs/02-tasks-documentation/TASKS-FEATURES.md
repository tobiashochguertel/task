# Transparent Mode — Features

New functionality required by the specification and Improvements.1.md that is not yet implemented.

---

### T001 - Implement Step-by-Step Template Evaluation Tracing

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
The specification (05-OUTPUT-FORMAT.md) and Improvements.1.md require a detailed step-by-step trace of template evaluation showing how each variable is resolved and each function is applied within a template expression. The current implementation only shows basic pipe steps (function name + args + output) but does NOT show the incremental resolution process.

The user wants to see exactly how a template like:
```
echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'
```
is resolved step by step, showing:
- Each variable resolution (`.ENGINE` → `"node\n"`)
- Each function application (`trim` → `"node"`)
- The intermediate state of the full expression after each step
- Using `I` (input), `O` (output), `F` (final command) labels

**Current Behavior**:
```
  Template Evaluations:
  [1] (cmds[0]) Input:  echo "{{spew (.ENGINE | trim)}}"
       Output: echo "(string) (len=4) "node""
       Vars used: ENGINE, SPACE
```

**Expected Behavior** (from Improvements.1.md):
```
  Template Evaluation — cmds[0]:
  ┌─ Steps:
  │ Step 1: Resolve a Variable (`.ENGINE`)
  |   I     "node\n"
  |   F     echo "{{spew ("node\n" | trim)}}"
  │ Step 2: Apply a Function (`trim`)
  |   I     trim "node\n"
  |   O     "node"
  |   F     echo "{{spew "node"}}"
  | Step 3: Apply a Function (`spew`)
  |   I     spew "node"
  |   O     "(string) (len=4) \"node\""
  |   F     echo "(string) (len=4) \"node\""
  └─
```

**Implementation**:

1. **Extend `pipe_analyzer.go`** — Add a new function `AnalyzeSteps()` that performs fine-grained step-by-step analysis:
   - Walk the template AST
   - For each `FieldNode` (`.VARNAME`): record "Resolve a Variable" step with I (resolved value) and F (expression with variable substituted)
   - For each function call in a pipe: record "Apply a Function" step with I (input args), O (output), F (expression after function applied)
   - Track the evolving expression string after each substitution

2. **Extend `model.go`** — Add a `TemplateStep` struct:
   ```go
   type TemplateStep struct {
       StepNum     int      // Sequential step number
       Operation   string   // "Resolve a Variable" or "Apply a Function"
       Target      string   // Variable name or function name
       Input       string   // Input value(s)
       Output      string   // Output value (empty for variable resolution)
       Expression  string   // Full expression state after this step
   }
   ```
   Add `DetailedSteps []TemplateStep` to `TemplateTrace`.

3. **Update `renderer.go`** — Render the steps in the box-drawing format with I/O/F labels.

4. **Update `renderer_json.go`** — Include detailed steps in JSON output.

5. **Update `templater.go`** — Pass additional context to enable step-by-step tracing.

**Acceptance Criteria**:
- [ ] Each template expression shows numbered steps
- [ ] Variable resolutions show step type "Resolve a Variable (`.NAME`)"
- [ ] Function applications show step type "Apply a Function (`funcname`)"
- [ ] Each step shows I (input), O (output where applicable), F (final expression state)
- [ ] Steps are rendered inside `┌─ Steps:` / `└─` box
- [ ] Multi-line values are properly indented with `|` prefix
- [ ] JSON output includes the same step data
- [ ] Works with nested pipes like `(.ENGINE | trim)`
- [ ] Works with multi-argument functions like `printf`

**Testing**:
- [ ] Unit tests for `AnalyzeSteps()` with simple variable resolution
- [ ] Unit tests with pipe chains
- [ ] Unit tests with nested parenthesized expressions
- [ ] Integration test with real Taskfile (example 03-template-pipes)
- [ ] Golden test updates for text and JSON

**Dependencies**: I001 (box-drawing format)

**Files to modify**:
- `internal/transparent/pipe_analyzer.go` — new `AnalyzeSteps()` function
- `internal/transparent/model.go` — new `TemplateStep` struct
- `internal/transparent/renderer.go` — render steps
- `internal/transparent/renderer_json.go` — JSON steps
- `internal/templater/templater.go` — wire step analysis

---

### T002 - Implement `--show-whitespaces` CLI Option

**Status**: 🔲 TODO
**Priority**: 🟡 Medium

**Description**:
Improvements.1.md requests a CLI option to make whitespace visible in the output. This is critical for debugging template expressions where leading/trailing spaces or tabs cause unexpected behavior. The option should replace spaces with `·` and tabs with `→` in variable values and template output.

**Example with `--show-whitespaces`**:
```
ENGINE  taskfile-vars  string  ·  ·node
CLI     taskfile-vars  string  ·  ·node···--experimental-strip-types··src/cli.ts
```

Without it (current default):
```
ENGINE  taskfile-vars  string  ·   node
CLI     taskfile-vars  string  ·   node   --experimental-strip-types  src/cli.ts
```

**Implementation**:

1. **Add CLI flag** in `internal/flags/flags.go`:
   - Add `ShowWhitespaces bool` field
   - Add pflag: `pflag.BoolVar(&ShowWhitespaces, "show-whitespaces", false, "Make whitespace visible (· for spaces, → for tabs)")`

2. **Add to Executor** in `executor.go`:
   - Add `ShowWhitespaces bool` field
   - Add `WithShowWhitespaces()` option

3. **Add to RenderOptions** in `renderer.go`:
   - Add `ShowWhitespaces bool` field

4. **Create whitespace replacement function**:
   ```go
   func makeWhitespaceVisible(s string) string {
       s = strings.ReplaceAll(s, " ", "·")
       s = strings.ReplaceAll(s, "\t", "→")
       return s
   }
   ```

5. **Apply in renderers** — When `ShowWhitespaces` is true, apply the replacement to:
   - Variable values
   - Template input/output
   - Command raw/resolved strings

6. **Add legend** — When whitespace mode is active, display a legend at the top:
   ```
   Legend: · = space, → = tab
   ```

**Acceptance Criteria**:
- [ ] `--show-whitespaces` flag recognized by CLI
- [ ] Spaces replaced with `·` in variable values when flag is active
- [ ] Tabs replaced with `→` when flag is active
- [ ] Legend displayed at top of report when flag is active
- [ ] Default behavior unchanged (no whitespace visualization without flag)
- [ ] Works in both text and JSON output modes
- [ ] JSON output includes a `whitespace_visible: true` field when active

**Testing**:
- [ ] Unit test for `makeWhitespaceVisible()` function
- [ ] Integration test with `--show-whitespaces` flag
- [ ] Verify legend is displayed
- [ ] Verify default behavior unchanged

**Dependencies**: None

**Files to modify**:
- `internal/flags/flags.go` — new flag
- `executor.go` — new field + option
- `internal/transparent/renderer.go` — whitespace replacement + legend
- `internal/transparent/renderer_json.go` — whitespace in JSON
- `transparent.go` — pass option to renderer
