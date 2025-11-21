# shopt

`shopt` allows you to enable Bash-specific shell options.

## Type

`[]string`

## Options

| Option | Description | Example |
| :--- | :--- | :--- |
| `expand_aliases` | Aliases are expanded. | `shopt: [expand_aliases]` |
| `globstar` | Enable the `**` pattern for recursive globbing. | `shopt: [globstar]` |
| `nullglob` | If a pattern matches no files, it expands to a null string rather than itself. | `shopt: [nullglob]` |

## Usage

```yaml
shopt: [globstar]
```
