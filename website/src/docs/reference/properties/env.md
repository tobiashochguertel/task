# env

Environment variables for task execution.

## Type

`env`


## Description

Define environment variables that will be available during task execution.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level




## Precedence

Task level → Taskfile level → System environment





## Examples


### Task-level environment

Set environment for specific task

```yaml
tasks:
  build:
    env:
      CGO_ENABLED: 0
      GOOS: linux
    cmds:
      - go build

```



### Global environment

Set environment for all tasks

```yaml
version: '3'

env:
  NODE_ENV: production

tasks:
  build:
    cmds:
      - npm run build

```






## Related

- [vars](./vars.md)
- [dotenv](./dotenv.md)




