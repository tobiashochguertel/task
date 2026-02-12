# 01 — Problem Statement & Vision

## Problem

<!-- ✅ CLOSED — All 4 problem areas are addressed by the transparent mode implementation. -->

When using Task's Go template features in `Taskfile.yml`, users face significant debugging challenges:

1. **Variable Origin Opacity** — It's unclear whether a variable value comes from:
   - Global `vars:` block
   - Task-level `vars:` block
   - Included Taskfile vars
   - Include directive vars
   - CLI arguments (`FOO=bar`)
   - Environment variables
   - Special vars (`TASK`, `ROOT_DIR`, etc.)
   - Dynamic shell vars (`sh: ...`)

2. **Template Expression Debugging** — Complex expressions like `{{printf "%s : %s" "NAME" .NAME | trim}}` are opaque:
   - What is the intermediate result of `printf` before `trim`?
   - Is `| trim` applied to the full output or just `.NAME`?
   - What is `.NAME`'s actual resolved value at render time?

3. **Variable Shadowing** — When a task-level var overrides a global var, there's no indication this happened.

4. **Instance Identity** — It's impossible to tell if two references to the same variable name resolve to the same value or different values from different scopes.

## Vision: Transparent Mode

<!-- ✅ CLOSED — --transparent/-T flag implemented; renders diagnostic report without executing tasks. -->

A **non-invasive overlay** that, when activated via `--transparent` (or `-T`), renders a human-readable diagnostic report **instead of executing tasks**. It shows:

- Every variable with its resolved value, origin (scope), and type
- Every template expression with step-by-step evaluation
- Variable shadowing warnings
- Pipe chain breakdowns for template functions
- A clear mapping from Taskfile line → resolved output

Transparent Mode is read-only — it never executes commands. It behaves like `--dry` but focuses on **template and variable introspection** rather than command listing.

## User Stories

<!-- ✅ CLOSED — All 5 stories implemented: origin display, pipe tracing, shadow warnings, sorted vars, step-by-step eval. -->

| #   | Story                                                                                           |
| --- | ----------------------------------------------------------------------------------------------- |
| 1   | As a user, I want to see what value `.NAME` resolves to and where it was defined.               |
| 2   | As a user, I want to understand why `{{printf "%s" .NAME \| trim}}` produces unexpected output. |
| 3   | As a user, I want to know when a task-level variable shadows a global one.                      |
| 4   | As a user, I want to see all variables available in a task's scope, sorted by origin.           |
| 5   | As a user, I want a step-by-step trace of template pipe evaluation.                             |

## Non-Goals (v1)

<!-- ✅ CLOSED — None of these non-goals were implemented; transparent mode remains read-only and non-invasive. -->

- Modifying any runtime behavior
- Adding interactive/REPL-style debugging
- Supporting breakpoints or stepping through execution
- Altering template evaluation order
