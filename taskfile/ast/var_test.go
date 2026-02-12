package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.yaml.in/yaml/v4"

	"github.com/go-task/task/v3/taskfile/ast"
)

func TestVarWithDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		content  string
		expected ast.Var
	}{
		{
			name: "variable with description and sh command",
			content: `
desc: "Application version number"
sh: "git describe --tags"
`,
			expected: ast.Var{
				Desc: "Application version number",
				Sh:   stringPtr("git describe --tags"),
			},
		},
		{
			name: "variable with description and ref",
			content: `
desc: "Reference to another variable"
ref: .BASE_VERSION
`,
			expected: ast.Var{
				Desc: "Reference to another variable",
				Ref:  ".BASE_VERSION",
			},
		},
		{
			name: "variable with description and map",
			content: `
desc: "Configuration map"
map:
  host: localhost
  port: 8080
`,
			expected: ast.Var{
				Desc:  "Configuration map",
				Value: map[string]any{"host": "localhost", "port": 8080},
			},
		},
		{
			name:    "sh command without description",
			content: `sh: "echo hello"`,
			expected: ast.Var{
				Sh: stringPtr("echo hello"),
			},
		},
		{
			name:    "static value without description",
			content: `"simple string"`,
			expected: ast.Var{
				Value: "simple string",
			},
		},
		{
			name:    "static number without description",
			content: `8080`,
			expected: ast.Var{
				Value: 8080,
			},
		},
		{
			name:    "static boolean without description",
			content: `true`,
			expected: ast.Var{
				Value: true,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			var v ast.Var
			err := yaml.Unmarshal([]byte(test.content), &v)
			require.NoError(t, err)
			assert.Equal(t, test.expected.Desc, v.Desc)
			if test.expected.Sh != nil {
				require.NotNil(t, v.Sh)
				assert.Equal(t, *test.expected.Sh, *v.Sh)
			}
			if test.expected.Ref != "" {
				assert.Equal(t, test.expected.Ref, v.Ref)
			}
			if test.expected.Value != nil {
				assert.Equal(t, test.expected.Value, v.Value)
			}
		})
	}
}

func TestVarsWithDescriptions(t *testing.T) {
	t.Parallel()

	content := `
APP_NAME:
  desc: "The application name"
  sh: "echo myapp"
VERSION:
  desc: "Application version"
  sh: "git describe --tags"
PORT: 8080
DEBUG:
  sh: "echo false"
`
	var vars ast.Vars
	err := yaml.Unmarshal([]byte(content), &vars)
	require.NoError(t, err)

	// Check APP_NAME
	appName, ok := vars.Get("APP_NAME")
	require.True(t, ok)
	assert.Equal(t, "The application name", appName.Desc)
	require.NotNil(t, appName.Sh)
	assert.Equal(t, "echo myapp", *appName.Sh)

	// Check VERSION
	version, ok := vars.Get("VERSION")
	require.True(t, ok)
	assert.Equal(t, "Application version", version.Desc)
	require.NotNil(t, version.Sh)
	assert.Equal(t, "git describe --tags", *version.Sh)

	// Check PORT (no description)
	port, ok := vars.Get("PORT")
	require.True(t, ok)
	assert.Equal(t, "", port.Desc)
	assert.Equal(t, 8080, port.Value)

	// Check DEBUG (no description)
	debug, ok := vars.Get("DEBUG")
	require.True(t, ok)
	assert.Equal(t, "", debug.Desc)
	require.NotNil(t, debug.Sh)
}

func TestVarDescriptionInheritance(t *testing.T) {
	t.Parallel()

	// Create parent vars with descriptions
	parentVars := ast.NewVars()
	parentVars.Set("VERSION", ast.Var{
		Desc: "Global version number",
		Sh:   stringPtr("echo 1.0.0"),
	})
	parentVars.Set("PORT", ast.Var{
		Desc: "Server port number",
		Sh:   stringPtr("echo 8080"),
	})

	// Create child vars
	childVars := ast.NewVars()
	// VERSION without description - should inherit
	childVars.Set("VERSION", ast.Var{
		Sh: stringPtr("echo 2.0.0"),
	})
	// PORT with description - should override
	childVars.Set("PORT", ast.Var{
		Desc: "Custom port for this task",
		Sh:   stringPtr("echo 3000"),
	})

	// Merge child into parent
	parentVars.Merge(childVars, nil)

	// Check VERSION inherited description
	version, ok := parentVars.Get("VERSION")
	require.True(t, ok)
	assert.Equal(t, "Global version number", version.Desc, "VERSION should inherit parent description")
	require.NotNil(t, version.Sh)
	assert.Equal(t, "echo 2.0.0", *version.Sh, "VERSION should have child value")

	// Check PORT overrode description
	port, ok := parentVars.Get("PORT")
	require.True(t, ok)
	assert.Equal(t, "Custom port for this task", port.Desc, "PORT should override parent description")
	require.NotNil(t, port.Sh)
	assert.Equal(t, "echo 3000", *port.Sh, "PORT should have child value")
}

func TestVarDescriptionInheritanceMultipleLevels(t *testing.T) {
	t.Parallel()

	// Level 1: Global vars
	globalVars := ast.NewVars()
	globalVars.Set("APP_NAME", ast.Var{
		Desc: "Application name",
		Sh:   stringPtr("echo myapp"),
	})

	// Level 2: Include vars (no description)
	includeVars := ast.NewVars()
	includeVars.Set("APP_NAME", ast.Var{
		Sh: stringPtr("echo override1"),
	})

	// Level 3: Task vars (new description)
	taskVars := ast.NewVars()
	taskVars.Set("APP_NAME", ast.Var{
		Desc: "Task-specific app name",
		Sh:   stringPtr("echo override2"),
	})

	// Merge: global -> include
	globalVars.Merge(includeVars, nil)
	appName, _ := globalVars.Get("APP_NAME")
	assert.Equal(t, "Application name", appName.Desc, "Should inherit from global")
	assert.Equal(t, "echo override1", *appName.Sh)

	// Merge: global+include -> task
	globalVars.Merge(taskVars, nil)
	appName, _ = globalVars.Get("APP_NAME")
	assert.Equal(t, "Task-specific app name", appName.Desc, "Should override with task description")
	assert.Equal(t, "echo override2", *appName.Sh)
}

func TestVarDescriptionWithEmptyString(t *testing.T) {
	t.Parallel()

	// Parent with description
	parentVars := ast.NewVars()
	parentVars.Set("VAR", ast.Var{
		Desc: "Original description",
		Sh:   stringPtr("echo parent"),
	})

	// Child with empty description - should still inherit
	childVars := ast.NewVars()
	childVars.Set("VAR", ast.Var{
		Desc: "",
		Sh:   stringPtr("echo child"),
	})

	parentVars.Merge(childVars, nil)
	v, _ := parentVars.Get("VAR")
	assert.Equal(t, "Original description", v.Desc, "Empty description should inherit from parent")
}

func TestVarOnlyDescriptionField(t *testing.T) {
	t.Parallel()

	// Test that a variable can have only a description (though not particularly useful)
	content := `
desc: "Just a description"
`
	var v ast.Var
	err := yaml.Unmarshal([]byte(content), &v)
	require.NoError(t, err)
	assert.Equal(t, "Just a description", v.Desc)
	assert.Nil(t, v.Sh)
	assert.Equal(t, "", v.Ref)
	assert.Nil(t, v.Value)
}

func TestVarStaticValueWithDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		yaml     string
		wantDesc string
		wantVal  any
	}{
		{
			name: "string value with description",
			yaml: `
desc: "Application name"
value: "super-app"
`,
			wantDesc: "Application name",
			wantVal:  "super-app",
		},
		{
			name: "number value with description",
			yaml: `
desc: "Server port"
value: 8080
`,
			wantDesc: "Server port",
			wantVal:  8080,
		},
		{
			name: "boolean value with description",
			yaml: `
desc: "Debug mode"
value: true
`,
			wantDesc: "Debug mode",
			wantVal:  true,
		},
		{
			name: "object value with description using value field",
			yaml: `
desc: "Configuration object"
value:
  host: localhost
  port: 5432
`,
			wantDesc: "Configuration object",
			wantVal: map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var v ast.Var
			err := yaml.Unmarshal([]byte(tt.yaml), &v)
			require.NoError(t, err)
			assert.Equal(t, tt.wantDesc, v.Desc)
			assert.Equal(t, tt.wantVal, v.Value)
			assert.Nil(t, v.Sh)
			assert.Equal(t, "", v.Ref)
		})
	}
}

func TestVarMapWithDescription(t *testing.T) {
	t.Parallel()

	// Test that 'map' field still works with description
	content := `
desc: "Database configuration"
map:
  host: localhost
  port: 5432
  ssl: true
`
	var v ast.Var
	err := yaml.Unmarshal([]byte(content), &v)
	require.NoError(t, err)
	assert.Equal(t, "Database configuration", v.Desc)
	assert.Equal(t, map[string]any{
		"host": "localhost",
		"port": 5432,
		"ssl":  true,
	}, v.Value)
}

func TestVarMutuallyExclusiveFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		yaml    string
		wantErr string
	}{
		{
			name: "sh and value",
			yaml: `
desc: "Test"
sh: echo test
value: "test"
`,
			wantErr: `cannot have more than one of`,
		},
		{
			name: "sh and ref",
			yaml: `
sh: echo test
ref: .OTHER
`,
			wantErr: `cannot have more than one of`,
		},
		{
			name: "value and ref",
			yaml: `
value: test
ref: .OTHER
`,
			wantErr: `cannot have more than one of`,
		},
		{
			name: "sh and map",
			yaml: `
sh: echo test
map:
  key: value
`,
			wantErr: `cannot have more than one of`,
		},
		{
			name: "value and map",
			yaml: `
value: test
map:
  key: value
`,
			wantErr: `cannot have more than one of`,
		},
		{
			name: "all three",
			yaml: `
sh: echo test
ref: .OTHER
value: test
`,
			wantErr: `cannot have more than one of`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var v ast.Var
			err := yaml.Unmarshal([]byte(tt.yaml), &v)
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantErr)
		})
	}
}

func TestVarDescOnlyIsAllowed(t *testing.T) {
	t.Parallel()

	// Description-only is allowed (though not particularly useful)
	content := `
desc: "Just a description, no value yet"
`
	var v ast.Var
	err := yaml.Unmarshal([]byte(content), &v)
	require.NoError(t, err)
	assert.Equal(t, "Just a description, no value yet", v.Desc)
	assert.Nil(t, v.Value)
	assert.Nil(t, v.Sh)
	assert.Equal(t, "", v.Ref)
}

func TestVarDescriptionBackwardCompatibility(t *testing.T) {
	t.Parallel()

	// Ensure old Taskfiles without descriptions still work
	content := `
OLD_VAR:
  sh: "echo old"
SIMPLE: "value"
NUMBER: 42
BOOL: true
ARRAY:
  - one
  - two
MAP_VAR:
  map:
    key: value
`
	var vars ast.Vars
	err := yaml.Unmarshal([]byte(content), &vars)
	require.NoError(t, err)

	// All variables should parse without errors and have no description
	oldVar, ok := vars.Get("OLD_VAR")
	require.True(t, ok)
	assert.Equal(t, "", oldVar.Desc)

	simple, ok := vars.Get("SIMPLE")
	require.True(t, ok)
	assert.Equal(t, "", simple.Desc)

	number, ok := vars.Get("NUMBER")
	require.True(t, ok)
	assert.Equal(t, "", number.Desc)

	boolean, ok := vars.Get("BOOL")
	require.True(t, ok)
	assert.Equal(t, "", boolean.Desc)

	array, ok := vars.Get("ARRAY")
	require.True(t, ok)
	assert.Equal(t, "", array.Desc)

	mapVar, ok := vars.Get("MAP_VAR")
	require.True(t, ok)
	assert.Equal(t, "", mapVar.Desc)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}
