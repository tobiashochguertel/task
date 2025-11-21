# dotenv

Load environment variables from .env files.

## Type

`array`


## Description

Specify .env files to load before task execution. Variables are loaded in order.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level








## Examples


### Load .env file

Load environment from .env

```yaml
version: '3'

dotenv: ['.env']

tasks:
  deploy:
    cmds:
      - echo "Deploying to $ENV"

```



### Multiple env files

Load from multiple files

```yaml
tasks:
  build:
    dotenv:
      - .env
      - .env.local
    cmds:
      - npm run build

```



### Environment-specific

Load different files per environment

```yaml
tasks:
  start:
    dotenv:
      - .env
      - .env.{{.ENV}}
    cmds:
      - npm start

```






## Related

- [env](./env.md)
- [vars](./vars.md)




