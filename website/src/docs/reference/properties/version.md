# version

Taskfile schema version.

## Type

`string`


## Description

Specify the Taskfile schema version. Required field. Current version is '3'.




## Contexts

This property can be used in:


- Taskfile (root level)








## Examples


### Version declaration

Declare schema version

```yaml
version: '3'

tasks:
  hello:
    cmds:
      - echo "Hello, World!"

```








