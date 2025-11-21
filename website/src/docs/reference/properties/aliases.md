# aliases

Alternative names for the task.

## Type

`array`


## Description

Define multiple names for the same task. Useful for shortcuts or compatibility.




## Contexts

This property can be used in:


- Task level








## Examples


### Basic aliases

Add shortcuts for task names

```yaml
tasks:
  build-production:
    aliases: [bp, prod]
    cmds:
      - npm run build -- --mode production

```



### Multiple aliases

Task with several alternative names

```yaml
tasks:
  run-tests:
    aliases: [test, t, tests]
    cmds:
      - npm test

```








