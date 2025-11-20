# Environment Variables

Task supports environment variables in various ways.

## .env files

Task automatically loads environment variables from `.env` files.

## env: key

You can define environment variables in the `Taskfile.yml`.

```yaml
env:
  FOO: bar
```

## Precedence

1.  Task-level `env`
2.  Taskfile-level `env`
3.  `.env` file
4.  System environment variables
