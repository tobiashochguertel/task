# silent

Suppress command echo before execution.

## Type

`boolean`


## Description

Disable echoing of commands before Task runs them. When enabled, only the command output is shown, not the command itself. There are four ways to enable silent mode: at command level, task level, globally at Taskfile level, or with the --silent/-s flag.

Silent mode only suppresses command echoing. To suppress STDOUT, redirect output to /dev/null in the command itself.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level

- Command level








## Examples


### Command-level silent

Suppress echo for specific command

```yaml
version: '3'

tasks:
  echo:
    cmds:
      - cmd: echo "Print something"
        silent: true

```



### Task-level silent

Suppress echo for all commands in task

```yaml
version: '3'

tasks:
  echo:
    cmds:
      - echo "Print something"
    silent: true

```



### Global silent mode

Apply to entire Taskfile

```yaml
version: '3'

silent: true

tasks:
  echo:
    cmds:
      - echo "Print something"

```



### Suppress STDOUT

Hide both command and output

```yaml
version: '3'

tasks:
  echo:
    cmds:
      - echo "This will print nothing" > /dev/null

```



### Mixed silent and verbose

Some commands silent, others not

```yaml
tasks:
  build:
    cmds:
      - cmd: echo "Step 1 (hidden)"
        silent: true
      - echo "Step 2 (visible)"
      - cmd: echo "Step 3 (hidden)"
        silent: true

```






## Related

- [output](./output.md)




