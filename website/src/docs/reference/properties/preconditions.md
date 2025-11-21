# preconditions

Conditions that must be met before task runs.

## Type

`array`


## Description

Check conditions before executing task. Task fails if any precondition is not met.




## Contexts

This property can be used in:


- Task level








## Examples


### Check command exists

Ensure required tool is installed

```yaml
tasks:
  deploy:
    preconditions:
      - sh: which docker
        msg: "Docker is not installed"
    cmds:
      - docker build -t myapp .

```



### Check file exists

Verify required file is present

```yaml
tasks:
  build:
    preconditions:
      - sh: test -f config.yaml
        msg: "config.yaml not found"
    cmds:
      - go build

```



### Multiple preconditions

Check several conditions

```yaml
tasks:
  release:
    preconditions:
      - sh: git diff --quiet
        msg: "Git working directory is not clean"
      - sh: test -f CHANGELOG.md
        msg: "CHANGELOG.md is missing"
    cmds:
      - task: build
      - task: publish

```






## Related

- [requires](./requires.md)
- [platforms](./platforms.md)




