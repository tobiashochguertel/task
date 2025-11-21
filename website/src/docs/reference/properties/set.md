# set

`set` allows you to enable POSIX shell options for the commands executed by Task.

## Type

`[]string`

## Options

| Option | Short | Description | Example |
| :--- | :--- | :--- | :--- |
| `allexport` | `a` | Export all variables. | `set: [allexport]` |
| `errexit` | `e` | Exit immediately if a command exits with a non-zero status. | `set: [errexit]` |
| `noexec` | `n` | Read commands but do not execute them. | `set: [noexec]` |
| `noglob` | `f` | Disable pathname expansion (globbing). | `set: [noglob]` |
| `nounset` | `u` | Treat unset variables as an error. | `set: [nounset]` |
| `xtrace` | `x` | Print commands and their arguments as they are executed. | `set: [xtrace]` |
| `pipefail` | - | The return value of a pipeline is the status of the last command to exit with a non-zero status. | `set: [pipefail]` |

## Usage

```yaml
set: [errexit, nounset, pipefail]
```
