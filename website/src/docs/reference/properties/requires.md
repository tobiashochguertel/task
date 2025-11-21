# requires

`requires` specifies variables that must be defined for the task to run.

## Type

`Requires`

## Properties

| Property | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `vars` | `[]string` or `[]Var` | List of required variables. | `vars: [VAR1]` |

### Var Object

| Property | Type | Description | Example |
| :--- | :--- | :--- | :--- |
| `name` | `string` | Name of the variable. | `name: ENV` |
| `enum` | `[]string` | Allowed values. | `enum: [prod, dev]` |

## Usage

If any required variable is missing, the task will fail.

### Simple List

```yaml
tasks:
  deploy:
    requires:
      vars: [API_KEY, ENV]
```

### With Enum Validation

```yaml
tasks:
  deploy:
    requires:
      vars:
        - name: ENV
          enum: [dev, prod]
```
