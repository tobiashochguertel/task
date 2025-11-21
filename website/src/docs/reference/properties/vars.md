# vars

Variables for tasks and templates.

## Type

`vars`


## Description

Define reusable values and dynamic content. Variables can be strings, booleans, numbers, arrays, or maps. They are available in templates using {{.VAR_NAME}} syntax.

Variable precedence (most important first):
1. Variables declared in task definition
2. Variables given when calling a task
3. Variables from included Taskfile
4. Variables from inclusion declaration
5. Global variables (root level vars:)
6. Environment variables

A special variable .TASK is always available containing the current task name.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level

- CLI arguments

- Include level

- Task calls




## Precedence

Task-level variables → Call-time variables → Include variables → Global variables → Environment variables





## Examples


### Basic types

All supported variable types

```yaml
version: '3'

tasks:
  example:
    vars:
      STRING: 'Hello, World!'
      BOOL: true
      INT: 42
      FLOAT: 3.14
      ARRAY: [1, 2, 3]
      MAP:
        map: { A: 1, B: 2, C: 3 }
    cmds:
      - echo "{{.STRING}}"
      - echo "{{.BOOL}}"
      - echo "{{.INT}}"
      - echo "{{.FLOAT}}"
      - echo "{{.ARRAY}}"
      - echo "{{index .ARRAY 0}}"
      - echo "{{.MAP}}"
      - echo "{{.MAP.A}}"

```



### Static variables

Simple string variables

```yaml
vars:
  GREETING: Hello
  VERSION: 1.0.0

tasks:
  greet:
    cmds:
      - echo "{{.GREETING}} - v{{.VERSION}}"

```



### Dynamic variables

Execute shell commands to get values

```yaml
tasks:
  build:
    vars:
      GIT_COMMIT:
        sh: git log -n 1 --format=%h
      BUILD_TIME:
        sh: date +%Y%m%d
    cmds:
      - go build -ldflags="-X main.Version={{.GIT_COMMIT}}"

```



### CLI arguments

Pass variables from command line

```yaml
# Usage: task greet USER_NAME="Alice"
tasks:
  greet:
    cmds:
      - echo "Hello, {{.USER_NAME}}!"

```



### Global and local variables

Override global vars at task level

```yaml
version: '3'

vars:
  ENV: development

tasks:
  dev:
    cmds:
      - echo "Running in {{.ENV}}"
  
  prod:
    vars:
      ENV: production
    cmds:
      - echo "Running in {{.ENV}}"

```



### Special variables

Using built-in .TASK variable

```yaml
tasks:
  build:
    cmds:
      - echo "Executing task: {{.TASK}}"

```






## Related

- [env](./env.md)
- [dotenv](./dotenv.md)




