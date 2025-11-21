# watch

Enable watch mode for continuous execution.

## Type

`boolean`


## Description

When true, task watches for file changes and re-runs automatically.




## Contexts

This property can be used in:


- Task level








## Examples


### Watch mode

Auto-rebuild on changes

```yaml
tasks:
  dev:
    watch: true
    sources:
      - src/**/*.ts
    cmds:
      - tsc
      - node dist/index.js

```



### Development workflow

Combined watch tasks

```yaml
tasks:
  watch-all:
    deps:
      - task: watch-backend
      - task: watch-frontend
  
  watch-backend:
    watch: true
    dir: backend
    sources:
      - "**/*.go"
    cmds:
      - go build
      - ./app
  
  watch-frontend:
    watch: true
    dir: frontend
    sources:
      - "**/*.ts"
    cmds:
      - npm run build

```






## Related

- [sources](./sources.md)




