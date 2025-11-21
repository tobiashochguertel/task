# status

Shell commands to check if task is up-to-date.

## Type

`array`


## Description

Execute commands to determine if task needs to run. Task skips if all commands succeed.




## Contexts

This property can be used in:


- Task level








## Examples


### Check file exists

Skip if output file exists

```yaml
tasks:
  build:
    status:
      - test -f bin/app
    cmds:
      - go build -o bin/app

```



### Compare timestamps

Check if sources are newer

```yaml
tasks:
  compile:
    status:
      - test dist/bundle.js -nt src/index.js
    cmds:
      - webpack build

```



### Multiple checks

All conditions must pass to skip

```yaml
tasks:
  setup:
    status:
      - test -d node_modules
      - test -f package-lock.json
    cmds:
      - npm install

```






## Related

- [sources](./sources.md)
- [generates](./generates.md)
- [method](./method.md)




