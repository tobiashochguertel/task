# 09 — Edge Cases & Go Template Pitfalls

## Common User Confusions (that Transparent Mode surfaces)

### 1. Pipe vs Parenthesization

```yaml
# User writes:
cmds:
  - echo {{printf "%s : %s" "NAME" .NAME | trim}}

# User expects: .NAME is trimmed, then passed to printf
# Actual: printf runs first, trim is applied to printf's full output
```

**Transparent Mode output:**
```
Step 1: printf "%s : %s" "NAME" "  hello  " → "NAME :   hello  "
Step 2: trim "NAME :   hello  "             → "NAME :   hello"
⚠ Tip: To trim .NAME before printf, use: {{printf "%s : %s" "NAME" (.NAME | trim)}}
```

### 2. `<no value>` Silently Eaten — ✅ IMPLEMENTED

Current behavior: `<no value>` is replaced with `""` (line 95 in `templater.go`). This silently hides undefined variables.

**Transparent Mode output:**
```
⚠  warning: template produced <no value> for one or more variables (replaced with empty string)
```

### 3. Dynamic Variable Not Resolved in Fast Mode

When using `--list` or `--list-all`, `FastGetVariables()` skips `sh:` evaluation. Variables with `sh:` show as empty.

**Transparent Mode output:**
```
DYNAMIC_VAR = ""  [task:vars]  type:string  ⚠ DYNAMIC (sh: "echo hello") — not evaluated in list mode
```

### 4. Variable Type Mismatch

```yaml
vars:
  COUNT: 42       # int
  NAME: "hello"   # string
cmds:
  - echo {{add .COUNT .NAME}}  # runtime error
```

**Transparent Mode output:**
```
⚠ Type mismatch in template expression:
  add(.COUNT=42 [int], .NAME="hello" [string])
  Expected: numeric arguments
```

### 5. Include Variable Scoping

```yaml
# Taskfile.yml
includes:
  app:
    taskfile: ./app/Taskfile.yml
    vars:
      ENV: production

# app/Taskfile.yml
vars:
  ENV: development
```

**Transparent Mode output:**
```
Task: app:deploy
  ENV = "production"  [include:vars]  ⚠ SHADOWS app/Taskfile.yml ENV="development" [included:taskfile:vars]
```

### 6. Ref Variables

```yaml
vars:
  LIST:
    - a
    - b
  ITEMS:
    ref: LIST
```

**Transparent Mode output:**
```
LIST  = ["a", "b"]              [taskfile:vars]  type:[]any
ITEMS = ["a", "b"]              [taskfile:vars]  type:[]any  ref:LIST  ← same instance
```

### 7. FOR Loop Iterator Variables — ✅ IMPLEMENTED

```yaml
cmds:
  - for:
      var: ITEMS
    cmd: echo {{.ITEM}}
```

**Transparent Mode output:**
```
Commands:
  [0] (ITEM=a) raw:      echo {{.ITEM}}
       resolved: echo a
  [1] (ITEM=b) raw:      echo {{.ITEM}}
       resolved: echo b
  [2] (ITEM=c) raw:      echo {{.ITEM}}
       resolved: echo c
```

## Template Function Chaining Rules

For user education, Transparent Mode can display a tip when it detects common patterns:

| Pattern | Tip |
|---------|-----|
| `{{.X \| trim}}` | ✅ Correct — trims .X |
| `{{printf "%s" .X \| trim}}` | ⚠ Trims printf output, not .X |
| `{{.X \| printf "%s"}}` | ✅ .X is piped as last arg to printf |
| `{{.X \| upper \| trim}}` | ✅ .X → upper → trim (left to right) |

## Instance Identity

When two variables reference the same underlying data:

```yaml
vars:
  ITEMS: [a, b, c]
  COPY:
    ref: ITEMS
```

Transparent Mode shows:
```
ITEMS = ["a","b","c"]  [taskfile:vars]  ptr:0xc0001a2000
COPY  = ["a","b","c"]  [taskfile:vars]  ptr:0xc0001a2000  ← same instance as ITEMS
```

Use `reflect.ValueOf(v).Pointer()` for slice/map types to detect identity.
