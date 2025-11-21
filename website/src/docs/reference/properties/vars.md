# Variables

Variables (`vars`) are a core feature of Task, allowing you to define reusable values and dynamic content.

## Definition Levels

Variables can be defined at different levels, with a specific precedence order.

### 1. Taskfile Level (Global)

Variables defined at the root of the `Taskfile.yml` are available to all tasks.

```yaml
version: '3'

vars:
  GREETING: Hello

tasks:
  greet:
    cmds:
      - echo "{{.GREETING}}"
```

### 2. Task Level

Variables defined within a task are available only to that task. They override global variables.

```yaml
version: '3'

vars:
  GREETING: Hello

tasks:
  greet:
    vars:
      GREETING: Hi
    cmds:
      - echo "{{.GREETING}}" # Output: Hi
```

### 3. Call Level (CLI or Dependency)

Variables can be passed when calling a task, either from the CLI or as a dependency.

**CLI:**
```bash
task greet GREETING=Hey
```

**Dependency:**
```yaml
tasks:
  greet-dependency:
    deps:
      - task: greet
        vars: { GREETING: "Hey from dep" }
```

## Dynamic Variables

Variables can be dynamic, meaning their value is the result of a shell command.

```yaml
vars:
  GIT_COMMIT:
    sh: git rev-parse --short HEAD
```

The command is executed once, lazily (when the variable is first used).

## Special Variables

Task provides some special variables:

*   `{{\.TASK}}`: The name of the current task.
*   `{{\.ROOT_DIR}}`: The absolute path to the root Taskfile directory.
*   `{{\.TASKFILE_DIR}}`: The absolute path to the directory of the included Taskfile (if applicable).
*   `{{\.CLI_ARGS}}`: Arguments passed to the task from the CLI.

## Templating

Task uses the [Go template engine](https://pkg.go.dev/text/template). You can use standard Go template functions and logic.

```yaml
vars:
  NOW: '{{now | date "2006-01-02"}}'
```
