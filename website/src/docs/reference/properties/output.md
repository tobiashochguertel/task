# output

`output` controls how the output of tasks is displayed in the console.

## Type

`string` or `object`

## Options

| Option | Description | Example |
| :--- | :--- | :--- |
| `interleaved` | Output from all tasks is printed as it happens. (Default) | `output: interleaved` |
| `group` | Output is buffered and printed only when the task finishes. | `output: group` |
| `prefixed` | Adds a prefix with the task name to each line of output. | `output: prefixed` |

## Syntax

### Simple String

```yaml
output: group
```

### Object

Allows customizing the group output.

```yaml
output:
  group:
    begin: "::group::{{\.TASK}}"
    end: "::endgroup::"
    error_only: false
```
