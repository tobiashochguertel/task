# dir

Working directory for task execution.

## Type

`string`


## Description

Specify the directory where commands should be executed. Relative to Taskfile location.




## Contexts

This property can be used in:


- Task level








## Examples


### Change directory

Run commands in specific directory

```yaml
tasks:
  frontend:
    dir: ./web
    cmds:
      - npm install
      - npm run build

```



### Using variables

Dynamic directory paths

```yaml
vars:
  PROJECT_DIR: ./projects/myapp

tasks:
  build:
    dir: '{{.PROJECT_DIR}}'
    cmds:
      - make build

```






## Related

- [cmds](./cmds.md)




