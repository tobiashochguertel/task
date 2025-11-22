# watch

Enable watch mode for continuous execution.

## Type

`boolean`


## Description

When set to true, the task automatically runs in watch mode, monitoring source files and re-executing on changes. Requires the sources attribute to be defined. Only runs in watch mode when invoked directly from CLI, not when called as a dependency.

The default watch interval is 100ms, but can be customized globally with interval setting or --interval flag. The interval is the debounce time - Task waits this long for duplicate events and only runs once.




## Contexts

This property can be used in:


- Task level








## Examples


### Basic watch mode

Auto-rebuild on file changes

```yaml
version: '3'

tasks:
  build:
    desc: Builds the Go application
    watch: true
    sources:
      - '**/*.go'
    cmds:
      - go build

```



### Watch with custom interval

Set global watch debounce interval

```yaml
version: '3'

interval: 500ms

tasks:
  build:
    desc: Builds the Go application
    watch: true
    sources:
      - '**/*.go'
    cmds:
      - go build

```



### Frontend development

Watch TypeScript files

```yaml
version: '3'

tasks:
  dev:
    watch: true
    sources:
      - 'src/**/*.ts'
      - 'src/**/*.tsx'
    cmds:
      - npm run build

```



### Multiple watch tasks

Different watch patterns per task

```yaml
version: '3'

interval: 300ms

tasks:
  watch-backend:
    watch: true
    sources:
      - 'backend/**/*.go'
    cmds:
      - go build ./backend
  
  watch-frontend:
    watch: true
    sources:
      - 'frontend/**/*.{ts,tsx,css}'
    cmds:
      - npm run build:frontend

```






## Related

- [sources](./sources.md)
- [interval](./interval.md)




