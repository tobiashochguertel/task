# Transparent Mode — Fixes

Bug fixes for incorrect or broken behavior that contradicts the specification.

---

### F001 - Remove Value Column Truncation

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
The current renderer truncates variable values to 60 characters via `truncate()` in `internal/transparent/renderer.go:162,166,178`. The user explicitly states in Improvements.1.md: "No column of the table should be truncated, otherwise we can't understand the output / report." Long paths like `/Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder` are cut off with `...`, making the report useless for debugging.

**Current Behavior**:
```
ROOT_DIR  special  string  ·  /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
```

**Expected Behavior**:
```
ROOT_DIR  special  string  ·  /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder
```

**Implementation**:
- Remove the `truncate()` call on variable values in `renderVars()` at `renderer.go:162`
- Remove the `truncate()` call on shadow values at `renderer.go:178`
- Keep the `truncate()` function available for optional use (e.g., `--show-whitespaces` mode may need it)
- Values should always be displayed in full — the terminal handles line wrapping

**Acceptance Criteria**:
- [ ] Variable values are never truncated in text output
- [ ] Shadow warning values are never truncated
- [ ] Long paths are fully visible
- [ ] Dynamic var `sh:` commands are not truncated
- [ ] JSON output already shows full values (verify no truncation there)
- [ ] Golden tests updated to reflect full values

**Testing**:
- [ ] Verify with example Taskfile that has long path values
- [ ] Run golden tests and update snapshots
- [ ] Verify JSON output is unaffected

**Dependencies**: None

**Files to modify**:
- `internal/transparent/renderer.go` — remove `truncate()` calls on values

---

### F002 - Fix Output Header to Match Spec

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
The spec (`05-OUTPUT-FORMAT.md`) defines the output header as a double-box format, but the current implementation uses a simpler single-line format.

**Current Behavior**:
```
══════ Transparent Mode Report ══════
```

**Expected Behavior** (from spec):
```
╔══════════════════════════════════════════════════════╗
║  TRANSPARENT MODE — Variable & Template Diagnostics  ║
╚══════════════════════════════════════════════════════╝
```

And the footer:
```
══════ End Report ══════
```
Should be similarly styled or removed (the spec doesn't show a footer).

**Implementation**:
- Update `RenderText()` in `renderer.go:65` to use the double-box header format
- Update or remove the footer at `renderer.go:83`
- Ensure colors (cyan, bold) are applied to the box characters

**Acceptance Criteria**:
- [ ] Header matches spec format with ╔═╗ / ║ ║ / ╚═╝ box
- [ ] Header text reads "TRANSPARENT MODE — Variable & Template Diagnostics"
- [ ] Footer is consistent with header style
- [ ] Colors are respected (cyan + bold)
- [ ] NO_COLOR environment variable still works
- [ ] Golden tests updated

**Testing**:
- [ ] Visual verification with terminal
- [ ] Golden test updates
- [ ] NO_COLOR test

**Dependencies**: None

**Files to modify**:
- `internal/transparent/renderer.go` — `RenderText()` header/footer

---

### F003 - Fix Commands Section Box-Drawing Format

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
The current Commands section uses a flat format (`[0] raw: / resolved:`) but both the spec and Improvements.1.md require box-drawing format with headers per command.

**Current Behavior**:
```
  Commands:
  [0] raw:      echo "{{spew (.ENGINE | trim)}}"
       resolved: echo "(string) (len=4) "node""
```

**Expected Behavior** (from Improvements.1.md):
```
  Commands — cmds[0]:
  ┌─ Raw:
  │ echo "{{spew (.ENGINE | trim)}}"
  └─
  ┌─ Resolved:
  │ echo "(string) (len=4) "node""
  └─
```

**Implementation**:
- Rewrite `renderCmds()` in `renderer.go:227-243` to use box-drawing format
- Each command gets its own header: `Commands — cmds[N]:`
- Raw content in `┌─ Raw:` / `│` / `└─` block
- Resolved content in `┌─ Resolved:` / `│` / `└─` block
- Multi-line content: each line prefixed with `│ `
- FOR-loop iteration label shown in header

**Acceptance Criteria**:
- [ ] Commands use box-drawing format (┌─, │, └─)
- [ ] Each command has its own `Commands — cmds[N]:` header
- [ ] Multi-line commands properly indented with `│ ` prefix
- [ ] FOR-loop iteration labels displayed in header
- [ ] Raw and resolved shown only when they differ (single block when same)
- [ ] Golden tests updated

**Testing**:
- [ ] Test with single-line and multi-line commands
- [ ] Test with FOR-loop expanded commands
- [ ] Test with commands that have no template substitution

**Dependencies**: I001 (shares box-drawing helper functions)

**Files to modify**:
- `internal/transparent/renderer.go` — `renderCmds()`

---

### F004 - Fix JSON Output Parity with Text Output

**Status**: 🔲 TODO
**Priority**: 🟢 High

**Description**:
The JSON output is missing information that the text output provides. The spec (05-OUTPUT-FORMAT.md) shows that JSON commands should include `pipe_steps`, but the current `jsonCmdTrace` struct only has `index`, `raw`, `resolved`, and `iteration`. Template evaluations and pipe steps at the command level are absent from JSON.

**Current JSON command structure**:
```json
{
  "index": 0,
  "raw": "echo ...",
  "resolved": "echo ..."
}
```

**Expected JSON command structure** (from spec):
```json
{
  "index": 0,
  "raw": "echo {{printf \"%s : %s\" .GREETING .NAME | trim}}",
  "resolved": "echo Hello : Task",
  "pipe_steps": [
    {"func": "printf", "args": [...], "output": "Hello : Task"},
    {"func": "trim", "input": "Hello : Task", "output": "Hello : Task"}
  ]
}
```

**Implementation**:
- Add `Templates []jsonTemplateTrace` and/or `PipeSteps []jsonPipeStep` to `jsonCmdTrace`
- Link template traces to their corresponding commands in the JSON renderer
- Ensure all information visible in text output is also available in JSON

**Acceptance Criteria**:
- [ ] JSON commands include pipe_steps when template substitution occurs
- [ ] JSON commands include template evaluation details
- [ ] JSON output contains all info that text output contains
- [ ] JSON golden tests updated

**Testing**:
- [ ] Compare JSON and text output for same Taskfile to verify parity
- [ ] Test with pipe-heavy templates

**Dependencies**: I007

**Files to modify**:
- `internal/transparent/renderer_json.go` — `jsonCmdTrace`, `RenderJSON()`
