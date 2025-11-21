# silent

`silent` suppresses the output of the command being executed.

## Type

`bool`

## Default

`false`

## Usage

When set to `true`, Task will not print the command itself before executing it. The output of the command (stdout/stderr) is still printed unless the command itself suppresses it or `output` configuration handles it.

```yaml
tasks:
  echo:
    cmd: echo "Hello"
    silent: true
```
