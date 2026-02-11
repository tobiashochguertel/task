package transparent

import (
	"bytes"
	"fmt"
	"strings"
	"text/template/parse"

	"github.com/go-task/template"
)

// AnalyzePipes parses a template string and returns PipeStep entries for
// each pipe action that contains more than one command (i.e. uses the |
// operator). Each step records the function name, its raw arguments and
// the intermediate result obtained by executing the partial pipe up to
// that step.
//
// funcs must be the same FuncMap used for normal template execution so
// that intermediate evaluation produces the same results.
func AnalyzePipes(input string, data map[string]any, funcs template.FuncMap) []PipeStep {
	tpl, err := template.New("").Funcs(funcs).Parse(input)
	if err != nil {
		return nil
	}
	root := tpl.Tree.Root
	if root == nil {
		return nil
	}

	var steps []PipeStep
	for _, node := range root.Nodes {
		action, ok := node.(*parse.ActionNode)
		if !ok || action.Pipe == nil {
			continue
		}
		steps = append(steps, analyzePipe(action.Pipe, data, funcs)...)
	}
	return steps
}

// analyzePipe extracts step-by-step details from a single PipeNode.
// Only pipes with ≥2 commands are interesting (single-command pipes
// have no intermediate results).
func analyzePipe(pipe *parse.PipeNode, data map[string]any, funcs template.FuncMap) []PipeStep {
	if len(pipe.Cmds) < 2 {
		return nil
	}

	var steps []PipeStep

	// Build partial pipes of increasing length and evaluate each one to
	// obtain the intermediate result.
	for i, cmd := range pipe.Cmds {
		funcName, args := describeCommand(cmd)

		// Build a partial template containing only the first i+1 commands.
		partial := partialPipeString(pipe.Cmds[:i+1])
		output := evalPartial(partial, data, funcs)

		steps = append(steps, PipeStep{
			FuncName:   funcName,
			Args:       args,
			ArgsValues: resolveArgs(cmd, data),
			Output:     output,
		})
	}
	return steps
}

// describeCommand extracts the function/field name and raw argument
// representations from a CommandNode.
func describeCommand(cmd *parse.CommandNode) (string, []string) {
	if len(cmd.Args) == 0 {
		return "", nil
	}

	funcName := nodeString(cmd.Args[0])
	var args []string
	for _, arg := range cmd.Args[1:] {
		args = append(args, nodeString(arg))
	}
	return funcName, args
}

// resolveArgs evaluates each argument node individually to produce its
// resolved value representation.
func resolveArgs(cmd *parse.CommandNode, data map[string]any) []string {
	var vals []string
	for _, arg := range cmd.Args[1:] {
		vals = append(vals, resolveNodeValue(arg, data))
	}
	return vals
}

// partialPipeString reconstructs a template expression string from a
// slice of CommandNodes: {{cmd0 | cmd1 | ... | cmdN}}.
func partialPipeString(cmds []*parse.CommandNode) string {
	var parts []string
	for _, cmd := range cmds {
		parts = append(parts, cmd.String())
	}
	return "{{" + strings.Join(parts, " | ") + "}}"
}

// evalPartial executes a small template fragment and returns the result
// string, or an error placeholder if evaluation fails.
func evalPartial(tmpl string, data map[string]any, funcs template.FuncMap) string {
	t, err := template.New("").Funcs(funcs).Parse(tmpl)
	if err != nil {
		return fmt.Sprintf("<parse error: %v>", err)
	}
	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return fmt.Sprintf("<exec error: %v>", err)
	}
	return strings.ReplaceAll(buf.String(), "<no value>", "")
}

// nodeString returns a human-readable representation of a parse node.
func nodeString(n parse.Node) string {
	return n.String()
}

// resolveNodeValue attempts to produce the resolved value of a node
// given the template data context. For field/variable nodes it looks up
// the value; for literals it returns the literal text.
func resolveNodeValue(n parse.Node, data map[string]any) string {
	switch v := n.(type) {
	case *parse.FieldNode:
		// .FOO or .FOO.BAR
		val := lookupField(data, v.Ident)
		return fmt.Sprintf("%v", val)
	case *parse.VariableNode:
		// $var
		return v.String()
	case *parse.DotNode:
		return fmt.Sprintf("%v", data)
	default:
		return v.String()
	}
}

// lookupField traverses a map following the ident chain (e.g. ["FOO","BAR"]).
func lookupField(data map[string]any, ident []string) any {
	var cur any = data
	for _, key := range ident {
		m, ok := cur.(map[string]any)
		if !ok {
			return nil
		}
		cur = m[key]
	}
	return cur
}
