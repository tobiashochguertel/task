# Properties Reference

This section provides detailed documentation for all properties available in the Taskfile schema.

## Root Properties

- [version](./version.md) - Taskfile schema version.
- [includes](./includes.md) - Include other Taskfiles.
- [vars](./vars.md) - Global variables.
- [env](./env.md) - Global environment variables.
- [dotenv](./dotenv.md) - Load environment variables from files.
- [output](./output.md) - Output formatting.
- [method](./method.md) - Default up-to-date check method.
- [run](./run.md) - Default execution behavior.
- [interval](./interval.md) - Watch interval.
- [silent](./silent.md) - Suppress output by default.
- [set](./set.md) - POSIX shell options.
- [shopt](./shopt.md) - Bash shell options.

## Task Properties

- [cmds](./cmds.md) - Commands to execute.
- [deps](./deps.md) - Task dependencies.
- [desc](./desc.md) - Short description.
- [summary](./summary.md) - Detailed description.
- [sources](./sources.md) - Source files to monitor.
- [generates](./generates.md) - Generated files.
- [status](./status.md) - Programmatic up-to-date check.
- [preconditions](./preconditions.md) - Pre-execution checks.
- [requires](./requires.md) - Required variables.
- [dir](./dir.md) - Working directory.
- [prompt](./prompt.md) - User confirmation.
- [aliases](./aliases.md) - Alternative names.
- [watch](./watch.md) - Watch mode.
- [platforms](./platforms.md) - Supported platforms.
