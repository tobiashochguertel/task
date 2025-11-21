# interval

Interval for watch mode polling.

## Type

`string`


## Description

Set the polling interval for watch mode. Default is 5 seconds.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level








## Examples


### Custom interval

Set watch polling interval

```yaml
version: '3'

interval: 2s

tasks:
  watch:
    watch: true
    sources:
      - src/**/*.go
    cmds:
      - go build

```






## Related

- [watch](./watch.md)




