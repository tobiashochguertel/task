# silent

Suppress command output.

## Type

`boolean`


## Description

When true, commands won't print to stdout/stderr unless they fail.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level

- Command level








## Examples


### Silent task

Suppress all output for task

```yaml
tasks:
  cleanup:
    silent: true
    cmds:
      - rm -rf tmp/
      - echo "Cleaned up"

```



### Silent command

Suppress output for specific command

```yaml
tasks:
  build:
    cmds:
      - cmd: echo "Building..."
        silent: true
      - go build

```



### Global silent

Make all tasks silent by default

```yaml
version: '3'

silent: true

tasks:
  build:
    cmds:
      - go build

```






## Related

- [output](./output.md)




