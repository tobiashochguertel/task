# run

`run` defines the default execution behavior for tasks.

## Type

`string`

## Options

| Option | Description | Example |
| :--- | :--- | :--- |
| `always` | The task will run every time it is called. (Default) | `run: always` |
| `once` | The task will run only once per execution of the `task` command. | `run: once` |
| `when_changed` | The task will run only if its sources have changed. | `run: when_changed` |

## Usage

```yaml
run: once
```
