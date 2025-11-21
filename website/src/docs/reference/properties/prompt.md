# prompt

Show confirmation prompt before executing.

## Type

`union`


## Description

Display a yes/no prompt before running the task. Useful for dangerous operations.




## Contexts

This property can be used in:


- Task level








## Examples


### Simple prompt

Ask for confirmation

```yaml
tasks:
  delete-db:
    prompt: This will delete the database. Continue?
    cmds:
      - rm -rf data/db

```



### Destructive operation

Protect production deployment

```yaml
tasks:
  deploy-prod:
    prompt: Deploy to PRODUCTION environment?
    env:
      ENV: production
    cmds:
      - task: build
      - task: push
      - task: update-service

```








