# deps

Task dependencies that run before this task.

## Type

`deps`


## Description

Define tasks that must run before the current task. Dependencies run in parallel by default.




## Contexts

This property can be used in:


- Task level








## Examples


### Simple dependencies

Run tasks before current task

```yaml
tasks:
  build:
    deps: [clean, install]
    cmds:
      - go build

```



### Dependencies with variables

Pass variables to dependencies

```yaml
tasks:
  deploy:
    deps:
      - task: build
        vars:
          ENV: production
    cmds:
      - ./deploy.sh

```






## Related

- [cmds](./cmds.md)




