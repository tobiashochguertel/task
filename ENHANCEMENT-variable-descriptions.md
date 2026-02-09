# Enhancement: Variable Descriptions in Taskfile

## Overview

This document describes how to extend the Task CLI and JSON Schema to support descriptions for variables defined in Taskfiles. This enhancement will allow users to document their variables with descriptive text, improving code readability and maintainability.

## Current Implementation

### Variable Structure (`taskfile/ast/var.go`)

Currently, the `Var` struct supports the following fields:

```go
type Var struct {
    Value any
    Live  any
    Sh    *string
    Ref   string
    Dir   string
}
```

Variables can be defined in several ways:
- Static values (string, number, boolean, array, null)
- Dynamic shell commands (`sh`)
- Variable references (`ref`)
- Map values (`map`)

### JSON Schema (`website/src/public/schema.json`)

The current JSON Schema definition for variables (lines 273-312):

```json
"vars": {
  "type": "object",
  "patternProperties": {
    "^.*$": {
      "anyOf": [
        {
          "type": ["boolean", "integer", "null", "number", "string", "array"]
        },
        {
          "$ref": "#/definitions/var_subkey"
        }
      ]
    }
  }
},
"var_subkey": {
  "type": "object",
  "properties": {
    "sh": {
      "type": "string",
      "description": "The value will be treated as a command and the output assigned to the variable"
    },
    "ref": {
      "type": "string",
      "description": "The value will be used to lookup the value of another variable which will then be assigned to this variable"
    },
    "map": {
      "type": "object",
      "description": "The value will be treated as a literal map type and stored in the variable"
    }
  },
  "additionalProperties": false
}
```

## Proposed Enhancement

### 1. Extend the `Var` Struct

Add a `Desc` field to the `Var` struct in `taskfile/ast/var.go`:

```go
type Var struct {
    Value any
    Live  any
    Sh    *string
    Ref   string
    Dir   string
    Desc  string  // NEW: Description of the variable
}
```

### 2. Update UnmarshalYAML Method

Modify the `UnmarshalYAML` method in `taskfile/ast/var.go` to handle the `desc` field:

```go
func (v *Var) UnmarshalYAML(node *yaml.Node) error {
    switch node.Kind {
    case yaml.MappingNode:
        key := "<none>"
        if len(node.Content) > 0 {
            key = node.Content[0].Value
        }
        switch key {
        case "sh", "ref", "map", "desc":  // Add "desc" to the valid keys
            var m struct {
                Sh   *string
                Ref  string
                Map  any
                Desc string  // NEW: Add description field
            }
            if err := node.Decode(&m); err != nil {
                return errors.NewTaskfileDecodeError(err, node)
            }
            v.Sh = m.Sh
            v.Ref = m.Ref
            v.Value = m.Map
            v.Desc = m.Desc  // NEW: Assign description
            return nil
        default:
            return errors.NewTaskfileDecodeError(nil, node).WithMessage(`%q is not a valid variable type. Try "sh", "ref", "map", "desc" or using a scalar value`, key)
        }
    default:
        var value any
        if err := node.Decode(&value); err != nil {
            return errors.NewTaskfileDecodeError(err, node)
        }
        v.Value = value
        return nil
    }
}
```

### 3. Update JSON Schema

Modify the `var_subkey` definition in `website/src/public/schema.json`:

```json
"var_subkey": {
  "type": "object",
  "properties": {
    "sh": {
      "type": "string",
      "description": "The value will be treated as a command and the output assigned to the variable"
    },
    "ref": {
      "type": "string",
      "description": "The value will be used to lookup the value of another variable which will then be assigned to this variable"
    },
    "map": {
      "type": "object",
      "description": "The value will be treated as a literal map type and stored in the variable"
    },
    "desc": {
      "type": "string",
      "description": "A description of what this variable is used for"
    }
  },
  "additionalProperties": false
}
```

### 4. Add CLI Display Support

To make variable descriptions useful, add support to display them when listing variables or in help output. This could be added to:

- `task --list-vars` command (if it exists or needs to be created)
- `task --summary` command to show variable descriptions
- Interactive variable selection/autocomplete

Example implementation in the appropriate CLI command handler:

```go
// Display variable with description
if v.Desc != "" {
    fmt.Printf("%s: %s\n", varName, v.Desc)
} else {
    fmt.Printf("%s\n", varName)
}
```

## Usage Examples

### Basic Variable Description

```yaml
version: '3'

vars:
  APP_NAME:
    desc: The name of the application used in build artifacts
    sh: echo "myapp"
  
  VERSION:
    desc: Application version, automatically derived from git tags
    sh: git describe --tags --always
  
  DEBUG:
    desc: Enable debug logging and verbose output
    sh: echo "false"
```

### With Different Variable Types

```yaml
version: '3'

vars:
  # Static variable with description
  PORT:
    desc: The port number the application listens on
    map: 8080  # Using map syntax to allow desc
  
  # Dynamic variable with description
  BUILD_TIME:
    desc: Timestamp when the build was created
    sh: date -u +"%Y-%m-%dT%H:%M:%SZ"
  
  # Reference variable with description
  FULL_VERSION:
    desc: Complete version string including commit hash
    ref: .VERSION
  
  # Map variable with description
  DATABASE:
    desc: Database configuration settings
    map:
      host: localhost
      port: 5432
      name: mydb
```

### In Task Context

```yaml
version: '3'

vars:
  ENVIRONMENT:
    desc: Deployment environment (development, staging, production)
    sh: echo "development"

tasks:
  deploy:
    desc: Deploy the application to the specified environment
    vars:
      TARGET_HOST:
        desc: The hostname or IP address of the deployment target
        sh: echo "localhost"
      DEPLOY_USER:
        desc: SSH user for deployment
        sh: echo "deploy"
    cmds:
      - echo "Deploying to {{.ENVIRONMENT}} at {{.TARGET_HOST}}"
      - ssh {{.DEPLOY_USER}}@{{.TARGET_HOST}} "deploy-script.sh"
```

## Implementation Checklist

### Phase 1: Core Variable Description Support

1. **Core Changes:**
   - [ ] Add `Desc` field to `Var` struct in `taskfile/ast/var.go`
   - [ ] Update `UnmarshalYAML` method in `taskfile/ast/var.go` to handle `desc` field
   - [ ] Update error message in `UnmarshalYAML` to include `desc` as valid key

2. **Schema Updates:**
   - [ ] Add `desc` property to `var_subkey` in `website/src/public/schema.json`
   - [ ] Validate schema against test Taskfiles with variable descriptions

3. **Testing:**
   - [ ] Add unit tests for variable description parsing in `taskfile/ast/var_test.go`
   - [ ] Add tests for variables with and without descriptions
   - [ ] Test YAML unmarshaling with `desc` field
   - [ ] Test schema validation with variable descriptions

### Phase 2: Variable Description Inheritance

4. **Inheritance Implementation:**
   - [ ] Modify `Merge` method in `taskfile/ast/vars.go` to implement description inheritance
   - [ ] Add logic to preserve parent description when child doesn't specify one
   - [ ] Add logic to allow child description to override parent description
   - [ ] Update `getVariables` in `compiler.go` if needed for proper inheritance

5. **Inheritance Testing:**
   - [ ] Add tests for description inheritance from global to task level
   - [ ] Add tests for description override at task level
   - [ ] Add tests for description inheritance through call variables
   - [ ] Add tests for multi-level inheritance (global → include → task)
   - [ ] Add integration tests with complex Taskfile hierarchies

### Phase 3: CLI and Tab Completion

6. **CLI Enhancements:**
   - [ ] Add `--list-vars` flag in `internal/flags/flags.go`
   - [ ] Implement `ListVariables` method in `help.go`
   - [ ] Add variable description display (both table and JSON formats)
   - [ ] Update help text to mention `--list-vars` flag
   - [ ] Add `--json` support for `--list-vars` output

7. **Editor Integration (JSON Output):**
   - [ ] Add `Variable` struct to `internal/editors/output.go`
   - [ ] Add `Vars` field to `Namespace` struct
   - [ ] Update `ToEditorOutput` in `help.go` to include variable information
   - [ ] Populate variable descriptions in JSON output
   - [ ] Include static variable values, exclude dynamic ones

8. **Shell Completion:**
   - [ ] Update bash completion script (`completion/bash/task.bash`) for variable completion
   - [ ] Update fish completion script (`completion/fish/task.fish`) for variable completion
   - [ ] Update zsh completion script (`completion/zsh/_task`) for variable completion
   - [ ] Update PowerShell completion script (`completion/ps/task.ps1`) for variable completion
   - [ ] Add variable name completion after `VAR_NAME=` pattern

9. **CLI Testing:**
   - [ ] Test `--list-vars` output in table format
   - [ ] Test `--list-vars --json` output
   - [ ] Test JSON output includes variable descriptions
   - [ ] Test completion scripts with variable descriptions
   - [ ] Test editor integration with enhanced JSON output

### Phase 4: Documentation

10. **User Documentation:**
    - [ ] Update `website/src/docs/guide.md` Variables section with description examples
    - [ ] Update `website/src/docs/reference/schema.md` Variable section
    - [ ] Add examples showing variable descriptions at different levels
    - [ ] Add examples showing description inheritance
    - [ ] Document `--list-vars` flag in CLI reference
    - [ ] Add section on variable description best practices

11. **API Documentation:**
    - [ ] Update JSON output schema documentation
    - [ ] Document the `Variable` type in editor output
    - [ ] Add examples of JSON output with variables
    - [ ] Update any API documentation for completions

### Phase 5: Final Integration

12. **End-to-End Testing:**
    - [ ] Test complete workflow: define → inherit → override → display
    - [ ] Test JSON output in real editor integrations (VSCode, etc.)
    - [ ] Test shell completions in real terminals
    - [ ] Performance test with large Taskfiles containing many variables
    - [ ] Backward compatibility test with old Taskfiles

13. **Code Quality:**
    - [ ] Ensure all tests pass
    - [ ] Run linters and fix any issues
    - [ ] Add code comments for new functionality
    - [ ] Update CHANGELOG.md with new feature
    - [ ] Consider adding deprecation notices if needed

## Benefits

1. **Self-Documentation**: Variables become self-documenting, reducing the need for external documentation
2. **IDE Support**: Better autocomplete and hints in IDEs that support JSON Schema
3. **Team Collaboration**: Makes it easier for team members to understand variable purposes
4. **Maintenance**: Easier to maintain and update Taskfiles over time
5. **Consistency**: Follows the same pattern as task descriptions (`desc` field)

## Compatibility

This is a **backwards-compatible** enhancement:
- Existing Taskfiles without variable descriptions will continue to work
- The `desc` field is optional
- No breaking changes to existing functionality

## Alternative Approaches Considered

1. **Using YAML Comments**: While comments can document variables, they:
   - Are not accessible programmatically
   - Don't provide IDE autocomplete support
   - Cannot be validated via JSON Schema

2. **Separate Documentation File**: External documentation:
   - Gets out of sync easily
   - Requires switching contexts
   - Not available in CLI help

3. **Using a `help` Field Instead**: The `desc` field is preferred because:
   - Consistent with task descriptions
   - Familiar to existing Task users
   - Shorter and more common convention

## Variable Description Inheritance

### Overview

Variables can be defined at multiple levels in a Taskfile:
1. **Global level** - `vars:` at the root of the Taskfile
2. **Task level** - `vars:` within a specific task definition
3. **Call level** - Variables passed when calling a task from another task

When a variable with the same name is defined at multiple levels, the current implementation overwrites the entire variable object during merging (see `taskfile/ast/vars.go`, line 122-134).

### Proposed Inheritance Behavior

When a variable is redefined at a deeper scope with the same name:

1. **Description Inheritance**: If the new definition doesn't include a `desc` field, it should inherit the description from the parent scope
2. **Description Override**: If the new definition includes a `desc` field, it should override the inherited description
3. **Optional Field**: The `desc` field remains optional at all levels - variables without descriptions work as before

### Implementation Details

The inheritance logic should be implemented in the `Merge` method of `taskfile/ast/vars.go`:

```go
// Merge loops over other and merges it values with the variables in vars. If
// the include parameter is not nil and its it is an advanced import, the
// directory is set set to the value of the include parameter.
func (vars *Vars) Merge(other *Vars, include *Include) {
	if vars == nil || vars.om == nil || other == nil {
		return
	}
	defer other.mutex.RUnlock()
	other.mutex.RLock()
	for pair := other.om.Front(); pair != nil; pair = pair.Next() {
		newVar := pair.Value
		
		// Handle description inheritance
		if existingVar, exists := vars.om.Get(pair.Key); exists {
			// If the new variable doesn't have a description, inherit it
			if newVar.Desc == "" && existingVar.Desc != "" {
				newVar.Desc = existingVar.Desc
			}
		}
		
		if include != nil && include.AdvancedImport {
			newVar.Dir = include.Dir
		}
		vars.om.Set(pair.Key, newVar)
	}
}
```

### Variable Merging Order

According to `compiler.go` (lines 106-142), variables are merged in this order:

1. Environment variables
2. Special variables (TASK, etc.)
3. Taskfile environment (`TaskfileEnv`)
4. Taskfile global variables (`TaskfileVars`)
5. Include variables (if task is from an included Taskfile)
6. Included Taskfile variables (if task is from an included Taskfile)
7. Call variables (variables passed when calling a task)
8. Task-level variables (`t.Vars`)

Description inheritance should apply at each merging step where a variable with the same name already exists.

### Example Use Cases

#### Case 1: Global Description Inherited by Task

```yaml
version: '3'

vars:
  DATABASE_HOST:
    desc: The hostname of the database server
    sh: echo "localhost"

tasks:
  deploy:
    vars:
      DATABASE_HOST:
        # No desc specified - inherits "The hostname of the database server"
        sh: echo "production-db.example.com"
    cmds:
      - echo "Deploying to {{.DATABASE_HOST}}"
```

#### Case 2: Task Overrides Global Description

```yaml
version: '3'

vars:
  PORT:
    desc: Default application port
    sh: echo "8080"

tasks:
  dev:
    vars:
      PORT:
        desc: Development server port (different from production)
        sh: echo "3000"
    cmds:
      - echo "Starting on port {{.PORT}}"
```

#### Case 3: Call Variables Inherit Task Description

```yaml
version: '3'

tasks:
  deploy-service:
    vars:
      SERVICE_NAME:
        desc: Name of the service to deploy
        sh: echo "api"
    cmds:
      - echo "Deploying {{.SERVICE_NAME}}"
  
  deploy-all:
    cmds:
      - task: deploy-service
        vars:
          SERVICE_NAME: frontend  # Inherits desc from deploy-service task
      - task: deploy-service
        vars:
          SERVICE_NAME: backend   # Inherits desc from deploy-service task
```

## Tab Completion Enhancement

### Current Implementation

The tab completion system works by:
1. Shell completion scripts call `task --silent --list-all` to get task names (see `completion/bash/task.bash`, line 45)
2. For JSON output, the CLI supports `task --list --json` or `task --list-all --json` flags
3. The JSON output is generated by `ToEditorOutput` in `help.go` (line 140)
4. The output structure is defined in `internal/editors/output.go`

### Current JSON Output Structure

```json
{
  "tasks": [
    {
      "name": "build",
      "task": "build",
      "desc": "Build the application",
      "summary": "Compiles the source code...",
      "aliases": ["compile"],
      "up_to_date": false,
      "location": {
        "line": 10,
        "column": 3,
        "taskfile": "/path/to/Taskfile.yml"
      }
    }
  ],
  "location": "/path/to/Taskfile.yml"
}
```

### Proposed Enhancement

Add variable information to the JSON output to support tab completion for variables.

#### 1. Extend the Output Structure

Add a new `Vars` field to the `Namespace` struct in `internal/editors/output.go`:

```go
type Namespace struct {
	Tasks      []Task                `json:"tasks"`
	Vars       []Variable            `json:"vars,omitempty"`     // NEW
	Namespaces map[string]*Namespace `json:"namespaces,omitempty"`
	Location   string                `json:"location,omitempty"`
}

// Variable describes a variable with its metadata
type Variable struct {
	Name  string `json:"name"`
	Desc  string `json:"desc,omitempty"`
	Value string `json:"value,omitempty"`  // For static vars, omit for dynamic
}
```

#### 2. Add CLI Flag for Variable Listing

Add a new flag `--list-vars` to list variables:

In `internal/flags/flags.go`:

```go
var (
	// ... existing flags ...
	ListVars bool
)

func init() {
	// ... existing flags ...
	pflag.BoolVar(&ListVars, "list-vars", false, "Lists global variables with descriptions.")
}
```

#### 3. Implement Variable Listing

In `help.go`, add a new method to list variables:

```go
// ListVariables prints a list of global variables with their descriptions
func (e *Executor) ListVariables(formatAsJson bool) error {
	vars := e.Taskfile.Vars
	
	if formatAsJson {
		output := make([]map[string]string, 0, vars.Len())
		for k, v := range vars.All() {
			varInfo := map[string]string{
				"name": k,
			}
			if v.Desc != "" {
				varInfo["desc"] = v.Desc
			}
			// Include value for static variables only
			if v.Value != nil && v.Sh == nil {
				varInfo["value"] = fmt.Sprintf("%v", v.Value)
			}
			output = append(output, varInfo)
		}
		
		encoder := json.NewEncoder(e.Stdout)
		encoder.SetIndent("", "  ")
		return encoder.Encode(output)
	}
	
	// Format as human-readable table
	e.Logger.Outf(logger.Default, "task: Available variables for this project:\n")
	w := tabwriter.NewWriter(e.Stdout, 0, 8, 6, ' ', 0)
	for k, v := range vars.All() {
		e.Logger.FOutf(w, logger.Yellow, "* ")
		e.Logger.FOutf(w, logger.Green, k)
		if v.Desc != "" {
			e.Logger.FOutf(w, logger.Default, ": \t%s", v.Desc)
		}
		_, _ = fmt.Fprint(w, "\n")
	}
	return w.Flush()
}
```

#### 4. Update ToEditorOutput to Include Variables

Modify `help.go` to populate the `Vars` field:

```go
func (e *Executor) ToEditorOutput(tasks []*ast.Task, noStatus bool, nested bool) (*editors.Namespace, error) {
	// ... existing task processing code ...
	
	// Add global variables to the output
	editorVars := make([]editors.Variable, 0, e.Taskfile.Vars.Len())
	for k, v := range e.Taskfile.Vars.All() {
		editorVar := editors.Variable{
			Name: k,
			Desc: v.Desc,
		}
		// Include value for static variables only (not dynamic sh/ref)
		if v.Value != nil && v.Sh == nil && v.Ref == "" {
			editorVar.Value = fmt.Sprintf("%v", v.Value)
		}
		editorVars = append(editorVars, editorVar)
	}
	
	rootNamespace := &editors.Namespace{
		Tasks:    make([]editors.Task, tasksLen),
		Vars:     editorVars,  // NEW
		Location: e.Taskfile.Location,
	}
	
	// ... rest of existing code ...
}
```

### Enhanced JSON Output Example

```json
{
  "tasks": [
    {
      "name": "build",
      "task": "build",
      "desc": "Build the application",
      "summary": "",
      "aliases": [],
      "location": {
        "line": 10,
        "column": 3,
        "taskfile": "/path/to/Taskfile.yml"
      }
    }
  ],
  "vars": [
    {
      "name": "APP_NAME",
      "desc": "The name of the application used in build artifacts",
      "value": "myapp"
    },
    {
      "name": "VERSION",
      "desc": "Application version, automatically derived from git tags"
    },
    {
      "name": "DEBUG",
      "desc": "Enable debug logging and verbose output",
      "value": "false"
    }
  ],
  "location": "/path/to/Taskfile.yml"
}
```

### Shell Completion Integration

Shell completion scripts can be enhanced to use the variable information:

#### Bash Completion Example

```bash
# Add a function to complete variable names
_task_vars()
{
  local vars=( $( "${words[@]}" --list-vars --json 2> /dev/null | \
    jq -r '.[].name' 2> /dev/null ) )
  COMPREPLY=( $( compgen -W "${vars[*]}" -- "$cur" ) )
}

# In the main completion function, detect variable context
# For example, after VAR_NAME= pattern
case "$cur" in
  *=)
    # User is trying to set a variable value
    _task_vars
    return 0
  ;;
esac
```

### Benefits of Tab Completion Enhancement

1. **Variable Discovery**: Users can discover available variables without reading the Taskfile
2. **IDE Integration**: Editors can provide autocomplete for variable names
3. **Documentation Access**: Variable descriptions appear in autocomplete suggestions
4. **Type Safety**: Knowing which variables exist reduces typos
5. **Consistency**: Variables and tasks both support JSON output and descriptions

## References

- Task struct with `Desc` field: `taskfile/ast/task.go`
- Current variable implementation: `taskfile/ast/var.go`
- Variable merging logic: `taskfile/ast/vars.go` (lines 122-134)
- Variable compilation: `compiler.go` (lines 47-145)
- JSON output structure: `internal/editors/output.go`
- List tasks implementation: `help.go` (lines 58-200)
- CLI flags: `internal/flags/flags.go`
- Bash completion script: `completion/bash/task.bash`
- JSON Schema: `website/src/public/schema.json`
- Variable documentation: `website/src/docs/guide.md` (line 1106+)
- Schema reference: `website/src/docs/reference/schema.md` (line 326+)
