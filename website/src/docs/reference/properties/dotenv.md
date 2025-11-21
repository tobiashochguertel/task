# dotenv

Load environment variables from .env files.

## Type

`array`


## Description

Load environment variables from .env files before executing tasks. Supports multiple files and templating in file paths. Later files in the list override earlier ones.

Dotenv files can be specified at both global (root) and task level. Task-level dotenv files are loaded after global ones.




## Contexts

This property can be used in:


- Taskfile (root level)

- Task level








## Examples


### Basic dotenv

Load from .env file

```yaml
version: '3'

dotenv: ['.env']

tasks:
  deploy:
    cmds:
      - echo "Using API key: $API_KEY"

```



### Multiple env files with templating

Load from different locations

```yaml
version: '3'

env:
  ENV: testing

dotenv:
  - .env
  - '{{.ENV}}/.env'
  - '{{.HOME}}/.env'

tasks:
  greet:
    cmds:
      - echo "Using $KEYNAME and endpoint $ENDPOINT"

```



### Task-level dotenv

Different env files per task

```yaml
version: '3'

env:
  ENV: testing

tasks:
  greet:
    dotenv: ['{{.ENV}}/.env']
    cmds:
      - echo "Hello from $ENV environment"

```



### Environment-specific configs

Load based on environment variable

```yaml
version: '3'

tasks:
  test:
    dotenv: ['.env', '.env.test']
    cmds:
      - npm test
  
  production:
    dotenv: ['.env', '.env.production']
    cmds:
      - npm run deploy

```






## Related

- [env](./env.md)
- [vars](./vars.md)




