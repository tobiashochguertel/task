package ast_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/go-task/task/v3/taskfile/ast"
)

func stringPtr(s string) *string {
	return &s
}

func TestVarsMerge(t *testing.T) {
	t.Parallel()

	t.Run("merge simple values", func(t *testing.T) {
		v1 := ast.NewVars()
		v1.Set("VAR1", ast.Var{Value: "value1"})
		
		v2 := ast.NewVars()
		v2.Set("VAR2", ast.Var{Value: "value2"})
		
		v1.Merge(v2, nil)
		
		val1, ok := v1.Get("VAR1")
		assert.True(t, ok)
		assert.Equal(t, "value1", val1.Value)
		
		val2, ok := v1.Get("VAR2")
		assert.True(t, ok)
		assert.Equal(t, "value2", val2.Value)
	})

	t.Run("child overrides parent value", func(t *testing.T) {
		v1 := ast.NewVars()
		v1.Set("VAR", ast.Var{Value: "parent"})
		
		v2 := ast.NewVars()
		v2.Set("VAR", ast.Var{Value: "child"})
		
		v1.Merge(v2, nil)
		
		val, ok := v1.Get("VAR")
		assert.True(t, ok)
		assert.Equal(t, "child", val.Value)
	})
}

func TestVarsMergeWithDescriptions(t *testing.T) {
	t.Parallel()

	t.Run("child inherits parent description", func(t *testing.T) {
		parent := ast.NewVars()
		parent.Set("APP_NAME", ast.Var{
			Desc: "Application name",
			Sh:   stringPtr("echo parent"),
		})

		child := ast.NewVars()
		child.Set("APP_NAME", ast.Var{
			Desc: "", // Empty description
			Sh:   stringPtr("echo child"),
		})

		parent.Merge(child, nil)
		
		val, ok := parent.Get("APP_NAME")
		assert.True(t, ok)
		assert.Equal(t, "Application name", val.Desc, "Should inherit parent description")
		assert.Equal(t, "echo child", *val.Sh, "Should use child value")
	})

	t.Run("child overrides parent description", func(t *testing.T) {
		parent := ast.NewVars()
		parent.Set("APP_NAME", ast.Var{
			Desc: "Parent description",
			Sh:   stringPtr("echo parent"),
		})

		child := ast.NewVars()
		child.Set("APP_NAME", ast.Var{
			Desc: "Child description",
			Sh:   stringPtr("echo child"),
		})

		parent.Merge(child, nil)
		
		val, ok := parent.Get("APP_NAME")
		assert.True(t, ok)
		assert.Equal(t, "Child description", val.Desc, "Should override with child description")
		assert.Equal(t, "echo child", *val.Sh)
	})

	t.Run("multi-level inheritance", func(t *testing.T) {
		// Global vars
		global := ast.NewVars()
		global.Set("VERSION", ast.Var{
			Desc: "Global version",
			Sh:   stringPtr("echo 1.0.0"),
		})

		// Include vars (no description)
		include := ast.NewVars()
		include.Set("VERSION", ast.Var{
			Sh: stringPtr("echo 2.0.0"),
		})

		// Task vars (override description)
		taskVars := ast.NewVars()
		taskVars.Set("VERSION", ast.Var{
			Desc: "Task-specific version",
			Sh:   stringPtr("echo 3.0.0"),
		})

		// Merge chain: global -> include -> task
		global.Merge(include, nil)
		val, _ := global.Get("VERSION")
		assert.Equal(t, "Global version", val.Desc, "Include should inherit from global")
		assert.Equal(t, "echo 2.0.0", *val.Sh)

		global.Merge(taskVars, nil)
		val, _ = global.Get("VERSION")
		assert.Equal(t, "Task-specific version", val.Desc, "Task should override")
		assert.Equal(t, "echo 3.0.0", *val.Sh)
	})

	t.Run("preserve non-empty description when merging value-only var", func(t *testing.T) {
		parent := ast.NewVars()
		parent.Set("PORT", ast.Var{
			Desc:  "Server port",
			Value: 8080,
		})

		child := ast.NewVars()
		child.Set("PORT", ast.Var{
			Value: 9090, // No description
		})

		parent.Merge(child, nil)
		
		val, ok := parent.Get("PORT")
		assert.True(t, ok)
		assert.Equal(t, "Server port", val.Desc, "Should keep parent description")
		assert.Equal(t, 9090, val.Value, "Should use child value")
	})

	t.Run("description-only var doesn't replace existing value", func(t *testing.T) {
		parent := ast.NewVars()
		parent.Set("CONFIG", ast.Var{
			Value: map[string]any{"key": "value"},
		})

		child := ast.NewVars()
		child.Set("CONFIG", ast.Var{
			Desc: "Configuration settings",
			// No value set
		})

		parent.Merge(child, nil)
		
		val, ok := parent.Get("CONFIG")
		assert.True(t, ok)
		assert.Equal(t, "Configuration settings", val.Desc)
		assert.Nil(t, val.Value, "Description-only var should override with nil value")
	})
}

func TestVarsMergeEdgeCases(t *testing.T) {
	t.Parallel()

	t.Run("merge empty vars", func(t *testing.T) {
		v1 := ast.NewVars()
		v2 := ast.NewVars()
		
		v1.Merge(v2, nil)
		
		assert.Equal(t, 0, v1.Len())
	})

	t.Run("merge with nil", func(t *testing.T) {
		v1 := ast.NewVars()
		v1.Set("VAR", ast.Var{Value: "value"})
		
		// Should not panic
		v1.Merge(nil, nil)
		
		val, ok := v1.Get("VAR")
		assert.True(t, ok)
		assert.Equal(t, "value", val.Value)
	})

	t.Run("multiple vars with descriptions", func(t *testing.T) {
		parent := ast.NewVars()
		parent.Set("VAR1", ast.Var{Desc: "Desc 1", Value: "val1"})
		parent.Set("VAR2", ast.Var{Desc: "Desc 2", Value: "val2"})
		parent.Set("VAR3", ast.Var{Value: "val3"}) // No description

		child := ast.NewVars()
		child.Set("VAR1", ast.Var{Value: "new1"}) // Should inherit Desc 1
		child.Set("VAR2", ast.Var{Desc: "New Desc 2", Value: "new2"}) // Should override
		child.Set("VAR3", ast.Var{Desc: "New Desc 3", Value: "new3"}) // Should add description

		parent.Merge(child, nil)
		
		val1, _ := parent.Get("VAR1")
		assert.Equal(t, "Desc 1", val1.Desc)
		assert.Equal(t, "new1", val1.Value)
		
		val2, _ := parent.Get("VAR2")
		assert.Equal(t, "New Desc 2", val2.Desc)
		assert.Equal(t, "new2", val2.Value)
		
		val3, _ := parent.Get("VAR3")
		assert.Equal(t, "New Desc 3", val3.Desc)
		assert.Equal(t, "new3", val3.Value)
	})
}
