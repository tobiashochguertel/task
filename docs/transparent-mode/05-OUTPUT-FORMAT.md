# 05 — Output Format Specification

## Design Goal

<!-- ✅ CLOSED — Output renders as an X-ray overlay with vars, origins, shadow warnings, and template eval steps. -->

The output should feel like an **X-ray overlay** on the Taskfile — showing what the template engine sees, not what the user wrote.

## Output Modes

<!-- ✅ CLOSED — Both text (renderer.go) and JSON (renderer_json.go) output modes implemented. -->

| Flag                   | Format                      | Use Case                         |
| ---------------------- | --------------------------- | -------------------------------- |
| `--transparent`        | Human-readable colored text | Terminal debugging               |
| `--transparent --json` | JSON                        | Tooling integration, IDE plugins |

---

## Human-Readable Output Example

<!-- ✅ CLOSED — Text renderer outputs global/task vars table, template eval steps, shadow warnings, pipe tips. -->

Given this Taskfile:

```yaml
version: '3'
vars:
  NAME: '  World  '
  GREETING: 'Hello'

tasks:
  greet:
    vars:
      NAME: 'Task'
    cmds:
      - 'echo {{printf "%s : %s" .GREETING .NAME | trim}}'
```

Running `task greet --transparent` outputs:

```log
╔══════════════════════════════════════════════════════╗
║  TRANSPARENT MODE — Variable & Template Diagnostics  ║
╚══════════════════════════════════════════════════════╝

── Global Variables ───────────────────────────────────

  NAME        = "  World  "          [taskfile:vars]  type:string
  GREETING    = "Hello"              [taskfile:vars]  type:string
  ROOT_DIR    = "/path/to/project"   [special]        type:string
  TASK_VERSION= "3.40.0"             [special]        type:string
  ... (5 more special vars)

── Task: greet ────────────────────────────────────────

  Variables in scope:
  ┌─────────────┬────────────┬───────────────┬───────────────┐
  │ Name        │ Value      │ Origin        │ Shadows?      │
  ├─────────────┼────────────┼───────────────┼───────────────┤
  │ GREETING    │ "Hello"    │ taskfile:vars │               │
  │ NAME        │ "Task"     │ task:vars     │ ⚠ SHADOWS     │
  │             │            │               │ global NAME   │
  │             │            │               │ ="  World  "  │
  └─────────────┴────────────┴───────────────┴───────────────┘

  Template Evaluation — cmds[0]:
  ┌──────────────────────────────────────────────────────────┐
  │ Input:  echo {{printf "%s : %s" .GREETING .NAME | trim}} │
  │                                                          │
  │ Step 1: Resolve .GREETING → "Hello"                      │
  │ Step 2: Resolve .NAME    → "Task"    (from task:vars)    │
  │ Step 3: printf "%s : %s" "Hello" "Task"                  │
  │         → "Hello : Task"                                 │
  │ Step 4: trim "Hello : Task"                              │
  │         → "Hello : Task"  (no change — no whitespace)    │
  │                                                          │
  │ Output: echo Hello : Task                                │
  └──────────────────────────────────────────────────────────┘

  ℹ Note: If you intended to trim .NAME before printf, use:
    {{printf "%s : %s" .GREETING (.NAME | trim)}}
```

## JSON Output Example

<!-- ✅ CLOSED — JSON output includes version:"1.0", global_vars, tasks with vars/cmds/templates/shadows/tips. -->

```json
{
  "version": "1.0",
  "global_vars": [
    {
      "name": "NAME",
      "value": "  World  ",
      "origin": "taskfile:vars",
      "type": "string",
      "dynamic": false
    }
  ],
  "tasks": [
    {
      "name": "greet",
      "vars": [
        {
          "name": "NAME",
          "value": "Task",
          "origin": "task:vars",
          "type": "string",
          "shadows": {
            "name": "NAME",
            "value": "  World  ",
            "origin": "taskfile:vars"
          }
        }
      ],
      "cmds": [
        {
          "index": 0,
          "raw": "echo {{printf \"%s : %s\" .GREETING .NAME | trim}}",
          "resolved": "echo Hello : Task",
          "pipe_steps": [
            {
              "func": "printf",
              "args": ["\"%s : %s\"", ".GREETING→Hello", ".NAME→Task"],
              "output": "Hello : Task"
            },
            {
              "func": "trim",
              "input": "Hello : Task",
              "output": "Hello : Task"
            }
          ]
        }
      ]
    }
  ]
}
```

## Color Scheme

<!-- ✅ CLOSED — fatih/color used: Cyan headers, Green names, Magenta origins, Yellow warnings. NO_COLOR respected. -->

| Element         | Color   | Logger Constant  |
| --------------- | ------- | ---------------- |
| Section headers | Cyan    | `logger.Cyan`    |
| Variable names  | Green   | `logger.Green`   |
| Variable values | Default | `logger.Default` |
| Origin tags     | Magenta | `logger.Magenta` |
| Shadow warnings | Yellow  | `logger.Yellow`  |
| Template steps  | Default | `logger.Default` |
| Errors          | Red     | `logger.Red`     |

## Verbosity Levels

<!-- ✅ CLOSED — -v verbose mode shows all CLI_* and env vars; default hides them with count. -->

| Flags                      | What's shown                                                     |
| -------------------------- | ---------------------------------------------------------------- |
| `--transparent`            | Task-specific vars, shadowing, template eval for requested tasks |
| `--transparent -v`         | All of the above + all global/special vars, all scopes           |
| `--transparent --list-all` | All tasks' variable scopes (no template eval)                    |
