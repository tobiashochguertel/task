# output

Control how task output is displayed.

## Type

`string`


## Description

Configure output handling for task execution. Options: interleaved, group, prefixed.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level






## Options

| Option | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `interleaved` | `string` | Output appears immediately as it happens (default). | `output: interleaved` |
| `group` | `string` | Output is grouped by task. | `output: group` |
| `prefixed` | `string` | Each line is prefixed with task name. | `output: prefixed` |




## Examples


### Grouped output

Show output per task

```yaml
version: '3'

output: group

tasks:
  build:
    cmds:
      - go build
      - echo "Build complete"

```



### Prefixed output

Prefix lines with task name

```yaml
tasks:
  dev:
    output: prefixed
    deps:
      - task: watch-backend
      - task: watch-frontend

```






## Related

- [silent](./silent.md)




