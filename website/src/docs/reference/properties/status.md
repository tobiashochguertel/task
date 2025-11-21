# status

`status` defines a list of commands to check if a task is up-to-date.

## Type

`[]string`

## Usage

If all commands in the `status` list exit with code `0`, the task is considered up-to-date and will be skipped.

```yaml
tasks:
  generate-file:
    status:
      - test -f generated.txt
    cmds:
      - echo "content" > generated.txt
```
