# vars

Global variables available to all tasks.

## Type

`vars`


## Description

Variables allow you to define reusable values and dynamic content.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level

- CLI arguments




## Precedence

CLI arguments → Task level → Taskfile level





## Examples


### Static variables

Define simple string variables

```yaml
vars:
  GREETING: Hello
  VERSION: 1.0.0

```



### Dynamic variables

Execute shell commands

```yaml
vars:
  COMMIT:
    sh: git rev-parse HEAD
  BUILD_TIME:
    sh: date +%Y%m%d

```






## Related

- [env](./env.md)
- [dotenv](./dotenv.md)




