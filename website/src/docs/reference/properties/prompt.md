# prompt

`prompt` requires user confirmation before executing the task.

## Type

`string` or `[]string`

## Usage

If the user denies the prompt (by typing `n` or `no`), the task execution is aborted.

### Single Prompt

```yaml
tasks:
  clean:
    prompt: Are you sure you want to delete all files?
    cmds: ...
```

### Multiple Prompts

```yaml
tasks:
  deploy:
    prompt:
      - This will deploy to production.
      - Are you absolutely sure?
    cmds: ...
```
