# dir

`dir` sets the working directory for the task or included Taskfile.

## Type

`string`

## Usage

### Task Level

Sets the working directory for the commands in the task.

```yaml
tasks:
  build:
    dir: ./src
    cmds:
      - go build
```

### Include Level

Sets the working directory for all tasks in the included Taskfile.

```yaml
includes:
  backend:
    taskfile: ./backend
    dir: ./backend
```
