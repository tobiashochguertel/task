# method

Method for checking if task is up-to-date.

## Type

`string`


## Description

Determines how Task checks if re-execution is needed. Options: checksum, timestamp, none.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level






## Options

| Option | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `checksum` | `string` | Check content hash of source files (default). | `method: checksum` |
| `timestamp` | `string` | Check file modification times. | `method: timestamp` |
| `none` | `string` | Always run the task. | `method: none` |




## Examples


### Using checksums

Default method - check file content

```yaml
tasks:
  build:
    method: checksum
    sources:
      - src/**/*.go
    generates:
      - bin/app
    cmds:
      - go build

```



### Using timestamps

Faster but less accurate

```yaml
tasks:
  compile:
    method: timestamp
    sources:
      - src/**/*.ts
    generates:
      - dist/**/*.js
    cmds:
      - tsc

```






## Related

- [sources](./sources.md)
- [generates](./generates.md)
- [status](./status.md)




