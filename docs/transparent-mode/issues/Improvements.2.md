# Improvements - 2

**the output is not yet correct:**

```shell
❯ task-dev --show-whitespaces -v debug

╔════════════════════════════════════════════════════════╗
║  TRANSPARENT MODE — Variable & Template Diagnostics  ║
╚════════════════════════════════════════════════════════╝
Legend: · = space, → = tab

── Global Variables
  Variables in scope:
  ┌──────────────────┬───────────────┬──────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┬──────────┐
  │ Name             │ Origin        │ Type     │ Value                                                                                                                 │ Shadows? │
  ├──────────────────┼───────────────┼──────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┼──────────┤
  │ ALIAS            │ special       │ string   │ debug                                                                                                                 │          │
  │ ROOT_DIR         │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder              │          │
  │ ROOT_TASKFILE    │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder/Taskfile.yml │          │
  │ TASK             │ special       │ string   │ debug                                                                                                                 │          │
  │ TASKFILE         │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder/Taskfile.yml │          │
  │ TASKFILE_DIR     │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder              │          │
  │ TASK_DIR         │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder              │          │
  │ TASK_EXE         │ special       │ string   │ task-dev                                                                                                              │          │
  │ TASK_VERSION     │ special       │ string   │ 3.48.0                                                                                                                │          │
  │ USER_WORKING_DIR │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder              │          │
  │ BUN_ARGS         │ taskfile-vars │ string   │ --bun                                                                                                                 │          │
  │ CLI              │ taskfile-vars │ string   │ ·node···--experimental-strip-types··src/cli.ts                                                                        │          │
  │ CLI_ARGS         │ taskfile-vars │ string   │                                                                                                                       │          │
  │ CLI_ARGS_LIST    │ taskfile-vars │ []string │ []                                                                                                                    │          │
  │                  │               │          │ ptr: 0x1070a1600                                                                                              │          │
  │ CLI_ASSUME_YES   │ taskfile-vars │ bool     │ false                                                                                                                 │          │
  │ CLI_FORCE        │ taskfile-vars │ bool     │ false                                                                                                                 │          │
  │ CLI_OFFLINE      │ taskfile-vars │ bool     │ false                                                                                                                 │          │
  │ CLI_SILENT       │ taskfile-vars │ bool     │ false                                                                                                                 │          │
  │ CLI_VERBOSE      │ taskfile-vars │ bool     │ true                                                                                                                  │          │
  │ ENGINE           │ taskfile-vars │ string   │ ·node·                                                                                                                │          │
  │ ENGINE_ARGS      │ taskfile-vars │ string   │ ·--experimental-strip-types·                                                                                          │          │
  │ LAUNCH           │ taskfile-vars │ string   │ ·node···--experimental-strip-types··src/launch.ts                                                                     │          │
  │ NODE_ARGS        │ taskfile-vars │ string   │ --experimental-strip-types                                                                                            │          │
  │ OUTPUT_DIR       │ taskfile-vars │ string   │ ./output                                                                                                              │          │
  │ OUT_DIR          │ taskfile-vars │ string   │ ./out                                                                                                                 │          │
  │ SCHEMAS_DIR      │ taskfile-vars │ string   │ ./schemas                                                                                                             │          │
  └──────────────────┴───────────────┴──────────┴───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┴──────────┘

── Task: debug
  Variables in scope:
  ┌────────────────┬───────────┬──────────┬───────────────┬──────────┐
  │ Name           │ Origin    │ Type     │ Value         │ Shadows? │
  ├────────────────┼───────────┼──────────┼───────────────┼──────────┤
  │ MATCH          │ call-vars │ []string │ []            │          │
  │                │           │          │ ptr: 0x1070a1600 │          │
  │ COLOR_AQUA     │ task-vars │ string   │ 33[38;5;87m  │          │
  │ COLOR_BLACK    │ task-vars │ string   │ 33[30m       │          │
  │ COLOR_BLUE     │ task-vars │ string   │ 33[34m       │          │
  │ COLOR_BOLD     │ task-vars │ string   │ 33[1m        │          │
  │ COLOR_BROWN    │ task-vars │ string   │ 33[38;5;130m │          │
  │ COLOR_CYAN     │ task-vars │ string   │ 33[36m       │          │
  │ COLOR_DIM      │ task-vars │ string   │ 33[2m        │          │
  │ COLOR_GRAY     │ task-vars │ string   │ 33[90m       │          │
  │ COLOR_GREEN    │ task-vars │ string   │ 33[32m       │          │
  │ COLOR_INDIGO   │ task-vars │ string   │ 33[38;5;63m  │          │
  │ COLOR_LAVENDER │ task-vars │ string   │ 33[38;5;183m │          │
  │ COLOR_LIME     │ task-vars │ string   │ 33[38;5;154m │          │
  │ COLOR_MAGENTA  │ task-vars │ string   │ 33[35m       │          │
  │ COLOR_MAUVE    │ task-vars │ string   │ 33[38;5;135m │          │
  │ COLOR_ORANGE   │ task-vars │ string   │ 33[38;5;208m │          │
  │ COLOR_PINK     │ task-vars │ string   │ 33[38;5;213m │          │
  │ COLOR_PLUM     │ task-vars │ string   │ 33[38;5;93m  │          │
  │ COLOR_PURPLE   │ task-vars │ string   │ 33[38;5;170m │          │
  │ COLOR_RED      │ task-vars │ string   │ 33[31m       │          │
  │ COLOR_RESET    │ task-vars │ string   │ 33[0m        │          │
  │ COLOR_ROSE     │ task-vars │ string   │ 33[38;5;211m │          │
  │ COLOR_SALMON   │ task-vars │ string   │ 33[38;5;217m │          │
  │ COLOR_SKY      │ task-vars │ string   │ 33[38;5;117m │          │
  │ COLOR_TAN      │ task-vars │ string   │ 33[38;5;180m │          │
  │ COLOR_TEAL     │ task-vars │ string   │ 33[38;5;37m  │          │
  │ COLOR_VIOLET   │ task-vars │ string   │ 33[38;5;127m │          │
  │ COLOR_WHITE    │ task-vars │ string   │ 33[37m       │          │
  │ COLOR_YELLOW   │ task-vars │ string   │ 33[33m       │          │
  │ SPACE          │ task-vars │ int      │ 20            │          │
  └────────────────┴───────────┴──────────┴───────────────┴──────────┘
  Template Evaluation — cmds[0]:
  ┌─ Input:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "{{spew (.ENGINE | trim)}}"
  │ echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'
  │
  └─
  ┌─ Evaluation Steps:
  │ Step 1: Apply a Function — spew
  │   Input:  spew .ENGINE | trim
  │   Output: (string) (len=4) "node"

  │ Step 2: Resolve a Variable — .SPACE
  │   Input:  20
  │   Expr:   echo ":: Global Taskfile variables ::"
echo ""
echo "{{spew (.ENGINE | trim)}}"
echo '{{printf "%s: %*s %s" "ENGINE" "20" (.ENGINE | trim)}}'

  │ Step 3: Apply a Function — printf
  │   Input:  printf "%s: %*s %s" "ENGINE" 20 .ENGINE | trim
  │   Output: ENGINE:                 node %!s(MISSING)
  │   Expr:   echo ":: Global Taskfile variables ::"
echo ""
echo "ENGINE:                 node %!s(MISSING)'

  └─
  ┌─ Output:
  │ echo·"::·Global·Taskfile·variables·::"
  │ echo·""
  │ echo·"(string)·(len=4)·"node"
  │ "
  │ echo·'ENGINE:·················node·%!s(MISSING)'
  │
  └─
  ┌─ Vars used:
  │ ENGINE, SPACE
  └─
  ℹ Note: Hint: This looks like a printf format error. printf signature: printf(format string, args ...any) string
    Example: {{printf "%s: %s" .KEY .VALUE}}
  Commands — cmds[0]:
  ┌─ Raw:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "{{spew (.ENGINE | trim)}}"
  │ echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'
  │
  └─
  ┌─ Resolved:
  │ echo·"::·Global·Taskfile·variables·::"
  │ echo·""
  │ echo·"(string)·(len=4)·"node"
  │ "
  │ echo·'ENGINE:·················node·%!s(MISSING)'
  │
  └─

╚══ End of Transparent Mode Report ══╝
```

**Example how the output should be:**

```
  ┌─ Evaluation Steps:
  │ Step 1: Apply a Function — spew
  │   I     spew·.ENGINE·|·trim
  │   O     (string)·(len=4)·"node"
  |         ·
  │ Step 2: Resolve a Variable — .SPACE
  │   I     20
  │   E     echo·"::·Global·Taskfile·variables·::"
  |         echo·""
  |         echo·"{{spew·(.ENGINE·|·trim)}}"
  |         echo·'{{printf·"%s:·%*s·%s"·"ENGINE"·"20"·(.ENGINE·|·trim)}}'
  │ Step 3: Apply a Function — printf
  │   I     printf·"%s:·%*s·%s"·"ENGINE"·20·.ENGINE·|·trim
  │   O     ENGINE:·················node·%!s(MISSING)
  │   E     echo·"::·Global·Taskfile·variables·::"
  |         echo·""
  |         echo·"ENGINE:·················node·%!s(MISSING)'
  └─
```

---

```shell
❯ task-dev --show-whitespaces -v debug

╔════════════════════════════════════════════════════════╗
║  TRANSPARENT MODE — Variable & Template Diagnostics  ║
╚════════════════════════════════════════════════════════╝
Legend: · = space, → = tab

── Global Variables
  Variables in scope:
  ┌──────────────────┬───────────────┬──────────┬───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┬──────────┐
  │ Name             │ Origin        │ Type     │ Value                                                                                                                 │ Shadows? │
  ├──────────────────┼───────────────┼──────────┼───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┼──────────┤
  │ ALIAS            │ special       │ string   │ debug                                                                                                                 │          │
  │ ROOT_DIR         │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder              │          │
  │ ROOT_TASKFILE    │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder/Taskfile.yml │          │
  │ TASK             │ special       │ string   │ debug                                                                                                                 │          │
  │ TASKFILE         │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder/Taskfile.yml │          │
  │ TASKFILE_DIR     │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder              │          │
  │ TASK_DIR         │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder              │          │
  │ TASK_EXE         │ special       │ string   │ task-dev                                                                                                              │          │
  │ TASK_VERSION     │ special       │ string   │ 3.48.0                                                                                                                │          │
  │ USER_WORKING_DIR │ special       │ string   │ /Users/tobiashochgurtel/work-dev/temp-projects/compare-vscode-extension_inline-fold/vscode-demo-recorder              │          │
  │ BUN_ARGS         │ taskfile-vars │ string   │ --bun                                                                                                                 │          │
  │ CLI              │ taskfile-vars │ string   │ ·node···--experimental-strip-types··src/cli.ts                                                                        │          │
  │ CLI_ARGS         │ taskfile-vars │ string   │                                                                                                                       │          │
  │ CLI_ARGS_LIST    │ taskfile-vars │ []string │ []                                                                                                                    │          │
  │                  │               │          │ ptr: 0x1038cd600                                                                                                      │          │
  │ CLI_ASSUME_YES   │ taskfile-vars │ bool     │ false                                                                                                                 │          │
  │ CLI_FORCE        │ taskfile-vars │ bool     │ false                                                                                                                 │          │
  │ CLI_OFFLINE      │ taskfile-vars │ bool     │ false                                                                                                                 │          │
  │ CLI_SILENT       │ taskfile-vars │ bool     │ false                                                                                                                 │          │
  │ CLI_VERBOSE      │ taskfile-vars │ bool     │ true                                                                                                                  │          │
  │ ENGINE           │ taskfile-vars │ string   │ ·node·                                                                                                                │          │
  │ ENGINE_ARGS      │ taskfile-vars │ string   │ ·--experimental-strip-types·                                                                                          │          │
  │ LAUNCH           │ taskfile-vars │ string   │ ·node···--experimental-strip-types··src/launch.ts                                                                     │          │
  │ NODE_ARGS        │ taskfile-vars │ string   │ --experimental-strip-types                                                                                            │          │
  │ OUTPUT_DIR       │ taskfile-vars │ string   │ ./output                                                                                                              │          │
  │ OUT_DIR          │ taskfile-vars │ string   │ ./out                                                                                                                 │          │
  │ SCHEMAS_DIR      │ taskfile-vars │ string   │ ./schemas                                                                                                             │          │
  └──────────────────┴───────────────┴──────────┴───────────────────────────────────────────────────────────────────────────────────────────────────────────────────────┴──────────┘

── Task: debug
  Variables in scope:
  ┌────────────────┬───────────┬──────────┬──────────────────┬──────────┐
  │ Name           │ Origin    │ Type     │ Value            │ Shadows? │
  ├────────────────┼───────────┼──────────┼──────────────────┼──────────┤
  │ MATCH          │ call-vars │ []string │ []               │          │
  │                │           │          │ ptr: 0x1038cd600 │          │
  │ COLOR_AQUA     │ task-vars │ string   │ 33[38;5;87m     │          │
  │ COLOR_BLACK    │ task-vars │ string   │ 33[30m          │          │
  │ COLOR_BLUE     │ task-vars │ string   │ 33[34m          │          │
  │ COLOR_BOLD     │ task-vars │ string   │ 33[1m           │          │
  │ COLOR_BROWN    │ task-vars │ string   │ 33[38;5;130m    │          │
  │ COLOR_CYAN     │ task-vars │ string   │ 33[36m          │          │
  │ COLOR_DIM      │ task-vars │ string   │ 33[2m           │          │
  │ COLOR_GRAY     │ task-vars │ string   │ 33[90m          │          │
  │ COLOR_GREEN    │ task-vars │ string   │ 33[32m          │          │
  │ COLOR_INDIGO   │ task-vars │ string   │ 33[38;5;63m     │          │
  │ COLOR_LAVENDER │ task-vars │ string   │ 33[38;5;183m    │          │
  │ COLOR_LIME     │ task-vars │ string   │ 33[38;5;154m    │          │
  │ COLOR_MAGENTA  │ task-vars │ string   │ 33[35m          │          │
  │ COLOR_MAUVE    │ task-vars │ string   │ 33[38;5;135m    │          │
  │ COLOR_ORANGE   │ task-vars │ string   │ 33[38;5;208m    │          │
  │ COLOR_PINK     │ task-vars │ string   │ 33[38;5;213m    │          │
  │ COLOR_PLUM     │ task-vars │ string   │ 33[38;5;93m     │          │
  │ COLOR_PURPLE   │ task-vars │ string   │ 33[38;5;170m    │          │
  │ COLOR_RED      │ task-vars │ string   │ 33[31m          │          │
  │ COLOR_RESET    │ task-vars │ string   │ 33[0m           │          │
  │ COLOR_ROSE     │ task-vars │ string   │ 33[38;5;211m    │          │
  │ COLOR_SALMON   │ task-vars │ string   │ 33[38;5;217m    │          │
  │ COLOR_SKY      │ task-vars │ string   │ 33[38;5;117m    │          │
  │ COLOR_TAN      │ task-vars │ string   │ 33[38;5;180m    │          │
  │ COLOR_TEAL     │ task-vars │ string   │ 33[38;5;37m     │          │
  │ COLOR_VIOLET   │ task-vars │ string   │ 33[38;5;127m    │          │
  │ COLOR_WHITE    │ task-vars │ string   │ 33[37m          │          │
  │ COLOR_YELLOW   │ task-vars │ string   │ 33[33m          │          │
  │ SPACE          │ task-vars │ int      │ 20               │          │
  └────────────────┴───────────┴──────────┴──────────────────┴──────────┘
  Template Evaluation — cmds[0]:
  ┌─ Input:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "{{spew (.ENGINE | trim)}}"
  │ echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'
  └─
  ┌─ Evaluation Steps:
  │ Step 1: Apply a Function — spew
  │   I     spew·.ENGINE·|·trim
  │   O     (string)·(len=4)·"node"
  │ Step 2: Resolve a Variable — .SPACE
  │   I     20
  │   E     echo·"::·Global·Taskfile·variables·::"
  │         echo·""
  │         echo·"{{spew·(.ENGINE·|·trim)}}"
  │         echo·'{{printf·"%s:·%*s·%s"·"ENGINE"·"20"·(.ENGINE·|·trim)}}'
  │ Step 3: Apply a Function — printf
  │   I     printf·"%s:·%*s·%s"·"ENGINE"·20·.ENGINE·|·trim
  │   O     ENGINE:·················node·%!s(MISSING)
  │   E     echo·"::·Global·Taskfile·variables·::"
  │         echo·""
  │         echo·"ENGINE:·················node·%!s(MISSING)'
  └─
  ┌─ Output:
  │ echo·"::·Global·Taskfile·variables·::"
  │ echo·""
  │ echo·"(string)·(len=4)·"node"
  │ "
  │ echo·'ENGINE:·················node·%!s(MISSING)'
  └─
  ┌─ Vars used:
  │ ENGINE, SPACE
  └─
  ℹ Note: Hint: This looks like a printf format error. printf signature: printf(format string, args ...any) string
    Example: {{printf "%s: %s" .KEY .VALUE}}
  Commands — cmds[0]:
  ┌─ Raw:
  │ echo ":: Global Taskfile variables ::"
  │ echo ""
  │ echo "{{spew (.ENGINE | trim)}}"
  │ echo '{{printf "%s: %*s %s" "ENGINE" .SPACE (.ENGINE | trim)}}'
  └─
  ┌─ Resolved:
  │ echo·"::·Global·Taskfile·variables·::"
  │ echo·""
  │ echo·"(string)·(len=4)·"node"
  │ "
  │ echo·'ENGINE:·················node·%!s(MISSING)'
  └─

╚══ End of Transparent Mode Report ══╝
```
