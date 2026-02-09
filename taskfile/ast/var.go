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
		// Try to decode as a complex variable with sh/ref/map/desc
		var m struct {
			Sh   *string
			Ref  string
			Map  any
			Desc string
		}
		if err := node.Decode(&m); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		
		// Check if any of the expected fields are set
		if m.Sh != nil || m.Ref != "" || m.Map != nil || m.Desc != "" {
			v.Sh = m.Sh
			v.Ref = m.Ref
			v.Value = m.Map
			v.Desc = m.Desc
			return nil
		}
		
		// If none of the expected fields are set, this is an error
		key := "<none>"
		if len(node.Content) > 0 {
			key = node.Content[0].Value
		}
		return errors.NewTaskfileDecodeError(nil, node).WithMessage(`%q is not a valid variable type. Try "sh", "ref", "map", "desc" or using a scalar value`, key)
	default:
		var value any
		if err := node.Decode(&value); err != nil {
			return errors.NewTaskfileDecodeError(err, node)
		}
		v.Value = value
		return nil
	}
}
