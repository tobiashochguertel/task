# sources

`sources` defines the files that the task depends on. It is used to determine if a task needs to be re-run.

## Type

`[]string` (Glob patterns)

## Usage

If a task has `sources` defined, Task will check if the files matching the patterns have changed since the last successful run.

```yaml
tasks:
  build:
    sources:
      - '**/*.go'
      - go.mod
      - exclude: '**/*_test.go'
    cmds:
      - go build ./...
```

### Exclusions

You can exclude files using the `exclude` keyword.

```yaml
sources:
  - ./**/*
  - exclude: .git/**/*
  - exclude: node_modules/**/*
```

## Methods

The method used to check for changes is defined by the global `method` property (default: `checksum`).

- **checksum**: Calculates a hash of the file contents.
- **timestamp**: Checks the file modification time.
- **none**: Always run the task.
