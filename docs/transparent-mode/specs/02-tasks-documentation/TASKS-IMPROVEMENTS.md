# Transparent Mode — Improvements

Enhancements to existing output formatting and user experience.

---

### I001 - Improve Template Evaluation Box-Drawing Format

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
The current template evaluation section uses a flat format that is hard to read. The spec (05-OUTPUT-FORMAT.md) and Improvements.1.md require box-drawing characters to structure the output into clear Input, Output, and Vars used sections.

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
  ┌─ Input:
  │ echo "{{spew (.ENGINE | trim)}}"
  └─
  ┌─ Output:
  │ echo "(string) (len=4) "node""
  └─
  ┌─ Vars used:
  │ ENGINE, SPACE
  └─
```

**Implementation**:
- Create shared box-drawing helper functions in `renderer.go`:
  ```go
  func renderBoxStart(w io.Writer, label string)  // prints "  ┌─ Label:"
  func renderBoxLine(w io.Writer, line string)     // prints "  │ line"
  func renderBoxEnd(w io.Writer)                   // prints "  └─"
  func renderBoxContent(w io.Writer, label string, content string)  // full box with multi-line support
  ```
- Rewrite `renderTemplates()` to use box-drawing format
- Each template evaluation gets a header: `Template Evaluation — context:`
- Multi-line content: split on `\n` and prefix each line with `│ `

**Acceptance Criteria**:
- [ ] Template evaluations use `┌─`, `│`, `└─` box-drawing characters
- [ ] Each section (Input, Output, Vars used) is in its own box
- [ ] Multi-line content properly indented with `│ ` prefix
- [ ] Header shows context (e.g., `cmds[0]`)
- [ ] Box-drawing helper functions are reusable for Commands section (F003)
- [ ] Golden tests updated

**Testing**:
- [ ] Unit tests for box-drawing helper functions
- [ ] Test with single-line and multi-line template content
- [ ] Golden test updates

**Dependencies**: None

**Files to modify**:
- `internal/transparent/renderer.go` — new helper functions + rewrite `renderTemplates()`

---

### I002 - Add Note/Hint Section Display

**Status**: 🔲 TODO
**Priority**: 🟡 Medium

**Description**:
The spec (05-OUTPUT-FORMAT.md) shows an `ℹ Note:` section after template evaluations that provides helpful suggestions to users. Currently, tips are shown with `💡` inline but there is no structured Note section as shown in the spec.

**Expected Behavior** (from spec):
```
  ℹ Note: If you intended to trim .NAME before printf, use:
    {{printf "%s : %s" .GREETING (.NAME | trim)}}
```

**Implementation**:
- After rendering template evaluation boxes, render tips/notes as a distinct section
- Use `ℹ` icon with cyan coloring for the "Note:" label
- Indent the suggestion code by 4 spaces under the note
- Generate more contextual notes based on detected patterns (e.g., multi-arg functions piped)

**Acceptance Criteria**:
- [ ] Notes displayed after template evaluation with `ℹ Note:` prefix
- [ ] Suggestion code indented under the note
- [ ] Cyan coloring for the note icon and label
- [ ] Notes only shown when there is something useful to say
- [ ] JSON output includes notes in a `notes` array field

**Testing**:
- [ ] Test with templates that trigger pipe tips
- [ ] Golden test updates

**Dependencies**: I001 (box-drawing format should be in place first)

**Files to modify**:
- `internal/transparent/renderer.go` — note rendering
- `internal/transparent/renderer_json.go` — notes in JSON

---

### I003 - Add Error Highlighting for Template Errors

**Status**: 🔲 TODO
**Priority**: 🟡 Medium

**Description**:
When template evaluation produces errors like `%!s(MISSING)`, these should be visually highlighted with red color/style to make them immediately visible. Currently, errors appear as plain text in the output with no special formatting.

**Example from Improvements.1.md**:
The output `ENGINE:                 node %!s(MISSING)` contains `%!s(MISSING)` which indicates a printf formatting error. This should be highlighted in red.

**Implementation**:
- In the resolved output rendering, scan for common Go template error patterns:
  - `%!s(MISSING)`, `%!d(MISSING)`, `%!v(MISSING)`, etc.
  - `<no value>` (before replacement)
  - `<exec error: ...>`
  - `<parse error: ...>`
- Wrap matched patterns with red ANSI color codes
- Create a function `highlightErrors(s string) string` that replaces error patterns with colored versions

**Acceptance Criteria**:
- [ ] `%!s(MISSING)` and similar patterns highlighted in red in text output
- [ ] `<no value>` patterns highlighted before replacement
- [ ] Error patterns in resolved commands are highlighted
- [ ] JSON output includes an `errors` array field identifying error positions
- [ ] No false positives (only highlight known error patterns)
- [ ] NO_COLOR respected

**Testing**:
- [ ] Unit test for `highlightErrors()` with various error patterns
- [ ] Test with Taskfile that produces `%!s(MISSING)` error
- [ ] Golden test updates

**Dependencies**: None

**Files to modify**:
- `internal/transparent/renderer.go` — `highlightErrors()` function + apply in rendering

---

### I004 - Add Function Signature Hints on Errors

**Status**: 🔲 TODO
**Priority**: 🟡 Medium

**Description**:
When a template function produces an error (e.g., `printf` with wrong number of arguments producing `%!s(MISSING)`), the report should show the function signature as a hint so the user knows what arguments the function expects.

**Expected Behavior**:
```
  ┌─ Resolved:
  │ echo 'ENGINE:                 node %!s(MISSING)'
  └─
  ⚠ Error: printf format has more verbs than arguments
  ℹ Hint: printf signature: printf(format string, args ...any) string
    Example: {{printf "%s: %s" .KEY .VALUE}}
```

**Implementation**:
- Create a `funcSignatures` map in `pipe_analyzer.go` with known function signatures:
  ```go
  var funcSignatures = map[string]string{
      "printf":  "printf(format string, args ...any) string",
      "trim":    "trim(s string) string",
      "upper":   "upper(s string) string",
      ...
  }
  ```
- When an error is detected in template output, look up the function and show its signature
- Render as a hint section below the error

**Acceptance Criteria**:
- [ ] Function signature shown when template error is detected
- [ ] Covers common functions: printf, trim, upper, lower, replace, etc.
- [ ] Example usage shown below signature
- [ ] JSON output includes function signature hints
- [ ] Only shown when relevant (not for every template)

**Testing**:
- [ ] Test with printf that has wrong number of args
- [ ] Test with missing function arguments
- [ ] Golden test updates

**Dependencies**: I003 (error highlighting should be in place)

**Files to modify**:
- `internal/transparent/pipe_analyzer.go` — `funcSignatures` map
- `internal/transparent/renderer.go` — hint rendering
- `internal/transparent/renderer_json.go` — hints in JSON

---

### I005 - Improve Variables Table with Proper Borders

**Status**: 🔲 TODO
**Priority**: 🟡 Medium

**Description**:
The spec (05-OUTPUT-FORMAT.md) shows variables in a proper bordered table format with `┌─────┬────┐`, `├─────┼────┤`, `└─────┴────┘` borders. The current implementation uses plain text columns with minimal separators.

**Current Behavior**:
```
  Variables:
  Name              Origin          Type      Ref?    Value
  ──────────────────────────────────────────────────────
  ALIAS             special         string      ·     default
```

**Expected Behavior** (from spec):
```
  Variables in scope:
  ┌─────────────┬────────────┬───────────────┬───────────────┐
  │ Name        │ Value      │ Origin        │ Shadows?      │
  ├─────────────┼────────────┼───────────────┼───────────────┤
  │ GREETING    │ "Hello"    │ taskfile:vars │               │
  │ NAME        │ "Task"     │ task:vars     │ ⚠ SHADOWS     │
  └─────────────┴────────────┴───────────────┴───────────────┘
```

**Implementation**:
- Rewrite `renderVars()` to use bordered table format
- Compute column widths dynamically based on content
- Use Unicode box-drawing characters for borders
- Column order: Name, Value, Origin, Type, Ref?, Shadows?
- Values should be quoted for strings (e.g., `"Hello"`)

**Acceptance Criteria**:
- [ ] Variables table uses full box-drawing borders
- [ ] Column widths adapt to content
- [ ] Header row with separator
- [ ] Values properly quoted for strings
- [ ] Ref and Shadow info in dedicated columns
- [ ] Golden tests updated

**Testing**:
- [ ] Test with various variable types (string, []string, bool, int)
- [ ] Test with long variable names and values
- [ ] Test with ref variables
- [ ] Golden test updates

**Dependencies**: None

**Files to modify**:
- `internal/transparent/renderer.go` — rewrite `renderVars()`

---

### I006 - Improve Shadow Warning Display Format

**Status**: 🔲 TODO
**Priority**: 🟡 Medium

**Description**:
The spec shows shadow warnings as a dedicated column in the variables table with multi-line content showing the original value and origin. The current implementation appends shadow info after the value on the same line.

**Current Behavior**:
```
NAME  task-vars  string  ·  Task ⚠ SHADOWS NAME="  World  " [taskfile:vars]
```

**Expected Behavior** (from spec):
```
│ NAME   │ "Task"  │ task:vars  │ ⚠ SHADOWS          │
│        │         │            │ global NAME         │
│        │         │            │ ="  World  "        │
```

**Implementation**:
- With I005's bordered table, the Shadows column shows multi-line content
- First line: `⚠ SHADOWS`
- Second line: `global NAME` (or appropriate scope label)
- Third line: `="original value"`
- Use yellow coloring for the warning

**Acceptance Criteria**:
- [ ] Shadow warnings in dedicated Shadows? column
- [ ] Multi-line shadow info showing scope and original value
- [ ] Yellow coloring for warnings
- [ ] JSON output includes structured shadow information (already partially done)

**Testing**:
- [ ] Test with variable shadowing (example 02-variable-shadowing)
- [ ] Golden test updates

**Dependencies**: I005 (bordered table format)

**Files to modify**:
- `internal/transparent/renderer.go` — shadow display in table

---

### I007 - Ensure JSON Output Contains All Information

**Status**: 🔲 TODO
**Priority**: 🟡 Medium

**Description**:
The JSON output must contain all the same information as the human-readable text output. Currently, several pieces of information are missing from JSON:
- Commands don't include template evaluation details or pipe steps
- No notes/hints attached to commands
- Template evaluation steps (T001) not in JSON
- No error highlighting information

**Implementation**:
- Add `Templates []jsonTemplateTrace` to `jsonCmdTrace`
- Add `Notes []string` to `jsonCmdTrace` or `jsonTemplateTrace`
- Add `DetailedSteps []jsonTemplateStep` to `jsonTemplateTrace` (after T001)
- Add `Errors []string` to `jsonTemplateTrace` for detected error patterns
- Ensure `ShowWhitespaces` mode works in JSON (after T002)

**Acceptance Criteria**:
- [ ] JSON commands include template traces with pipe steps
- [ ] JSON includes notes/hints
- [ ] JSON includes error patterns detected
- [ ] JSON includes all step-by-step evaluation data (after T001)
- [ ] JSON schema is documented
- [ ] JSON version bumped if structure changes

**Testing**:
- [ ] Compare JSON and text output for completeness
- [ ] JSON golden test updates
- [ ] Validate JSON schema

**Dependencies**: None (can be done incrementally as other tasks complete)

**Files to modify**:
- `internal/transparent/renderer_json.go` — extend JSON structures
