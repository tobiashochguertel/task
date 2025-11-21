# dotenv

`dotenv` allows you to load environment variables from `.env` files.

## Type

`[]string`

## Usage

Task will read the specified files and export the variables defined in them to the environment.

```yaml
dotenv:
  - .env
  - .env.local
```

## Precedence

Variables defined in `env` take precedence over variables loaded from `dotenv` files.
Variables in later files in the list override variables in earlier files.
