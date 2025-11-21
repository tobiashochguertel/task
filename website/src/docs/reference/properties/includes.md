# includes

Include external Taskfiles.

## Type

`map[string]Include`


## Description

Import tasks from other Taskfiles. Useful for splitting large Taskfiles or sharing tasks into smaller, more manageable files.




## Contexts

This property can be used in:


- Taskfile (root level)






## Options

| Option | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `taskfile` | `string` | Path to the Taskfile or directory to include. Required. | `taskfile: ./lib/tasks.yml` |
| `dir` | `string` | The working directory for the included tasks. | `dir: ./lib` |
| `optional` | `bool` | If true, no error is raised if the included file is missing. | `optional: true` |
| `flatten` | `bool` | If true, tasks are included at the root level (no namespace). | `flatten: true` |
| `internal` | `bool` | If true, included tasks are hidden from --list. | `internal: true` |
| `aliases` | `[]string` | Alternative names for the namespace. | `aliases: [api, v1]` |
| `excludes` | `[]string` | List of tasks to exclude from the included Taskfile. | `excludes: [setup]` |
| `vars` | `map[string]Variable` | Variables to pass to the included Taskfile. | `vars: { ENV: prod }` |
| `checksum` | `string` | Expected checksum of the included file (for remote includes). | `checksum: "sha256:..."` |




## Examples


### Simple string format

Include a Taskfile by path

```yaml
includes:
  docs: ./Taskfile.yml

```



### Full object format

Include with all available options

```yaml
includes:
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



### Basic include

Include tasks from another file

```yaml
version: '3'

includes:
  docker: ./docker/Taskfile.yml
  ci: ./ci/Taskfile.yml

tasks:
  deploy:
    cmds:
      - task: docker:build
      - task: docker:push

```



### Include with custom directory

Set working directory for included tasks

```yaml
includes:
  frontend:
    taskfile: ./frontend/Taskfile.yml
    dir: ./frontend

```



### Optional includes

Include file if it exists

```yaml
includes:
  local:
    taskfile: ./Taskfile.local.yml
    optional: true

```








