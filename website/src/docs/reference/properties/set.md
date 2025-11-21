# set

Enable POSIX shell options.

## Type

`array`


## Description

Configure shell behavior by enabling specific POSIX options.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level






## Options

| Option | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `errexit` | `string` | Exit immediately if a command exits with a non-zero status. | `set: [errexit]` |
| `nounset` | `string` | Treat unset variables as an error. | `set: [nounset]` |
| `pipefail` | `string` | Return the exit status of the last command in a pipe that failed. | `set: [pipefail]` |




## Examples


### Common usage

Enable multiple shell options

```yaml
set: [errexit, nounset, pipefail]

```



### Task-level override

Set options for specific task

```yaml
tasks:
  strict-task:
    set: [errexit, nounset]
    cmds:
      - echo "Running with strict mode"

```








