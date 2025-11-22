# platforms

Operating systems where task can run.

## Type

`platforms`


## Description

Restrict task or command execution to specific platforms (OS and/or architecture). Uses Go's GOOS and GOARCH values. On a platform mismatch, the task or command is skipped without error.

Can specify OS only (e.g., windows), architecture only (e.g., amd64), or both (e.g., windows/amd64). Multiple platforms can be specified.




## Contexts

This property can be used in:


- Task level

- Command level








## Examples


### OS-only restriction

Run only on Windows

```yaml
version: '3'

tasks:
  build-windows:
    platforms: [windows]
    cmds:
      - echo 'Running command on Windows'

```



### OS and architecture

Run on specific OS/arch combination

```yaml
version: '3'

tasks:
  build-windows-amd64:
    platforms: [windows/amd64]
    cmds:
      - echo 'Running command on Windows (amd64)'

```



### Architecture-only restriction

Run on specific architecture regardless of OS

```yaml
version: '3'

tasks:
  build-amd64:
    platforms: [amd64]
    cmds:
      - echo 'Running command on amd64'

```



### Multiple platforms

Run on several platforms

```yaml
version: '3'

tasks:
  build:
    platforms: [windows/amd64, darwin]
    cmds:
      - echo 'Running command on Windows (amd64) and macOS'

```



### Command-level platforms

Different commands for different platforms

```yaml
version: '3'

tasks:
  build:
    cmds:
      - cmd: echo 'Running command on Windows (amd64) and macOS'
        platforms: [windows/amd64, darwin]
      - cmd: echo 'Running on all platforms'

```






## Related

- [preconditions](./preconditions.md)




