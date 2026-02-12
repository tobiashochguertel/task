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
	root := tpl.Root
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

// AnalyzeDetailedSteps produces fine-grained step-by-step evaluation traces
// for a template expression. Each step shows either a variable resolution
// or a function application, with the evolving expression state.
func AnalyzeDetailedSteps(input string, data map[string]any, funcs template.FuncMap) []TemplateStep {
	tpl, err := template.New("").Funcs(funcs).Parse(input)
	if err != nil {
		return nil
	}
	root := tpl.Root
	if root == nil {
		return nil
	}

	var steps []TemplateStep
	stepNum := 0
	expr := input // evolving expression

	for _, node := range root.Nodes {
		action, ok := node.(*parse.ActionNode)
		if !ok || action.Pipe == nil {
			continue
		}

		// Process each command in the pipe
		for _, cmd := range action.Pipe.Cmds {
			if len(cmd.Args) == 0 {
				continue
			}

			// Check for variable/field references in arguments
			for _, arg := range cmd.Args {
				switch v := arg.(type) {
				case *parse.FieldNode:
					// Variable resolution: .VARNAME
					varName := "." + strings.Join(v.Ident, ".")
					val := lookupField(data, v.Ident)
					valStr := fmt.Sprintf("%v", val)
					stepNum++

					// Build expression with variable substituted
					newExpr := strings.Replace(expr, varName, fmt.Sprintf("%q", valStr), 1)

					steps = append(steps, TemplateStep{
						StepNum:    stepNum,
						Operation:  "Resolve a Variable",
						Target:     varName,
						Input:      valStr,
						Output:     "",
						Expression: newExpr,
					})
					expr = newExpr
				}
			}

			// Check if this is a function call
			funcName := nodeString(cmd.Args[0])
			if _, isField := cmd.Args[0].(*parse.FieldNode); !isField {
				if _, isIdent := cmd.Args[0].(*parse.IdentifierNode); isIdent {
					// Function application
					// Evaluate partial to get intermediate output
					partial := "{{" + cmd.String() + "}}"
					output := evalPartial(partial, data, funcs)

					// Build input description
					var inputParts []string
					inputParts = append(inputParts, funcName)
					for _, a := range cmd.Args[1:] {
						inputParts = append(inputParts, resolveNodeValue(a, data))
					}
					inputStr := strings.Join(inputParts, " ")

					stepNum++
					steps = append(steps, TemplateStep{
						StepNum:   stepNum,
						Operation: "Apply a Function",
						Target:    funcName,
						Input:     inputStr,
						Output:    output,
					})
				}
			}
		}
	}

	// Set the final expression on the last step
	if len(steps) > 0 {
		steps[len(steps)-1].Expression = input[:strings.Index(input, "{{")] + steps[len(steps)-1].Output +
			input[strings.LastIndex(input, "}}")+2:]
	}

	return steps
}

// multiArgFuncs lists template functions that accept multiple positional
// arguments. When these appear as the first command in a pipe followed by
// more pipe stages, the evaluation order may surprise users.
var multiArgFuncs = map[string]bool{
	"printf":  true,
	"print":   true,
	"println": true,
	"slice":   true,
	"index":   true,
	"eq":      true,
	"ne":      true,
	"lt":      true,
	"le":      true,
	"gt":      true,
	"ge":      true,
}

// GeneratePipeTips analyzes pipe steps and returns user-friendly hints
// about potential pitfalls (e.g. pipe evaluation order with multi-arg
// functions).
func GeneratePipeTips(steps []PipeStep) []string {
	if len(steps) < 2 {
		return nil
	}

	var tips []string

	first := steps[0]
	if multiArgFuncs[first.FuncName] && len(first.Args) > 0 {
		// Multi-arg function piped: the pipe result goes to the LAST
		// argument of the next function. Users sometimes expect it to
		// modify one of the earlier arguments.
		tips = append(tips, fmt.Sprintf(
			"Tip: '%s' result is piped as last arg to '%s'. Use parentheses to change grouping: (func arg1 arg2) | next",
			first.FuncName, steps[1].FuncName,
		))
	}

	return tips
}

// funcSignatures maps common Go template function names to their signatures
// and example usage for display when errors are detected.
var funcSignatures = map[string]struct {
	Signature string
	Example   string
}{
	"printf":      {"printf(format string, args ...any) string", `{{printf "%s: %s" .KEY .VALUE}}`},
	"print":       {"print(args ...any) string", `{{print .A " " .B}}`},
	"println":     {"println(args ...any) string", `{{println .A .B}}`},
	"trim":        {"trim(s string) string", `{{.VAR | trim}}`},
	"trimAll":     {"trimAll(cutset string, s string) string", `{{trimAll "." .VAR}}`},
	"trimPrefix":  {"trimPrefix(prefix string, s string) string", `{{trimPrefix "v" .VERSION}}`},
	"trimSuffix":  {"trimSuffix(suffix string, s string) string", `{{trimSuffix ".exe" .FILE}}`},
	"upper":       {"upper(s string) string", `{{.VAR | upper}}`},
	"lower":       {"lower(s string) string", `{{.VAR | lower}}`},
	"title":       {"title(s string) string", `{{.VAR | title}}`},
	"replace":     {"replace(old string, new string, s string) string", `{{replace "-" "_" .VAR}}`},
	"contains":    {"contains(substr string, s string) bool", `{{if contains "test" .VAR}}...{{end}}`},
	"hasPrefix":   {"hasPrefix(prefix string, s string) bool", `{{if hasPrefix "v" .VERSION}}...{{end}}`},
	"hasSuffix":   {"hasSuffix(suffix string, s string) bool", `{{if hasSuffix ".go" .FILE}}...{{end}}`},
	"split":       {"split(sep string, s string) []string", `{{split "," .LIST}}`},
	"join":        {"join(sep string, list []string) string", `{{join "," .LIST}}`},
	"quote":       {"quote(s string) string", `{{.VAR | quote}}`},
	"squote":      {"squote(s string) string", `{{.VAR | squote}}`},
	"add":         {"add(a, b int) int", `{{add .X 1}}`},
	"sub":         {"sub(a, b int) int", `{{sub .X 1}}`},
	"mul":         {"mul(a, b int) int", `{{mul .X 2}}`},
	"div":         {"div(a, b int) int", `{{div .X 2}}`},
	"mod":         {"mod(a, b int) int", `{{mod .X 2}}`},
	"default":     {"default(defaultVal any, val any) any", `{{default "fallback" .VAR}}`},
	"ternary":     {"ternary(trueVal any, falseVal any, cond bool) any", `{{ternary "yes" "no" .FLAG}}`},
	"toJson":      {"toJson(v any) string", `{{.MAP | toJson}}`},
	"fromJson":    {"fromJson(s string) any", `{{fromJson .JSON_STR}}`},
	"toPrettyJson": {"toPrettyJson(v any) string", `{{.MAP | toPrettyJson}}`},
	"spew":        {"spew(v any) string", `{{.VAR | spew}}`},
	"catLines":    {"catLines(path string) string", `{{catLines .FILE}}`},
	"splitLines":  {"splitLines(s string) []string", `{{splitLines .CONTENT}}`},
	"len":         {"len(v any) int", `{{len .LIST}}`},
	"index":       {"index(collection any, key ...any) any", `{{index .MAP "key"}}`},
	"slice":       {"slice(collection any, indices ...int) any", `{{slice .LIST 0 2}}`},
}

// GenerateErrorHints returns hints with function signatures when template
// output contains error patterns like %!s(MISSING).
func GenerateErrorHints(output string, steps []PipeStep) []string {
	var hints []string

	// Check for MISSING format verb errors
	errorPatterns := []string{
		"%!s(MISSING)", "%!d(MISSING)", "%!v(MISSING)", "%!f(MISSING)",
		"%!q(MISSING)", "%!t(MISSING)",
	}

	hasFormatError := false
	for _, p := range errorPatterns {
		if strings.Contains(output, p) {
			hasFormatError = true
			break
		}
	}

	if hasFormatError {
		// Look for printf-like function in the steps
		for _, step := range steps {
			if sig, ok := funcSignatures[step.FuncName]; ok {
				hints = append(hints, fmt.Sprintf(
					"Hint: %s signature: %s\n    Example: %s",
					step.FuncName, sig.Signature, sig.Example))
				break
			}
		}
		if len(hints) == 0 {
			// Generic hint for printf errors
			if sig, ok := funcSignatures["printf"]; ok {
				hints = append(hints, fmt.Sprintf(
					"Hint: This looks like a printf format error. printf signature: %s\n    Example: %s",
					sig.Signature, sig.Example))
			}
		}
	}

	return hints
}

// numericFuncs lists template functions that require numeric arguments.
var numericFuncs = map[string]bool{
	"add": true, "sub": true, "mul": true, "div": true, "mod": true,
	"max": true, "min": true, "ceil": true, "floor": true, "round": true,
}

// DetectTypeMismatches inspects a parsed template AST and the data context to
// detect cases where a function receives arguments of the wrong type (e.g.
// add with a string argument). Returns human-readable warning strings.
func DetectTypeMismatches(input string, data map[string]any, funcs template.FuncMap) []string {
	tpl, err := template.New("").Funcs(funcs).Parse(input)
	if err != nil {
		return nil
	}
	root := tpl.Root
	if root == nil {
		return nil
	}

	var warnings []string
	for _, node := range root.Nodes {
		action, ok := node.(*parse.ActionNode)
		if !ok || action.Pipe == nil {
			continue
		}
		for _, cmd := range action.Pipe.Cmds {
			warnings = append(warnings, checkTypeMismatch(cmd, data)...)
		}
	}
	return warnings
}

// checkTypeMismatch checks a single command node for type mismatches.
func checkTypeMismatch(cmd *parse.CommandNode, data map[string]any) []string {
	if len(cmd.Args) == 0 {
		return nil
	}

	funcName := nodeString(cmd.Args[0])
	if !numericFuncs[funcName] {
		return nil
	}

	var warnings []string
	for _, arg := range cmd.Args[1:] {
		field, ok := arg.(*parse.FieldNode)
		if !ok {
			continue
		}
		val := lookupField(data, field.Ident)
		if val == nil {
			continue
		}
		if !isNumericType(val) {
			warnings = append(warnings, fmt.Sprintf(
				"⚠ Type mismatch: %s() expects numeric arguments, but .%s is %T (%q)",
				funcName, strings.Join(field.Ident, "."), val, fmt.Sprintf("%v", val),
			))
		}
	}
	return warnings
}

// isNumericType returns true if the value is a numeric type that Go template
// math functions can operate on.
func isNumericType(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	}
	return false
}
