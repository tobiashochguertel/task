# Templating

Task uses Go's `text/template` package for templating.

## Functions

Task provides a set of built-in template functions.

*   `OS`: Returns the operating system.
*   `ARCH`: Returns the architecture.
*   `catLines`: Returns the content of a file as a list of lines.
*   `splitLines`: Splits a string into a list of lines.
*   `fromSlash`: Replaces slashes with the OS path separator.
*   `toSlash`: Replaces the OS path separator with slashes.
*   `exeExt`: Returns the executable extension for the OS.

## Logic

You can use standard Go template logic.

```yaml
cmds:
  - '{{if eq OS "windows"}}echo "Windows"{{else}}echo "Unix"{{end}}'
```
