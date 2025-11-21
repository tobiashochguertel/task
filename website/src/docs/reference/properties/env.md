# env

`env` defines global environment variables that are available to all tasks and commands.

## Type

`map[string]Variable`

## Usage

Environment variables defined in `env` are exported to the shell environment where commands are executed.

```yaml
env:
  NODE_ENV: production
  DATABASE_URL: postgres://user:pass@localhost:5432/db
```

### Dynamic Environment Variables

You can also use shell commands to set environment variables.

```yaml
env:
  GOPATH:
    sh: go env GOPATH
```
