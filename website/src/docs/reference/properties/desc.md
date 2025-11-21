# desc

Short description shown in task list.

## Type

`string`


## Description

A brief description of the task. Displayed when running `task --list`.




## Contexts

This property can be used in:


- Task level








## Examples


### Basic description

Add description to task

```yaml
tasks:
  build:
    desc: Build the application
    cmds:
      - go build

```



### Hidden tasks

Tasks without desc are hidden from list

```yaml
tasks:
  public-task:
    desc: This appears in task --list
    cmds:
      - echo "visible"
  
  private-task:
    # No desc - hidden from list
    cmds:
      - echo "hidden"

```






## Related

- [summary](./summary.md)




