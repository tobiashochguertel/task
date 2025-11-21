# cmds

Commands to execute for this task.

## Type

`cmds`


## Description

A list of shell commands or Task calls to execute. Commands run sequentially.




## Contexts

This property can be used in:


- Task level








## Examples


### Simple commands

Run shell commands

```yaml
tasks:
  build:
    cmds:
      - go build -o app
      - chmod +x app

```



### Task calls

Call other tasks

```yaml
tasks:
  build:
    cmds:
      - task: clean
      - task: compile

```



### Mixed commands

Mix shell and task calls

```yaml
tasks:
  deploy:
    cmds:
      - echo "Building..."
      - task: build
      - echo "Deploying..."
      - ./deploy.sh

```






## Related

- [cmd](./cmd.md)
- [deps](./deps.md)




