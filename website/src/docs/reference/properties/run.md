# run

Control when task should run.

## Type

`run`


## Description

Specify task execution behavior. Options: always, once, when_changed.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level






## Options

| Option | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `always` | `string` | Always run the task (default). | `run: always` |
| `once` | `string` | Run only once per invocation. | `run: once` |
| `when_changed` | `string` | Run when sources change. | `run: when_changed` |




## Examples


### Run once

Task runs only once even if called multiple times

```yaml
tasks:
  setup:
    run: once
    cmds:
      - npm install
  
  build:
    deps: [setup]
    cmds:
      - npm run build
  
  test:
    deps: [setup]
    cmds:
      - npm test

```



### Run when changed

Run only if sources changed

```yaml
tasks:
  compile:
    run: when_changed
    sources:
      - src/**/*.ts
    generates:
      - dist/**/*.js
    cmds:
      - tsc

```






## Related

- [method](./method.md)
- [sources](./sources.md)




