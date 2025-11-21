# sources

Source files to check for staleness.

## Type

`array`


## Description

List of source files or glob patterns. Used with method to determine if task needs to run.




## Contexts

This property can be used in:


- Task level








## Examples


### Basic sources

Track source files

```yaml
tasks:
  build:
    sources:
      - src/**/*.go
      - go.mod
    generates:
      - bin/app
    cmds:
      - go build -o bin/app

```



### With exclusions

Use glob patterns with exclusions

```yaml
tasks:
  compile:
    sources:
      - "src/**/*.ts"
      - "!src/**/*.test.ts"
    generates:
      - dist/**/*.js
    cmds:
      - tsc

```






## Related

- [generates](./generates.md)
- [method](./method.md)
- [status](./status.md)




