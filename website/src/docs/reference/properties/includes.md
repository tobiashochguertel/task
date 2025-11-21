# includes

`includes` allows you to include other Taskfiles into your main Taskfile. This is useful for splitting a large Taskfile into smaller, more manageable files.

## Type

`map[string]Include`

## Syntax

```yaml
includes:
  # Simple string format (path to Taskfile)
  docs: ./Taskfile.yml

  # Full object format
  backend:
    taskfile: ./backend
    dir: ./backend
    optional: false
    flatten: false
    internal: false
    aliases: [api]
    excludes: [internal-task]
    vars:
      SERVICE_NAME: backend
```

## Properties

| Property | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `taskfile` | `string` | Path to the Taskfile or directory to include. Required. | `taskfile: ./lib/tasks.yml` |
| `dir` | `string` | The working directory for the included tasks. | `dir: ./lib` |
| `optional` | `bool` | If `true`, no error is raised if the included file is missing. | `optional: true` |
| `flatten` | `bool` | If `true`, tasks are included at the root level (no namespace). | `flatten: true` |
| `internal` | `bool` | If `true`, included tasks are hidden from `--list`. | `internal: true` |
| `aliases` | `[]string` | Alternative names for the namespace. | `aliases: [api, v1]` |
| `excludes` | `[]string` | List of tasks to exclude from the included Taskfile. | `excludes: [setup]` |
| `vars` | `map[string]Variable` | Variables to pass to the included Taskfile. | `vars: { ENV: prod }` |
| `checksum` | `string` | Expected checksum of the included file (for remote includes). | `checksum: "sha256:..."` |
