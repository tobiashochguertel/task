# Your First Taskfile

This tutorial will guide you through creating a complete Taskfile for a hypothetical Go project. By the end, you'll have a robust setup with build, test, and cleanup tasks.

## 1. Project Setup

Create a new directory for your project and initialize a `Taskfile.yml`.

```bash
mkdir my-project
cd my-project
task --init
```

This creates a basic `Taskfile.yml`. Open it in your editor.

## 2. Defining Variables

Let's define some global variables for our project, like the binary name and the source files.

```yaml
version: '3'

vars:
  BINARY_NAME: myapp
  SRC_DIR: ./src
```

## 3. The Build Task

Now, let's create a task to build our application. We'll use the variables we defined.

```yaml
tasks:
  build:
    desc: Build the application
    cmds:
      - go build -o {{.BINARY_NAME}} {{.SRC_DIR}}
    sources:
      - '{{.SRC_DIR}}/**/*.go'
    generates:
      - '{{.BINARY_NAME}}'
```

We added `sources` and `generates` so Task knows when to skip the build if nothing changed.

## 4. The Test Task

Next, a task to run tests.

```yaml
tasks:
  # ... build task ...

  test:
    desc: Run tests
    cmds:
      - go test -v {{.SRC_DIR}}/...
```

## 5. The Clean Task

It's good practice to have a task to clean up generated files.

```yaml
tasks:
  # ... other tasks ...

  clean:
    desc: Clean up generated files
    cmds:
      - rm -f {{.BINARY_NAME}}
```

## 6. The Default Task

Finally, let's define a `default` task that runs when you type just `task`. Usually, this runs tests and then builds.

```yaml
tasks:
  default:
    desc: Run tests and build
    deps: [test, build]
```

## Complete Taskfile

Here is the complete `Taskfile.yml`:

```yaml
version: '3'

vars:
  BINARY_NAME: myapp
  SRC_DIR: ./src

tasks:
  default:
    desc: Run tests and build
    deps: [test, build]

  build:
    desc: Build the application
    cmds:
      - go build -o {{.BINARY_NAME}} {{.SRC_DIR}}
    sources:
      - '{{.SRC_DIR}}/**/*.go'
    generates:
      - '{{.BINARY_NAME}}'

  test:
    desc: Run tests
    cmds:
      - go test -v {{.SRC_DIR}}/...

  clean:
    desc: Clean up generated files
    cmds:
      - rm -f {{.BINARY_NAME}}
```

## Running Your Tasks

Now you can run your tasks:

```bash
task test     # Run tests
task build    # Build the app
task          # Run tests and build
task clean    # Cleanup
```
