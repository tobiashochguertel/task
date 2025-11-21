# cmds

`cmds` defines the list of commands to be executed by the task.

## Type

`[]Command`

## Syntax

`cmds` is a list where each item can be a string or a command object.

### Simple String

A simple shell command.

```yaml
tasks:
  build:
    cmds:
      - go build ./...
      - echo "Build complete"
```

### Command Object

Allows more control over execution.

```yaml
tasks:
  example:
    cmds:
      - cmd: echo "Hello World"
        silent: true
        ignore_error: false
        platforms: [linux, darwin]
        set: [errexit]
        shopt: [globstar]
```

#### Properties

| Property | Type | Description |
| :--- | :--- | :--- |
| `cmd` | `string` | The command to execute. |
| `silent` | `bool` | If `true`, suppresses the command output. |
| `ignore_error` | `bool` | If `true`, continues execution even if the command fails. |
| `platforms` | `[]string` | List of platforms where this command should run (e.g., `linux`, `windows`). |
| `set` | `[]string` | POSIX shell options (e.g., `errexit`, `xtrace`). |
| `shopt` | `[]string` | Bash shell options (e.g., `globstar`). |

### Task Reference

Execute another task as a command.

```yaml
tasks:
  deploy:
    cmds:
      - task: build
        vars: { BUILD_TYPE: release }
```

### Deferred Commands

Schedule a command to run when the task finishes (successfully or not).

```yaml
tasks:
  cleanup:
    cmds:
      - mkdir tmp
      - defer: rm -rf tmp
      - echo "Working..."
```

### Loops

Run a command multiple times.

```yaml
tasks:
  greet:
    cmds:
      - for: [alice, bob]
        cmd: echo "Hello {{.ITEM}}"
```

See [Looping](../templating.md#loops) for more details.
