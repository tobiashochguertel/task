# cmds

Commands to execute for this task.

## Type

`cmds`


## Description

A list of shell commands or Task calls to execute sequentially. Unlike deps which run in parallel, cmds run one after another in order. Supports regular commands, task calls, and deferred cleanup commands.

Commands can be strings for simple shell commands, or objects for advanced features like passing variables, controlling silent mode, or using defer for cleanup.




## Contexts

This property can be used in:


- Task level








## Examples


### Simple shell commands

Run commands sequentially

```yaml
tasks:
  build:
    cmds:
      - go build -o app
      - chmod +x app
      - ./app --version

```



### Calling other tasks

Execute tasks serially

```yaml
version: '3'

tasks:
  main-task:
    cmds:
      - task: task-to-be-called
      - task: another-task
      - echo "Both done"
  
  task-to-be-called:
    cmds:
      - echo "Task to be called"
  
  another-task:
    cmds:
      - echo "Another task"

```



### Task calls with variables

Pass variables to called tasks

```yaml
version: '3'

tasks:
  greet:
    vars:
      RECIPIENT: '{{default "World" .RECIPIENT}}'
    cmds:
      - echo "Hello, {{.RECIPIENT}}!"
  
  greet-pessimistically:
    cmds:
      - task: greet
        vars: { RECIPIENT: 'Cruel World' }
        silent: true

```



### Deferred cleanup

Cleanup that runs after task completes

```yaml
version: '3'

tasks:
  default:
    cmds:
      - mkdir -p tmpdir/
      - defer: rm -rf tmpdir/
      - echo 'Do work on tmpdir/'

```



### Deferred task call

Cleanup via another task

```yaml
version: '3'

tasks:
  default:
    cmds:
      - mkdir -p tmpdir/
      - defer: { task: cleanup }
      - echo 'Do work on tmpdir/'
  
  cleanup:
    cmds:
      - rm -rf tmpdir/

```



### Using EXIT_CODE in defer

Check task success in cleanup

```yaml
version: '3'

tasks:
  build:
    cmds:
      - go build
      - defer: |
          if [ "$EXIT_CODE" != "0" ]; then
            echo "Build failed, cleaning up..."
            rm -rf dist/
          fi

```






## Related

- [deps](./deps.md)




