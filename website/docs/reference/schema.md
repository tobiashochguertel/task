# Taskfile Schema

This document describes the schema of the `Taskfile.yml` file.

## Version

The `version` key specifies the version of the Taskfile schema.

```yaml
version: '3'
```

## Includes

You can include other Taskfiles using the `includes` key.

```yaml
includes:
  docs: ./documentation/Taskfile.yml
```

## Tasks

The `tasks` key is where you define your tasks.

```yaml
tasks:
  build:
    cmds:
      - go build -v -i main.go
```
