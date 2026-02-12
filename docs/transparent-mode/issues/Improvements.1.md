# Improvements - 1

## 0. Environment

```shell
❯ whereis task-dev
task-dev: /Users/tobiashochgurtel/go/bin/task-dev

❯ alias task-dev
task-dev='task-dev --transparent'

❯ task-dev --version
Task version: 3.48.0
Branch:       feature/transparent-mode
Built:        2026-02-12T09:28:12Z
Built by:     tobiashochgurtel
```

## 1. General Formatting

The output of the transparent mode is good, but we have to improve it.

```shell
~/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder main* ❯
❯ task-dev -v

══════ Transparent Mode Report ══════

── Global Variables
  Variables:
  Name              Origin          Type      Ref?    Value
  ──────────────────────────────────────────────────────
  ALIAS             special         string      ·     default
  ROOT_DIR          special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  ROOT_TASKFILE     special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  TASK              special         string      ·     default
  TASKFILE          special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  TASKFILE_DIR      special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  TASK_DIR          special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  TASK_EXE          special         string      ·     task-dev
  TASK_VERSION      special         string      ·     3.48.0
  USER_WORKING_DIR  special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  BUN_ARGS          taskfile-vars   string      ·     --bun
  CLI               taskfile-vars   string      ·      node   --experimental-strip-types  src/cli.ts
  CLI_ARGS          taskfile-vars   string      ·
  CLI_ARGS_LIST     taskfile-vars   []string    ·     []
                    ptr: 0x1035a9680
  CLI_ASSUME_YES    taskfile-vars   bool        ·     false
  CLI_FORCE         taskfile-vars   bool        ·     false
  CLI_OFFLINE       taskfile-vars   bool        ·     false
  CLI_SILENT        taskfile-vars   bool        ·     false
  CLI_VERBOSE       taskfile-vars   bool        ·     true
  ENGINE            taskfile-vars   string      ·      node
  ENGINE_ARGS       taskfile-vars   string      ·      --experimental-strip-types
  LAUNCH            taskfile-vars   string      ·      node   --experimental-strip-types  src/launch.ts
  NODE_ARGS         taskfile-vars   string      ·     --experimental-strip-types
  OUTPUT_DIR        taskfile-vars   string      ·     ./output
  OUT_DIR           taskfile-vars   string      ·     ./out
  SCHEMAS_DIR       taskfile-vars   string      ·     ./schemas

── Task: default
  Variables:
  Name   Origin          Type      Ref?    Value
  ───────────────────────────────────────────
  MATCH  call-vars       []string    ·     []
         ptr: 0x1035a9680
  Commands:
  [0] task --list
echo ""
echo "::Development::"
echo ""
echo "task setup"
echo ""


══════ End Report ══════
```

The column `Value` is somehow truncate, that should never happen. No column of the table should be truncate, otherwise we can't understand the output / report.
we should also add a additional cli option for the transparent mode which makes white spaces visibile and understandable via a legend and coloring might be also good.

The above shown output is from the following project: `/Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder`if you need to playaround or verify something...

Here is a more advanced task output for the task `debug`:

```shell
❯ task-dev -v debug

══════ Transparent Mode Report ══════

── Global Variables
  Variables:
  Name              Origin          Type      Ref?    Value
  ──────────────────────────────────────────────────────
  ALIAS             special         string      ·     debug
  ROOT_DIR          special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  ROOT_TASKFILE     special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  TASK              special         string      ·     debug
  TASKFILE          special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  TASKFILE_DIR      special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  TASK_DIR          special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  TASK_EXE          special         string      ·     task-dev
  TASK_VERSION      special         string      ·     3.48.0
  USER_WORKING_DIR  special         string      ·     /Users/tobiashochgurtel/work-dev/temp-projects/compare-vs...
  BUN_ARGS          taskfile-vars   string      ·     --bun
  CLI               taskfile-vars   string      ·      node   --experimental-strip-types  src/cli.ts
  CLI_ARGS          taskfile-vars   string      ·
  CLI_ARGS_LIST     taskfile-vars   []string    ·     []
                    ptr: 0x102ce9680
  CLI_ASSUME_YES    taskfile-vars   bool        ·     false
  CLI_FORCE         taskfile-vars   bool        ·     false
  CLI_OFFLINE       taskfile-vars   bool        ·     false
  CLI_SILENT        taskfile-vars   bool        ·     false
  CLI_VERBOSE       taskfile-vars   bool        ·     true
  ENGINE            taskfile-vars   string      ·      node
  ENGINE_ARGS       taskfile-vars   string      ·      --experimental-strip-types
  LAUNCH            taskfile-vars   string      ·      node   --experimental-strip-types  src/launch.ts
  NODE_ARGS         taskfile-vars   string      ·     --experimental-strip-types
  OUTPUT_DIR        taskfile-vars   string      ·     ./output
  OUT_DIR           taskfile-vars   string      ·     ./out
  SCHEMAS_DIR       taskfile-vars   string      ·     ./schemas

── Task: debug
  Variables:
  Name            Origin          Type      Ref?    Value
  ────────────────────────────────────────────────────
  MATCH           call-vars       []string    ·     []
                  ptr: 0x102ce9680
  COLOR_AQUA      task-vars       string      ·     33[38;5;87m
  COLOR_BLACK     task-vars       string      ·     33[30m
  COLOR_BLUE      task-vars       string      ·     33[34m
  COLOR_BOLD      task-vars       string      ·     33[1m
  COLOR_BROWN     task-vars       string      ·     33[38;5;130m
  COLOR_CYAN      task-vars       string      ·     33[36m
  COLOR_DIM       task-vars       string      ·     33[2m
  COLOR_GRAY      task-vars       string      ·     33[90m
  COLOR_GREEN     task-vars       string      ·     33[32m
  COLOR_INDIGO    task-vars       string      ·     33[38;5;63m
  COLOR_LAVENDER  task-vars       string      ·     33[38;5;183m
  COLOR_LIME      task-vars       string      ·     33[38;5;154m
  COLOR_MAGENTA   task-vars       string      ·     33[35m
  COLOR_MAUVE     task-vars       string      ·     33[38;5;135m
  COLOR_ORANGE    task-vars       string      ·     33[38;5;208m
  COLOR_PINK      task-vars       string      ·     33[38;5;213m
  COLOR_PLUM      task-vars       string      ·     33[38;5;93m
  COLOR_PURPLE    task-vars       string      ·     33[38;5;170m
  COLOR_RED       task-vars       string      ·     33[31m
  COLOR_RESET     task-vars       string      ·     33[0m
  COLOR_ROSE      task-vars       string      ·     33[38;5;211m
  COLOR_SALMON    task-vars       string      ·     33[38;5;217m
  COLOR_SKY       task-vars       string      ·     33[38;5;117m
  COLOR_TAN       task-vars       string      ·     33[38;5;180m
  COLOR_TEAL      task-vars       string      ·     33[38;5;37m
  COLOR_VIOLET    task-vars       string      ·     33[38;5;127m
  COLOR_WHITE     task-vars       string      ·     33[37m
  COLOR_YELLOW    task-vars       string      ·     33[33m
  SPACE           task-vars       int         ·     20
  Template Evaluations:
  [1] (cmds[0]) Input:  echo ":: Global Taskfile variables ::"
echo ""
echo "{{spew (.ENGINE | trim)}}"
echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'

       Output: echo ":: Global Taskfile variables ::"
echo ""
echo "(string) (len=4) "node"
"
echo 'ENGINE:                 node %!s(MISSING)'

       Vars used: ENGINE, SPACE
  Commands:
  [0] raw:      echo ":: Global Taskfile variables ::"
echo ""
echo "{{spew (.ENGINE | trim)}}"
echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'

       resolved: echo ":: Global Taskfile variables ::"
echo ""
echo "(string) (len=4) "node"
"
echo 'ENGINE:                 node %!s(MISSING)'


══════ End Report ══════
```

## 2. Template Evaluation Steps

Here is the current output format:

```shell
  Template Evaluations:
  [1] (cmds[0]) Input:  echo ":: Global Taskfile variables ::"
echo ""
echo "{{spew (.ENGINE | trim)}}"
echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'

       Output: echo ":: Global Taskfile variables ::"
echo ""
echo "(string) (len=4) "node"
"
echo 'ENGINE:                 node %!s(MISSING)'

       Vars used: ENGINE, SPACE
  Commands:
  [0] raw:      echo ":: Global Taskfile variables ::"
echo ""
echo "{{spew (.ENGINE | trim)}}"
echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'

       resolved: echo ":: Global Taskfile variables ::"
echo ""
echo "(string) (len=4) "node"
"
echo 'ENGINE:                 node %!s(MISSING)'
```

We can improve the output to show it like this:

```shell
  Template Evaluation — cmds[0]:
  ┌─ Input:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "{{spew (.ENGINE | trim)}}"
  │ echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'
  └─
  ┌─ Output:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "(string) (len=4) "node"
  | "
  │ echo 'ENGINE:                 node %!s(MISSING)'
  └─
  ┌─ Vars used:
  │ ENGINE, SPACE
  └─

  Commands — cmds[0]:
  ┌─ Raw:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "{{spew (.ENGINE | trim)}}"
  │ echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'
  └─
  ┌─ resolved:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "(string) (len=4) "node"
  | "
  │ echo 'ENGINE:                 node %!s(MISSING)'
  └─
```

**The changes are:**

- Added a header `Template Evaluation` and `Commands` to separate the sections
  - The header shows the command index in the format `cmds[index]`
  - The header is separated from the content by a blank line
  - The header ends with a colon `:`
  - The header content is indented by 2 spaces
  - The header content is left-aligned
  - The header gets repeated for each command in the sequence `cmds[index]`
- Added separators `┌─` and `└─` to separate the sections
  - The opening separator `┌─` is aligned with the header content
    - The opening separator shows the section name in the format `Section Name` like `Input:` or `Output:` or `Vars used:` or `Raw:` or `resolved:` followed by a colon `:`
  - The closing separator `└─` is aligned with the header content
  - The separators are indented by 2 spaces
  - The separators are left-aligned
  - The separators are on the same line as the header content
- Added border around the content of each section
  - The border is made of `│` characters
  - The border is made of `│` characters on only the left side of the content
    - The border starts at the same indentation level as the header content
    - There is no border on the right side of the content, so the content can extend to the right
  - The border is indented by 2 spaces
  - The border is left-aligned
  - The border is on the same line as the header content
- Added `Input:` and `Output:` labels to show the input and output of each command
- Added `Vars used:` to show the variables used in each command
- Added `Raw:` and `resolved:` labels to show the raw and resolved commands
- Ensure that we not trim the "Raw" and "resolved" or the "Input" and "Output" content of the commands (keep all whitespace)
- Integrate the option to make Whitespaces visible in the output (e.g. use · for spaces and → for tabs)... as a toggle via CLI options (e.g. `--show-whitespaces`).
- Remove the square brackets and parentheses from the command numbers (e.g. [0] -> cmds[0])
- Ensure that the output is properly formatted and readable

**What I'm missing from the specifications:**

1. The note section should be displayed as a separate section with a header "Note:" and the content below it.

```
  ℹ Note: If you intended to trim .NAME before printf, use:
    {{printf "%s : %s" .GREETING (.NAME | trim)}}
```

see `## Human-Readable Output Example` in `/Users/tobiashochgurtel/work-dev/my-projects/task/docs/transparent-mode/05-OUTPUT-FORMAT.md` for more details.

1. In the `Template Evaluation` section, we should show the steps of the template evaluation, like in the example:

```yaml
        echo ":: Global Taskfile variables ::"
        echo ""
        echo "{{spew (.ENGINE | trim)}}"
        echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'
```

```
  Template Evaluation — cmds[0]:
  ┌─ Input:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "{{spew (.ENGINE | trim)}}"
  │ echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'
  └─
  ┌─ Steps:
  │ Step 1: Resolve a Variable (`.ENGINE`)
  |   I     "node"
  |         "
  |   F     echo "{{spew ("node"
  |         " | trim)}}"
  │ Step 2: Apply a Function (`trim`)
  |   I     trim "node"
  |         "
  |   O     "node"
  |         "
  |   F     echo "{{spew "node"
  |         "}}"
  | Step 3: Apply a Function (`spew`)
  |   I     spew "node"
  |         "
  |   O     "(string) (len=4) "node"
  |         "
  |   F     echo "(string) (len=4) "node"
  |         "
  | Step 4: Resolve a Variable (`.SPACE`)
  |   I     20
  |   F     echo '{{printf "%s: %*s %s" "ENGINE" 20 (.ENGINE | trim)}}'
  | Step 5: Resolve a Variable (`.ENGINE`)
  |   I     "node"
  |         "
  |   F     echo '{{printf "%s: %*s %s" "ENGINE" 20 ("node"
  |         " | trim)}}'
  | Step 6: Apply a Function (`trim`)
  |   I     trim "node"
  |         "
  |   O     "node"
  |         "
  |   F     echo '{{printf "%s: %*s %s" "ENGINE" 20 "node"
  |         "}}'
  | Step 7: Apply a Function (`printf`)
  |   I     printf "%s: %*s %s" "ENGINE" 20 "node"
  |         "
  |   O     ENGINE:                 node %!s(MISSING)
  |   F     echo 'ENGINE:                 node %!s(MISSING)'
  └─
  ┌─ Output:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "(string) (len=4) "node"
  | "
  │ echo 'ENGINE:                 node %!s(MISSING)'
  └─
  ┌─ Vars used:
  │ ENGINE, SPACE
  └─
```

**Explanation:**
Each step shows the input, the operation performed, the output, and the final command that gets executed.

- `I` shows the input to the operation
- `O` shows the output of the operation
- `F` shows the final command that gets executed
- The steps are numbered sequentially

The Step template looks like this:

```template
│ Step N: Operation Name
|   I     input
|   O     output
|   F     final command
```

Example:

```rendered
  │ Step 1: Resolve a Variable (`.ENGINE`)
  |   I     "node"
  |         "
  |   F     echo "{{spew ("node"
  |         " | trim)}}"
```

- `|` is the separator between the step number and the operation name
- `I` is the separator between the input and the operation
- `O` is the separator between the operation and the output
- `F` is the separator between the output and the final command
- Each line is prefixed with `|` to indicate that it is part of the step
- The step number is always followed by a colon and a space (e.g., `1:`)
- The operation name is always followed by a colon and a space (e.g., `Resolve a Variable:`)
- The input, output, and final command are always on the same line and aligned at the same column position (e.g., `|   I     input`, `|   O     output`, `|   F     final command`)
- Integrate the option to make Whitespaces visible in the output (e.g. use · for spaces and → for tabs)... as a toggle via CLI options (e.g. `--show-whitespaces`).

The template evaluation shows how the template is evaluated step by step, with the input, the operation, the output, and the final command that gets executed.

see `## Human-Readable Output Example` in `/Users/tobiashochgurtel/work-dev/my-projects/task/docs/transparent-mode/05-OUTPUT-FORMAT.md` for more details.

**Showing Errors**

When there are errors in the template evaluation, we should highlight them like `%!s(MISSING)` from the output with a red background or style to make them more visible.

For example, in the output above, the error is shown as `ENGINE:                 node %!s(MISSING)` where `%!s(MISSING)` is the error.

We should provide a hint to the user about the function signature that the used function expects, so they can fix the error, for example by showing the function signature in a tooltip or a separate section below the output.
We want to provide more visibility into how the variables are evaluated and what their values are at each step.
