# deps

`deps` defines a list of tasks that must be executed before the current task.

## Type

`[]Dependency`

## Execution

Dependencies are executed in parallel by default. To run them sequentially, you can use the `task` command in `cmds` instead.

## Syntax

`deps` is a list where each item can be a string or a dependency object.

### Simple String

Reference a task by name.

```yaml
tasks:
  deploy:
    deps: [build, test]
```

### Dependency Object

Pass variables or configure execution.

```yaml
tasks:
  deploy:
    deps:
      - task: build
        vars: { ENV: production }
      - task: test
        silent: true
```

#### Properties

| Property | Type | Description |
| :--- | :--- | :--- |
| `task` | `string` | The name of the task to run. |
| `vars` | `map[string]Variable` | Variables to pass to the dependency. |
| `silent` | `bool` | If `true`, suppresses the dependency output. |

### Loops

Run a dependency multiple times.

```yaml
tasks:
  test-all:
    deps:
      - for: [unit, integration]
        task: test
        vars: { TYPE: "{{.ITEM}}" }
```
