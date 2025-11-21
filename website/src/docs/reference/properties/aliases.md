# aliases

`aliases` defines alternative names for a task or namespace.

## Type

`[]string`

## Usage

Aliases can be used to run the task from the CLI just like the main task name.

```yaml
tasks:
  build:
    aliases: [b, compile]
    cmds: ...
```
