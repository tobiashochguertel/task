# Transparent Mode — Tasks Overview

**Branch**: `feature/transparent-mode`
**Specification**: `docs/transparent-mode/specs/01-10`
**Improvements**: `docs/transparent-mode/issues/Improvements.1.md`

**See detailed task descriptions in:**
- [Fixes](./TASKS-FIXES.md) - Bug fixes for incorrect/broken behavior
- [Features](./TASKS-FEATURES.md) - New functionality required by specs
- [Improvements](./TASKS-IMPROVEMENTS.md) - Enhancements to existing output formatting

---

## Quick Status Overview

| Category     | Total | Done | In Progress | TODO |
|--------------|-------|------|-------------|------|
| Fixes        | 4     | 0    | 0           | 4    |
| Features     | 2     | 0    | 0           | 2    |
| Improvements | 7     | 0    | 0           | 7    |
| **Total**    | **13**| **0**| **0**       |**13**|

---

## Task Sets

### Set 1: Output Formatting Fixes

**Priority**: High
**Description**: Fix broken or incorrect output behavior that contradicts the specification and user expectations.

| Order | Task ID | Title                                      | Status  |
|-------|---------|--------------------------------------------|---------|
| 1     | F001    | Remove Value Column Truncation             | 🔲 TODO |
| 2     | F002    | Fix Output Header to Match Spec            | 🔲 TODO |
| 3     | F003    | Fix Commands Section Box-Drawing Format    | 🔲 TODO |
| 4     | F004    | Fix JSON Output Parity with Text Output    | 🔲 TODO |

### Set 2: New Features (Spec-Required)

**Priority**: High
**Description**: Features explicitly required by the specification and Improvements.1.md that are not yet implemented.

| Order | Task ID | Title                                              | Status  |
|-------|---------|----------------------------------------------------|---------|
| 1     | T001    | Implement Step-by-Step Template Evaluation Tracing  | 🔲 TODO |
| 2     | T002    | Implement `--show-whitespaces` CLI Option           | 🔲 TODO |

### Set 3: Output Quality Improvements

**Priority**: Medium
**Description**: Improvements to the human-readable and JSON output quality, formatting, and user experience.

| Order | Task ID | Title                                              | Status  |
|-------|---------|----------------------------------------------------|---------|
| 1     | I001    | Improve Template Evaluation Box-Drawing Format     | 🔲 TODO |
| 2     | I002    | Add Note/Hint Section Display                      | 🔲 TODO |
| 3     | I003    | Add Error Highlighting for Template Errors         | 🔲 TODO |
| 4     | I004    | Add Function Signature Hints on Errors             | 🔲 TODO |
| 5     | I005    | Improve Variables Table with Proper Borders         | 🔲 TODO |
| 6     | I006    | Improve Shadow Warning Display Format              | 🔲 TODO |
| 7     | I007    | Ensure JSON Output Contains All Information        | 🔲 TODO |

---

## Task Summary

| ID   | Category    | Title                                             | Priority  | Status  | Dependencies |
|------|-------------|---------------------------------------------------|-----------|---------|--------------|
| F001 | Fix         | Remove Value Column Truncation                    | 🟢 High   | 🔲 TODO | -            |
| F002 | Fix         | Fix Output Header to Match Spec                   | 🟢 High   | 🔲 TODO | -            |
| F003 | Fix         | Fix Commands Section Box-Drawing Format           | 🟢 High   | 🔲 TODO | I001         |
| F004 | Fix         | Fix JSON Output Parity with Text Output           | 🟢 High   | 🔲 TODO | I007         |
| T001 | Feature     | Implement Step-by-Step Template Evaluation Tracing| 🟢 High   | 🔲 TODO | I001         |
| T002 | Feature     | Implement `--show-whitespaces` CLI Option         | 🟡 Medium | 🔲 TODO | -            |
| I001 | Improvement | Improve Template Evaluation Box-Drawing Format    | 🟢 High   | 🔲 TODO | -            |
| I002 | Improvement | Add Note/Hint Section Display                     | 🟡 Medium | 🔲 TODO | I001         |
| I003 | Improvement | Add Error Highlighting for Template Errors        | 🟡 Medium | 🔲 TODO | -            |
| I004 | Improvement | Add Function Signature Hints on Errors            | 🟡 Medium | 🔲 TODO | I003         |
| I005 | Improvement | Improve Variables Table with Proper Borders        | 🟡 Medium | 🔲 TODO | -            |
| I006 | Improvement | Improve Shadow Warning Display Format             | 🟡 Medium | 🔲 TODO | I005         |
| I007 | Improvement | Ensure JSON Output Contains All Information       | 🟡 Medium | 🔲 TODO | -            |

---

## Testing Notes

- All changes must pass existing tests (`go test ./...`)
- Golden file tests in `internal/transparent/testdata/golden/` must be updated when output format changes
- Both text and JSON renderers must be tested for each change
- Verify with real Taskfiles in `docs/transparent-mode/examples/`
- Build and install `task-dev` to verify manually: `go build -o ~/go/bin/task-dev ./cmd/task`

## Implementation Notes

- **Architecture**: All rendering changes go in `internal/transparent/renderer.go` (text) and `renderer_json.go` (JSON)
- **Model changes**: Data model in `internal/transparent/model.go` — add fields as needed
- **CLI flags**: New flags in `internal/flags/flags.go`
- **Tracer changes**: `internal/transparent/tracer.go` — for new data collection
- **Pipe analysis**: `internal/transparent/pipe_analyzer.go` — for template introspection
- **SOLID principle**: Keep Tracer (collect) and Renderer (format) separate
- **Both output modes**: Every feature must work in both `--transparent` (text) and `--transparent --json` (JSON)
