# method

`method` defines the default strategy used to check if a task is up-to-date (i.e., if it needs to run based on `sources` and `generates`).

## Type

`string`

## Options

| Option | Description | Example |
| :--- | :--- | :--- |
| `checksum` | Calculates a hash of the file contents. (Default) | `method: checksum` |
| `timestamp` | Checks the file modification time. | `method: timestamp` |
| `none` | Disables up-to-date checks. | `method: none` |

## Usage

```yaml
method: timestamp
```
