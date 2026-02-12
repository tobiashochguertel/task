package ast

import (
	"go.yaml.in/yaml/v4"

	"github.com/go-task/task/v3/errors"
)

// Var represents either a static or dynamic variable.
type Var struct {
	Value any
	Live  any
	Sh    *string
	Ref   string
	Dir   string
	Desc  string
}

func (v *Var) UnmarshalYAML(node *yaml.Node) error {
	switch node.Kind {
	case yaml.MappingNode:
		// Try to decode as a complex variable with sh/ref/map/value/desc
		var m struct {
			Sh    *string
			Ref   string
			Map   any
			Value any
			Desc  string
		}
		if err := node.Decode(&m); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}

		// Check if any of the expected fields are set
		if m.Sh != nil || m.Ref != "" || m.Map != nil || m.Value != nil || m.Desc != "" {
			// Validate mutually exclusive fields
			exclusiveCount := 0
			if m.Sh != nil {
				exclusiveCount++
			}
			if m.Ref != "" {
				exclusiveCount++
			}
			if m.Map != nil {
				exclusiveCount++
			}
			if m.Value != nil {
				exclusiveCount++
			}

			if exclusiveCount > 1 {
				return errors.NewTaskfileDecodeError(nil, node).WithMessage(
					`variable cannot have more than one of: "sh", "ref", "map", "value"`)
			}

			v.Sh = m.Sh
			v.Ref = m.Ref
			v.Desc = m.Desc

			// Set the value based on which type is present
			if m.Map != nil {
				v.Value = m.Map
			} else if m.Value != nil {
				v.Value = m.Value
			}

			return nil
		}

		// If none of the expected fields are set, this is an error
		key := "<none>"
		if len(node.Content) > 0 {
			key = node.Content[0].Value
		}
		return errors.NewTaskfileDecodeError(nil, node).WithMessage(`%q is not a valid variable type. Try "sh", "ref", "map", "value", "desc" or using a scalar value`, key)
	default:
		var value any
		if err := node.Decode(&value); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		v.Value = value
		return nil
	}
}
