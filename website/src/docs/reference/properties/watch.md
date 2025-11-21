# watch

`watch` enables watch mode for the task by default.

## Type

`bool`

## Default

`false`

## Usage

When set to `true`, running the task will automatically start watching for changes in its `sources` and re-run the task when they change.

```yaml
tasks:
  dev:
    watch: true
    sources: ['**/*.go']
    cmds: ...
```
