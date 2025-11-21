# Core Concepts

This guide introduces the fundamental concepts you need to understand to use Task effectively.

## What is a Task?

A **Task** is a named unit of work. It defines a set of commands to be executed. Tasks are defined in a `Taskfile.yml`.

```yaml
tasks:
  hello:
    cmds:
      - echo "Hello, World!"
```

In this example, `hello` is the task name, and it runs a single command `echo "Hello, World!"`.

## The Taskfile

The **Taskfile** (usually named `Taskfile.yml` or `Taskfile.yaml`) is the configuration file where you define your tasks and variables. It uses the [YAML](https://yaml.org/) format.

Every Taskfile starts with a version:

```yaml
version: '3'
```

## Dependencies

Tasks often depend on other tasks. For example, a `deploy` task might depend on a `build` task. Task allows you to define these relationships using **dependencies**.

```yaml
tasks:
  build:
    cmds:
      - go build .

  deploy:
    deps: [build]
    cmds:
      - ./deploy.sh
```

When you run `task deploy`, Task will ensure that `build` runs first. By default, dependencies run in parallel to speed up execution.

## Variables

**Variables** allow you to make your tasks dynamic and reusable. You can define variables globally or per task.

```yaml
vars:
  GREETING: Hello

tasks:
  greet:
    cmds:
      - echo "{{.GREETING}}, World!"
```

Task supports:
- **Static variables**: Simple values.
- **Dynamic variables**: Result of a shell command (e.g., getting the current git commit).

## Templating

Task uses **Go's template engine**. You can use variables and functions within your commands.

- `{{\.VAR_NAME}}`: Access a variable.
- `{{\.CLI_ARGS}}`: Access arguments passed from the command line.

## Running Tasks

You run tasks using the `task` CLI.

```bash
# Run the 'greet' task
task greet

# Run the default task (if defined)
task
```

## Next Steps

Now that you understand the core concepts, you can:

- **[Create your first Taskfile](./your-first-taskfile.md)**: A step-by-step tutorial.
- **[Explore the Cookbook](../cookbook/index.md)**: Find solutions for common problems.
- **[Browse the Reference](../reference/properties/index.md)**: detailed documentation for all properties.
