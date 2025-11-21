# summary

`summary` provides a detailed description of the task.

## Type

`string`

## Usage

This description is displayed when running `task --summary <task>`. It is useful for providing documentation, usage examples, or explaining complex tasks.

```yaml
tasks:
  deploy:
    summary: |
      Deploys the application to the production environment.
      
      Usage: task deploy [vars...]
    cmds: ...
```
