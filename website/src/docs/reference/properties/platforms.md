# platforms

`platforms` specifies which operating systems the task or command should run on.

## Type

`[]string`

## Common Values

| Platform | Description | Example |
| :--- | :--- | :--- |
| `linux` | Linux | `platforms: [linux]` |
| `darwin` | macOS | `platforms: [darwin]` |
| `windows` | Windows | `platforms: [windows]` |

## Usage

If the current operating system is not in the list, the task or command will be skipped.

```yaml
tasks:
  windows-only:
    platforms: [windows]
    cmd: echo "This is Windows"

  unix-only:
    platforms: [linux, darwin]
    cmd: echo "This is Unix"
```
