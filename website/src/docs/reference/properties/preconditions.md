# preconditions

`preconditions` defines a list of checks that must pass before the task can run.

## Type

`[]Precondition`

## Properties

| Property | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `sh` | `string` | The command to execute. | `sh: test -f file` |
| `msg` | `string` | Message to display if check fails. | `msg: "File missing"` |

## Usage

If any precondition fails (exits with non-zero code), the task will fail immediately.

### Simple Command

```yaml
tasks:
  deploy:
    preconditions:
      - test -f ./app
    cmds: ...
```

### With Message

```yaml
tasks:
  deploy:
    preconditions:
      - sh: test -n "$API_KEY"
        msg: "API_KEY is missing"
    cmds: ...
```
