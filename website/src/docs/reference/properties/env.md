# env

Environment variables for task execution.

## Type

`env`


## Description

Set custom environment variables that will be available during task execution. Environment variables set at task level override global level, which override system environment.

Precedence order (highest to lowest):
1. Task-level env
2. Global env (root level)
3. System environment variables




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
version: '3'

tasks:
  greet:
    env:
      GREETING: Hey, there!
    cmds:
      - echo $GREETING

```



### Global environment

Available to all tasks

```yaml
version: '3'

env:
  NODE_ENV: production
  API_URL: https://api.example.com

tasks:
  build:
    cmds:
      - npm run build
  
  test:
    cmds:
      - npm test

```



### Build configuration

Cross-compilation settings

```yaml
tasks:
  build-linux:
    env:
      CGO_ENABLED: "0"
      GOOS: linux
      GOARCH: amd64
    cmds:
      - go build -o app-linux
  
  build-windows:
    env:
      GOOS: windows
      GOARCH: amd64
    cmds:
      - go build -o app.exe

```



### Overriding global env

Task-level overrides global

```yaml
version: '3'

env:
  KEYNAME: GLOBAL_VALUE

tasks:
  task1:
    cmds:
      - echo $KEYNAME  # Outputs: GLOBAL_VALUE
  
  task2:
    env:
      KEYNAME: DIFFERENT_VALUE
    cmds:
      - echo $KEYNAME  # Outputs: DIFFERENT_VALUE

```






## Related

- [vars](./vars.md)
- [dotenv](./dotenv.md)




