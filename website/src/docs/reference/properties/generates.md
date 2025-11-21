# generates

`generates` defines the files that are created or updated by the task.

## Type

`[]string` (Glob patterns)

## Usage

`generates` is used in conjunction with `sources` to determine if a task is up-to-date. If the generated files exist and are newer than the source files (or have matching checksums), the task is considered up-to-date and will be skipped.

```yaml
tasks:
  build:
    sources:
      - '**/*.go'
    generates:
      - ./app
    cmds:
      - go build -o app ./...
```

### Exclusions

You can exclude files from the check using the `exclude` keyword.

```yaml
generates:
  - ./dist/**/*
  - exclude: ./dist/*.tmp
```
