# Passing Data Between Tasks

## Problem
You want to generate a value in one task and use it in another.

## Solution
Use CLI arguments or environment variables.

### Using CLI Arguments

You can pass variables to another task when calling it using the `task` command or the `task:` syntax in `cmds`.

```yaml
tasks:
  build:
    cmds:
      - task: docker-build
        vars: { TAG: "v1.0.0" }

  docker-build:
    cmds:
      - docker build -t myapp:{{.TAG}} .
```

### Using Environment Variables

You can write to a file (like `.env`) in one task and read it in another using `dotenv`.

```yaml
tasks:
  generate-token:
    cmds:
      - echo "TOKEN=123" > .env

  deploy:
    deps: [generate-token]
    dotenv: [.env]
    cmds:
      - echo "Deploying with token $TOKEN"
```
