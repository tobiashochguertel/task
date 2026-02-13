package templater

import (
	"bytes"
	"fmt"
	"maps"
	"strings"

	"github.com/go-task/template"

	"github.com/go-task/task/v3/internal/deepcopy"
	"github.com/go-task/task/v3/internal/transparent"
	"github.com/go-task/task/v3/taskfile/ast"
)

// Cache is a help struct that allow us to call "replaceX" funcs multiple
// times, without having to check for error each time. The first error that
// happen will be assigned to r.err, and consecutive calls to funcs will just
// return the zero value.
type Cache struct {
	Vars   *ast.Vars
	Tracer *transparent.Tracer

	cacheMap map[string]any
	err      error
}

func (r *Cache) ResetCache() {
	r.cacheMap = r.Vars.ToCacheMap()
}

func (r *Cache) Err() error {
	return r.err
}

func ResolveRef(ref string, cache *Cache) any {
	// If there is already an error, do nothing
	if cache.err != nil {
		return nil
	}

	// Initialize the cache map if it's not already initialized
	if cache.cacheMap == nil {
		cache.cacheMap = cache.Vars.ToCacheMap()
	}

	if ref == "." {
		return cache.cacheMap
	}
	t, err := template.New("resolver").Funcs(templateFuncs).Parse(fmt.Sprintf("{{%s}}", ref))
	if err != nil {
		cache.err = err
		return nil
	}
	val, err := t.Resolve(cache.cacheMap)
	if err != nil {
		cache.err = err
		return nil
	}
	return val
}

func Replace[T any](v T, cache *Cache) T {
	return ReplaceWithExtra(v, cache, nil)
}

func ReplaceWithExtra[T any](v T, cache *Cache, extra map[string]any) T {
	// If there is already an error, do nothing
	if cache.err != nil {
		return v
	}

	// Initialize the cache map if it's not already initialized
	if cache.cacheMap == nil {
		cache.cacheMap = cache.Vars.ToCacheMap()
	}

	// Create a copy of the cache map to avoid editing the original
	// If there is extra data, merge it with the cache map
	data := maps.Clone(cache.cacheMap)
	if extra != nil {
		maps.Copy(data, extra)
	}

	// Traverse the value and parse any template variables
	copy, err := deepcopy.TraverseStringsFunc(v, func(v string) (string, error) {
		tpl, err := template.New("").Funcs(templateFuncs).Parse(v)
		if err != nil {
			return v, err
		}
		var b bytes.Buffer
		if err := tpl.Execute(&b, data); err != nil {
			return v, err
		}
		result := strings.ReplaceAll(b.String(), "<no value>", "")

		// Record template trace if tracer is active and input contained template delimiters
		if cache.Tracer != nil && strings.Contains(v, "{{") {
			pipeSteps := transparent.AnalyzePipes(v, data, template.FuncMap(templateFuncs))
			evalActions := transparent.AnalyzeEvalActions(v, data, template.FuncMap(templateFuncs))
			trace := transparent.TemplateTrace{
				Input:       v,
				Output:      result,
				VarsUsed:    extractVarNames(v),
				Steps:       pipeSteps,
				EvalActions: evalActions,
				Tips:        transparent.GeneratePipeTips(pipeSteps),
			}
			// Detect undefined variables (replaced <no value> → "")
			if strings.Contains(b.String(), "<no value>") {
				trace.Error = "warning: template produced <no value> for one or more variables (replaced with empty string)"
			}
			// Detect type mismatches (e.g. add with string args)
			typeWarnings := transparent.DetectTypeMismatches(v, data, template.FuncMap(templateFuncs))
			trace.Tips = append(trace.Tips, typeWarnings...)
			// Collect structured diagnostics (exec errors + output anomalies)
			trace.Diagnostics = transparent.CollectDiagnostics(evalActions, pipeSteps)
			// Generate function signature hints when errors are detected
			errorHints := transparent.GenerateErrorHints(result, pipeSteps, evalActions)
			trace.Tips = append(trace.Tips, errorHints...)
			cache.Tracer.RecordTemplate(trace)
		}

		return result, nil
	})
	if err != nil {
		cache.err = err
		return v
	}

	return copy
}

func ReplaceGlobs(globs []*ast.Glob, cache *Cache) []*ast.Glob {
	if cache.err != nil || len(globs) == 0 {
		return nil
	}

	new := make([]*ast.Glob, len(globs))
	for i, g := range globs {
		new[i] = &ast.Glob{
			Glob:   Replace(g.Glob, cache),
			Negate: g.Negate,
		}
	}
	return new
}

func ReplaceVar(v ast.Var, cache *Cache) ast.Var {
	return ReplaceVarWithExtra(v, cache, nil)
}

func ReplaceVarWithExtra(v ast.Var, cache *Cache, extra map[string]any) ast.Var {
	if v.Ref != "" {
		return ast.Var{Value: ResolveRef(v.Ref, cache)}
	}
	return ast.Var{
		Value: ReplaceWithExtra(v.Value, cache, extra),
		Sh:    ReplaceWithExtra(v.Sh, cache, extra),
		Live:  v.Live,
		Ref:   v.Ref,
		Dir:   v.Dir,
	}
}

func ReplaceVars(vars *ast.Vars, cache *Cache) *ast.Vars {
	return ReplaceVarsWithExtra(vars, cache, nil)
}

func ReplaceVarsWithExtra(vars *ast.Vars, cache *Cache, extra map[string]any) *ast.Vars {
	if cache.err != nil || vars.Len() == 0 {
		return nil
	}

	newVars := ast.NewVars()
	for k, v := range vars.All() {
		newVars.Set(k, ReplaceVarWithExtra(v, cache, extra))
	}

	return newVars
}

// extractVarNames extracts variable names like .FOO from a template string.
func extractVarNames(tmpl string) []string {
	var names []string
	seen := make(map[string]bool)
	i := 0
	for i < len(tmpl) {
		// Find .VARNAME patterns (preceded by space, paren, or start of action)
		if tmpl[i] == '.' && i+1 < len(tmpl) && isUpperOrUnderscore(tmpl[i+1]) {
			j := i + 1
			for j < len(tmpl) && isVarChar(tmpl[j]) {
				j++
			}
			name := tmpl[i+1 : j]
			if !seen[name] {
				seen[name] = true
				names = append(names, name)
			}
			i = j
		} else {
			i++
		}
	}
	return names
}

func isUpperOrUnderscore(b byte) bool {
	return (b >= 'A' && b <= 'Z') || b == '_'
}

func isVarChar(b byte) bool {
	return (b >= 'A' && b <= 'Z') || (b >= 'a' && b <= 'z') || (b >= '0' && b <= '9') || b == '_'
}
