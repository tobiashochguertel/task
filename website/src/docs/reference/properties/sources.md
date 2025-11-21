# sources

Source files to check for staleness.

## Type

`array`


## Description

Glob patterns for input files that Task monitors to determine if a task needs to run. Task compares checksums (or timestamps with method: timestamp) of source files against generated files. If sources haven't changed, Task skips execution and shows "Task is up to date".

Use exclude: to ignore specific files from patterns. Sources are evaluated in order, so exclude must come after the positive glob it negates.




## Contexts

This property can be used in:


- Task level








## Examples


### Basic sources and generates

Skip task if sources unchanged

```yaml
version: '3'

tasks:
  js:
    sources:
      - src/js/**/*.js
    generates:
      - public/bundle.js
    cmds:
      - esbuild --bundle --minify js/index.js > public/bundle.js

```



### Multiple file patterns

Watch several file types

```yaml
tasks:
  build:
    sources:
      - src/**/*.go
      - go.mod
      - go.sum
    generates:
      - bin/app
    cmds:
      - go build -o bin/app

```



### Excluding files

Ignore specific files from patterns

```yaml
tasks:
  css:
    sources:
      - src/css/**/*.css
      - exclude: src/css/vendor/**
      - exclude: src/css/test.css
    generates:
      - public/bundle.css
    cmds:
      - sass --style=compressed src/css/main.scss public/bundle.css

```



### With multiple generates

One source produces multiple outputs

```yaml
tasks:
  compile:
    sources:
      - src/**/*.ts
    generates:
      - dist/**/*.js
      - dist/**/*.d.ts
    cmds:
      - tsc

```






## Related

- [generates](./generates.md)
- [method](./method.md)
- [status](./status.md)




