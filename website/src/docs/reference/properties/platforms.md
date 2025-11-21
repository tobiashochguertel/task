# platforms

Operating systems where task can run.

## Type

`platforms`


## Description

Restrict task execution to specific platforms. Task will skip if platform doesn't match.




## Contexts

This property can be used in:


- Task level








## Examples


### Linux only

Task runs only on Linux

```yaml
tasks:
  install-deps:
    platforms: [linux]
    cmds:
      - apt-get install -y build-essential

```



### Multiple platforms

Task runs on Linux or macOS

```yaml
tasks:
  build:
    platforms: [linux, darwin]
    cmds:
      - make build

```



### Windows specific

Task for Windows only

```yaml
tasks:
  setup:
    platforms: [windows]
    cmds:
      - choco install golang

```






## Related

- [preconditions](./preconditions.md)




