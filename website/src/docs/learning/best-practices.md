# Best Practices

Here are some recommended patterns and practices for writing maintainable and efficient Taskfiles.

## 1. Use Descriptions

Always add a `desc` to your tasks. This makes `task --list` useful for other developers (and your future self).

```yaml
tasks:
  build:
    desc: Build the application
    cmds: ...
```

## 2. Use `sources` and `generates`

Whenever possible, define `sources` and `generates` for your tasks. This allows Task to skip unnecessary work (incremental builds).

```yaml
tasks:
  build:
    sources: ['*.go']
    generates: [app]
    cmds: [go build -o app]
```

## 3. Keep Tasks Small and Focused

Break down complex operations into smaller, reusable tasks. Use dependencies (`deps`) to compose them.

```yaml
tasks:
  deploy:
    deps: [lint, test, build]
    cmds: [./deploy.sh]
```

## 4. Use `defer` for Cleanup

If a task creates temporary files, use `defer` to ensure they are cleaned up even if the task fails.

```yaml
tasks:
  temp-work:
    cmds:
      - touch temp.txt
      - defer: rm temp.txt
      - ./do-work.sh
```

## 5. Organize with Includes

For large projects, split your `Taskfile.yml` into multiple files and use `includes`.

```yaml
includes:
  docker: ./taskfiles/docker.yml
  db: ./taskfiles/db.yml
```

## 6. Use Variables for Configuration

Avoid hardcoding values in commands. Use `vars` to make your tasks configurable.

```yaml
vars:
  BUILD_DIR: ./dist

tasks:
  build:
    cmds: [mkdir -p {{.BUILD_DIR}}]
```

## 7. Use `set: [errexit]`

It is generally a good idea to enable `errexit` (exit on error) to stop execution immediately if a command fails.

```yaml
set: [errexit]
```
