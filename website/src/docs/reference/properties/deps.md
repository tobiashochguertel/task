# deps

Task dependencies that run before this task.

## Type

`deps`


## Description

Define tasks that must run before the current task. Dependencies run in parallel by default, making builds faster. If tasks need to run serially, use task calls in cmds instead.

A task can have only dependencies and no commands to group tasks together. Dependencies of a task should not depend on each other since they run concurrently.




## Contexts

This property can be used in:


- Task level








## Examples


### Simple dependencies

Dependencies run in parallel before task

```yaml
version: '3'

tasks:
  build:
    deps: [assets]
    cmds:
      - go build -v -i main.go
  
  assets:
    cmds:
      - esbuild --bundle --minify css/index.css > public/bundle.css

```



### Multiple dependencies

All deps run concurrently

```yaml
tasks:
  build:
    deps: [clean, install, lint]
    cmds:
      - go build

```



### Grouping tasks

Task with only dependencies, no commands

```yaml
tasks:
  assets:
    deps: [js, css, images]
  
  js:
    cmds:
      - esbuild --bundle js/index.js > public/bundle.js
  
  css:
    cmds:
      - sass styles/main.scss public/main.css
  
  images:
    cmds:
      - imagemin src/images/* --out-dir=public/images

```



### Dependencies with variables

Pass variables to dependency tasks

```yaml
tasks:
  greet:
    vars:
      RECIPIENT: '{{default "World" .RECIPIENT}}'
    cmds:
      - echo "Hello, {{.RECIPIENT}}!"
  
  greet-pessimistically:
    deps:
      - task: greet
        vars: { RECIPIENT: 'Cruel World' }
    cmds:
      - echo "Greeting sent"

```



### Dependencies with silent mode

Control output of dependency execution

```yaml
tasks:
  deploy:
    deps:
      - task: build
        vars:
          ENV: production
        silent: true
    cmds:
      - ./deploy.sh

```



### Calling root tasks from includes

Reference root Taskfile tasks from included files

```yaml
# In included Taskfile
tasks:
  local-task:
    deps:
      - :root-task  # Leading colon calls root Taskfile task
    cmds:
      - echo "Local task"

```






## Related

- [cmds](./cmds.md)




