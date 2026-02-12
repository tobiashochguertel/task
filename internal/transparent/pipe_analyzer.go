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

// AnalyzeEvalActions produces action-grouped evaluation traces for a template.
// Each EvalAction corresponds to one {{...}} expression in the template,
// with source line tracking and correctly ordered steps that follow
// Go's template evaluation order (depth-first for sub-pipelines).
func AnalyzeEvalActions(input string, data map[string]any, funcs template.FuncMap) []EvalAction {
	tpl, err := template.New("").Funcs(funcs).Parse(input)
	if err != nil {
		return nil
	}
	root := tpl.Root
	if root == nil {
		return nil
	}

	inputLines := strings.Split(input, "\n")
	var actions []EvalAction
	stepNum := 0
	actionIdx := 0

	for _, node := range root.Nodes {
		action, ok := node.(*parse.ActionNode)
		if !ok || action.Pipe == nil {
			continue
		}

		// Compute source line number from node position
		lineNum := lineNumber(input, action.Position())
		srcLine := ""
		if lineNum >= 1 && lineNum <= len(inputLines) {
			srcLine = inputLines[lineNum-1]
		}

		// Walk the pipe to generate steps
		steps := walkPipeCommands(action.Pipe.Cmds, data, funcs, &stepNum)

		// Evaluate the full action to get the result value
		actionExpr := "{{" + action.Pipe.String() + "}}"
		resultVal := evalPartial(actionExpr, data, funcs)

		// Build the result line by replacing the {{...}} in the source line
		resultLine := srcLine
		if srcLine != "" {
			// Find the {{...}} in the source line and replace with result
			startIdx := strings.Index(srcLine, "{{")
			endIdx := strings.LastIndex(srcLine, "}}")
			if startIdx >= 0 && endIdx >= 0 && endIdx+2 <= len(srcLine) {
				resultLine = srcLine[:startIdx] + resultVal + srcLine[endIdx+2:]
			}
		}

		actions = append(actions, EvalAction{
			ActionIndex: actionIdx,
			SourceLine:  lineNum,
			Source:      srcLine,
			Result:      resultLine,
			Steps:       steps,
		})
		actionIdx++
	}

	return actions
}

// lineNumber computes the 1-based line number for a given byte offset position.
func lineNumber(input string, pos parse.Pos) int {
	offset := int(pos)
	if offset > len(input) {
		offset = len(input)
	}
	return strings.Count(input[:offset], "\n") + 1
}

// walkPipeCommands recursively walks a slice of pipe commands and generates
// TemplateStep entries in the correct evaluation order:
// 1. For each command, process arguments first (depth-first for sub-pipelines)
// 2. Then record the function application or variable resolution
//
// cmds is the full list of commands in this pipe. Each function application
// evaluates the partial pipe from cmds[0..i] to get the correct piped output.
func walkPipeCommands(cmds []*parse.CommandNode, data map[string]any, funcs template.FuncMap, stepNum *int) []TemplateStep {
	var steps []TemplateStep

	for i, cmd := range cmds {
		if len(cmd.Args) == 0 {
			continue
		}

		firstArg := cmd.Args[0]

		// Check if this command is a function call (first arg is identifier)
		if ident, isIdent := firstArg.(*parse.IdentifierNode); isIdent {
			// Process non-first arguments first (depth-first)
			for _, arg := range cmd.Args[1:] {
				steps = append(steps, walkArg(arg, data, funcs, stepNum)...)
			}

			// Build input description with resolved argument values
			var inputParts []string
			inputParts = append(inputParts, ident.Ident)
			// If this is not the first command in the pipe, show the piped input
			if i > 0 {
				pipedVal := evalPartial(partialPipeString(cmds[:i]), data, funcs)
				inputParts = append(inputParts, fmt.Sprintf("%q", pipedVal))
			}
			for _, a := range cmd.Args[1:] {
				inputParts = append(inputParts, resolveArgValue(a, data, funcs))
			}
			inputStr := strings.Join(inputParts, " ")

			// Evaluate the partial pipe up to and including this command
			// so that piped values flow correctly through the chain
			partial := partialPipeString(cmds[:i+1])
			output := evalPartial(partial, data, funcs)

			*stepNum++
			steps = append(steps, TemplateStep{
				StepNum:   *stepNum,
				Operation: "Apply a Function",
				Target:    ident.Ident,
				Input:     inputStr,
				Output:    output,
			})
		} else if field, isField := firstArg.(*parse.FieldNode); isField {
			// Variable access at start of pipe: .VARNAME
			varName := "." + strings.Join(field.Ident, ".")
			val := lookupField(data, field.Ident)
			valStr := fmt.Sprintf("%v", val)

			*stepNum++
			steps = append(steps, TemplateStep{
				StepNum:   *stepNum,
				Operation: "Resolve a Variable",
				Target:    varName,
				Input:     valStr,
			})
		}
	}

	return steps
}

// resolveArgValue returns the resolved value of an argument node for display
// in the Input field. Unlike resolveNodeValue, this evaluates sub-pipelines
// to show their result rather than raw AST text.
func resolveArgValue(n parse.Node, data map[string]any, funcs template.FuncMap) string {
	switch v := n.(type) {
	case *parse.PipeNode:
		// Sub-pipeline: evaluate it to get the result
		return evalPartial("{{"+v.String()+"}}", data, funcs)
	case *parse.FieldNode:
		val := lookupField(data, v.Ident)
		return fmt.Sprintf("%v", val)
	default:
		return resolveNodeValue(n, data)
	}
}

// walkArg processes a single argument node, generating steps for variable
// resolutions and recursing into sub-pipelines.
func walkArg(arg parse.Node, data map[string]any, funcs template.FuncMap, stepNum *int) []TemplateStep {
	switch v := arg.(type) {
	case *parse.FieldNode:
		varName := "." + strings.Join(v.Ident, ".")
		val := lookupField(data, v.Ident)
		valStr := fmt.Sprintf("%v", val)
		*stepNum++
		return []TemplateStep{{
			StepNum:   *stepNum,
			Operation: "Resolve a Variable",
			Target:    varName,
			Input:     valStr,
		}}
	case *parse.PipeNode:
		// Sub-pipeline: recursively walk its commands
		return walkPipeCommands(v.Cmds, data, funcs, stepNum)
	default:
		// Literals (string, number, etc.) don't generate steps
		return nil
	}
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
	"printf":       {"printf(format string, args ...any) string", `{{printf "%s: %s" .KEY .VALUE}}`},
	"print":        {"print(args ...any) string", `{{print .A " " .B}}`},
	"println":      {"println(args ...any) string", `{{println .A .B}}`},
	"trim":         {"trim(s string) string", `{{.VAR | trim}}`},
	"trimAll":      {"trimAll(cutset string, s string) string", `{{trimAll "." .VAR}}`},
	"trimPrefix":   {"trimPrefix(prefix string, s string) string", `{{trimPrefix "v" .VERSION}}`},
	"trimSuffix":   {"trimSuffix(suffix string, s string) string", `{{trimSuffix ".exe" .FILE}}`},
	"upper":        {"upper(s string) string", `{{.VAR | upper}}`},
	"lower":        {"lower(s string) string", `{{.VAR | lower}}`},
	"title":        {"title(s string) string", `{{.VAR | title}}`},
	"replace":      {"replace(old string, new string, s string) string", `{{replace "-" "_" .VAR}}`},
	"contains":     {"contains(substr string, s string) bool", `{{if contains "test" .VAR}}...{{end}}`},
	"hasPrefix":    {"hasPrefix(prefix string, s string) bool", `{{if hasPrefix "v" .VERSION}}...{{end}}`},
	"hasSuffix":    {"hasSuffix(suffix string, s string) bool", `{{if hasSuffix ".go" .FILE}}...{{end}}`},
	"split":        {"split(sep string, s string) []string", `{{split "," .LIST}}`},
	"join":         {"join(sep string, list []string) string", `{{join "," .LIST}}`},
	"quote":        {"quote(s string) string", `{{.VAR | quote}}`},
	"squote":       {"squote(s string) string", `{{.VAR | squote}}`},
	"add":          {"add(a, b int) int", `{{add .X 1}}`},
	"sub":          {"sub(a, b int) int", `{{sub .X 1}}`},
	"mul":          {"mul(a, b int) int", `{{mul .X 2}}`},
	"div":          {"div(a, b int) int", `{{div .X 2}}`},
	"mod":          {"mod(a, b int) int", `{{mod .X 2}}`},
	"default":      {"default(defaultVal any, val any) any", `{{default "fallback" .VAR}}`},
	"ternary":      {"ternary(trueVal any, falseVal any, cond bool) any", `{{ternary "yes" "no" .FLAG}}`},
	"toJson":       {"toJson(v any) string", `{{.MAP | toJson}}`},
	"fromJson":     {"fromJson(s string) any", `{{fromJson .JSON_STR}}`},
	"toPrettyJson": {"toPrettyJson(v any) string", `{{.MAP | toPrettyJson}}`},
	"spew":         {"spew(v any) string", `{{.VAR | spew}}`},
	"catLines":     {"catLines(path string) string", `{{catLines .FILE}}`},
	"splitLines":   {"splitLines(s string) []string", `{{splitLines .CONTENT}}`},
	"len":          {"len(v any) int", `{{len .LIST}}`},
	"index":        {"index(collection any, key ...any) any", `{{index .MAP "key"}}`},
	"slice":        {"slice(collection any, indices ...int) any", `{{slice .LIST 0 2}}`},
}

// sigParam describes one parameter from a parsed function signature.
type sigParam struct {
	Name     string // e.g. "format", "args"
	Type     string // e.g. "string", "...any"
	Variadic bool   // true if the param is variadic (starts with ...)
}

// parseSigParams extracts parameter names and types from a signature string
// like "printf(format string, args ...any) string".
func parseSigParams(sig string) []sigParam {
	start := strings.Index(sig, "(")
	end := strings.LastIndex(sig, ")")
	if start < 0 || end < 0 || end <= start+1 {
		return nil
	}
	paramStr := sig[start+1 : end]
	parts := strings.Split(paramStr, ",")
	var params []sigParam
	for _, p := range parts {
		p = strings.TrimSpace(p)
		fields := strings.Fields(p)
		if len(fields) >= 2 {
			name := fields[0]
			typ := strings.Join(fields[1:], " ")
			variadic := strings.HasPrefix(typ, "...")
			params = append(params, sigParam{Name: name, Type: typ, Variadic: variadic})
		}
	}
	return params
}

// buildCallDetail builds a multi-line string showing the parameter-to-value
// mapping for a function call that produced an error. It uses the function
// signature from funcSignatures and the resolved argument values.
func buildCallDetail(funcName string, argValues []string) string {
	sig, ok := funcSignatures[funcName]
	if !ok {
		return ""
	}
	params := parseSigParams(sig.Signature)
	if len(params) == 0 {
		return ""
	}

	var b strings.Builder
	b.WriteString(fmt.Sprintf("Hint: %s format error detected\n", funcName))
	b.WriteString(fmt.Sprintf("    Signature: %s\n", sig.Signature))

	// Build the call line: funcName(arg1, arg2, ...)
	b.WriteString(fmt.Sprintf("    Call:      %s(%s)\n", funcName, strings.Join(argValues, ", ")))

	// Map each argument value to the corresponding parameter name
	b.WriteString("    Params:")
	argIdx := 0
	for _, param := range params {
		if param.Variadic {
			// Variadic parameter consumes all remaining args
			if argIdx < len(argValues) {
				for vi := 0; argIdx < len(argValues); vi++ {
					b.WriteString(fmt.Sprintf("\n             %s[%d] = %s", param.Name, vi, argValues[argIdx]))
					argIdx++
				}
			} else {
				b.WriteString(fmt.Sprintf("\n             %s = (none provided)", param.Name))
			}
		} else {
			if argIdx < len(argValues) {
				b.WriteString(fmt.Sprintf("\n             %s = %s", param.Name, argValues[argIdx]))
				argIdx++
			} else {
				b.WriteString(fmt.Sprintf("\n             %s = ⚠ MISSING", param.Name))
			}
		}
	}
	b.WriteString(fmt.Sprintf("\n    Example:  %s", sig.Example))

	return b.String()
}

// GenerateErrorHints returns hints with function signatures when template
// output contains error patterns like %!s(MISSING).
// When evalActions are provided, the hints include the actual parameter-to-value
// mapping showing exactly how the function was called.
func GenerateErrorHints(output string, steps []PipeStep, evalActions []EvalAction) []string {
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

	if !hasFormatError {
		return hints
	}

	// Try to build a detailed call hint using EvalAction steps + PipeStep arg values
	for _, ea := range evalActions {
		for _, ds := range ea.Steps {
			if ds.Operation != "Apply a Function" {
				continue
			}
			if _, ok := funcSignatures[ds.Target]; !ok {
				continue
			}
			// Check if this function step's output contains the error
			if !containsFormatError(ds.Output) && !containsFormatError(ea.Result) {
				continue
			}
			// Find matching PipeStep to get individual resolved argument values
			var argValues []string
			for _, ps := range steps {
				if ps.FuncName == ds.Target {
					argValues = ps.ArgsValues
					break
				}
			}
			if len(argValues) == 0 && ds.Input != "" {
				// Fallback: use the Input field as a single representation
				argValues = parseInputArgs(ds.Input, ds.Target)
			}
			detail := buildCallDetail(ds.Target, argValues)
			if detail != "" {
				hints = append(hints, detail)
				return hints
			}
		}
	}

	// Fallback: use PipeStep data if EvalActions didn't produce a match
	for _, step := range steps {
		if sig, ok := funcSignatures[step.FuncName]; ok {
			if len(step.ArgsValues) > 0 {
				detail := buildCallDetail(step.FuncName, step.ArgsValues)
				if detail != "" {
					hints = append(hints, detail)
					return hints
				}
			}
			hints = append(hints, fmt.Sprintf(
				"Hint: %s signature: %s\n    Example: %s",
				step.FuncName, sig.Signature, sig.Example))
			return hints
		}
	}

	// Generic fallback for printf errors
	if sig, ok := funcSignatures["printf"]; ok {
		hints = append(hints, fmt.Sprintf(
			"Hint: This looks like a printf format error. printf signature: %s\n    Example: %s",
			sig.Signature, sig.Example))
	}

	return hints
}

// containsFormatError checks if a string contains Go format error patterns.
func containsFormatError(s string) bool {
	patterns := []string{
		"%!s(MISSING)", "%!d(MISSING)", "%!v(MISSING)", "%!f(MISSING)",
		"%!q(MISSING)", "%!t(MISSING)",
	}
	for _, p := range patterns {
		if strings.Contains(s, p) {
			return true
		}
	}
	return false
}

// parseInputArgs splits an Input string like `printf "%s: %s" "hello" "world"`
// into individual argument values (excluding the function name).
func parseInputArgs(input, funcName string) []string {
	// Strip the function name prefix
	rest := strings.TrimPrefix(input, funcName)
	rest = strings.TrimSpace(rest)
	if rest == "" {
		return nil
	}

	// Simple tokenizer: split by spaces but respect quoted strings
	var args []string
	var current strings.Builder
	inQuote := false
	quoteChar := byte(0)

	for i := 0; i < len(rest); i++ {
		ch := rest[i]
		if inQuote {
			current.WriteByte(ch)
			if ch == quoteChar {
				inQuote = false
			}
		} else if ch == '"' || ch == '\'' {
			inQuote = true
			quoteChar = ch
			current.WriteByte(ch)
		} else if ch == ' ' || ch == '\t' {
			if current.Len() > 0 {
				args = append(args, current.String())
				current.Reset()
			}
		} else {
			current.WriteByte(ch)
		}
	}
	if current.Len() > 0 {
		args = append(args, current.String())
	}

	return args
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
