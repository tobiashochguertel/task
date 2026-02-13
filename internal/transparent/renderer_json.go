package transparent

import (
	"encoding/json"
	"fmt"
	"io"
)

// jsonReport mirrors TraceReport with JSON-friendly struct tags.
type jsonReport struct {
	Version           string          `json:"version"`
	WhitespaceVisible bool            `json:"whitespace_visible,omitempty"`
	GlobalVars        []jsonVarTrace  `json:"global_vars,omitempty"`
	Tasks             []jsonTaskTrace `json:"tasks"`
}

type jsonTaskTrace struct {
	Name         string              `json:"name"`
	Variables    []jsonVarTrace      `json:"variables"`
	Templates    []jsonTemplateTrace `json:"templates,omitempty"`
	Commands     []jsonCmdTrace      `json:"commands,omitempty"`
	Dependencies []string            `json:"dependencies,omitempty"`
}

type jsonVarTrace struct {
	Name      string          `json:"name"`
	Origin    string          `json:"origin"`
	Type      string          `json:"type"`
	Value     any             `json:"value"`
	ValueID   string          `json:"value_id,omitempty"`
	IsRef     bool            `json:"is_ref,omitempty"`
	RefName   string          `json:"ref_name,omitempty"`
	IsDynamic bool            `json:"is_dynamic,omitempty"`
	ShCmd     string          `json:"sh_cmd,omitempty"`
	Shadows   *jsonShadowInfo `json:"shadows,omitempty"`
	Warning   string          `json:"warning,omitempty"`
}

type jsonShadowInfo struct {
	Name   string `json:"name"`
	Value  any    `json:"value"`
	Origin string `json:"origin"`
}

type jsonTemplateTrace struct {
	Input       string               `json:"input"`
	Output      string               `json:"output"`
	Context     string               `json:"context,omitempty"`
	VarsUsed    []string             `json:"vars_used,omitempty"`
	Steps       []jsonPipeStep       `json:"pipe_steps,omitempty"`
	EvalActions []jsonEvalAction     `json:"eval_actions,omitempty"`
	Diagnostics []jsonFuncDiagnostic `json:"diagnostics,omitempty"`
	Tips        []string             `json:"tips,omitempty"`
	Notes       []string             `json:"notes,omitempty"`
	Error       string               `json:"error,omitempty"`
}

type jsonFuncDiagnostic struct {
	DiagType   string             `json:"diag_type"`
	FuncName   string             `json:"func_name"`
	StepNum    int                `json:"step_num"`
	Expression string             `json:"expression,omitempty"`
	Signature  string             `json:"signature,omitempty"`
	Example    string             `json:"example,omitempty"`
	Call       string             `json:"call,omitempty"`
	Params     []jsonParamMapping `json:"params,omitempty"`
	ErrorMsg   string             `json:"error_msg"`
	Output     string             `json:"output,omitempty"`
}

type jsonParamMapping struct {
	Name     string `json:"name"`
	Type     string `json:"type"`
	Value    string `json:"value,omitempty"`
	Variadic bool   `json:"variadic,omitempty"`
	Missing  bool   `json:"missing,omitempty"`
}

type jsonEvalAction struct {
	ActionIndex int                `json:"action_index"`
	SourceLine  int                `json:"source_line"`
	Source      string             `json:"source"`
	Result      string             `json:"result"`
	Steps       []jsonTemplateStep `json:"steps"`
}

type jsonTemplateStep struct {
	StepNum   int    `json:"step"`
	Operation string `json:"operation"`
	Target    string `json:"target"`
	Input     string `json:"input,omitempty"`
	Output    string `json:"output,omitempty"`
}

type jsonPipeStep struct {
	FuncName   string   `json:"func"`
	Args       []string `json:"args,omitempty"`
	ArgsValues []string `json:"args_values,omitempty"`
	Output     string   `json:"output"`
}

type jsonCmdTrace struct {
	Index          int    `json:"index"`
	RawCmd         string `json:"raw"`
	ResolvedCmd    string `json:"resolved"`
	IterationLabel string `json:"iteration,omitempty"`
}

// RenderJSON writes the trace report as JSON to the given writer.
// Returns nil for a nil or empty report.
func RenderJSON(w io.Writer, report *TraceReport, opts *RenderOptions) error {
	if report == nil {
		_, err := w.Write([]byte("{\"version\":\"1.0\",\"tasks\":[]}\n"))
		return err
	}
	if opts == nil {
		opts = &RenderOptions{}
	}

	// ShowWhitespaces: apply transformation to report values
	if opts.ShowWhitespaces {
		report = applyWhitespaceVisibility(report)
	}

	jr := jsonReport{
		Version:           "1.0",
		WhitespaceVisible: opts.ShowWhitespaces,
		Tasks:             make([]jsonTaskTrace, 0, len(report.Tasks)),
	}

	// Global variables (apply verbose filter)
	globals := filterGlobals(report.GlobalVars, opts.Verbose)
	for _, v := range globals {
		jr.GlobalVars = append(jr.GlobalVars, varToJSON(v))
	}

	for _, task := range report.Tasks {
		jt := jsonTaskTrace{
			Name:         task.TaskName,
			Variables:    make([]jsonVarTrace, 0, len(task.Vars)),
			Dependencies: task.Deps,
		}

		for _, v := range task.Vars {
			jt.Variables = append(jt.Variables, varToJSON(v))
		}

		for _, tmpl := range task.Templates {
			jtt := jsonTemplateTrace{
				Input:    tmpl.Input,
				Output:   tmpl.Output,
				Context:  tmpl.Context,
				VarsUsed: tmpl.VarsUsed,
				Tips:     tmpl.Tips,
				Error:    tmpl.Error,
			}
			for _, step := range tmpl.Steps {
				jtt.Steps = append(jtt.Steps, jsonPipeStep(step))
			}
			for _, ea := range tmpl.EvalActions {
				jea := jsonEvalAction{
					ActionIndex: ea.ActionIndex,
					SourceLine:  ea.SourceLine,
					Source:      ea.Source,
					Result:      ea.Result,
				}
				for _, ds := range ea.Steps {
					jea.Steps = append(jea.Steps, jsonTemplateStep(ds))
				}
				jtt.EvalActions = append(jtt.EvalActions, jea)
			}
			for _, d := range tmpl.Diagnostics {
				jd := jsonFuncDiagnostic{
					DiagType:   d.DiagType,
					FuncName:   d.FuncName,
					StepNum:    d.StepNum,
					Expression: d.Expression,
					Signature:  d.Signature,
					Example:    d.Example,
					Call:       d.Call,
					ErrorMsg:   d.ErrorMsg,
					Output:     d.Output,
				}
				for _, p := range d.Params {
					jd.Params = append(jd.Params, jsonParamMapping(p))
				}
				jtt.Diagnostics = append(jtt.Diagnostics, jd)
			}
			jt.Templates = append(jt.Templates, jtt)
		}

		for _, cmd := range task.Cmds {
			jt.Commands = append(jt.Commands, jsonCmdTrace{
				Index:          cmd.Index,
				RawCmd:         cmd.RawCmd,
				ResolvedCmd:    cmd.ResolvedCmd,
				IterationLabel: cmd.IterationLabel,
			})
		}

		jr.Tasks = append(jr.Tasks, jt)
	}

	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(jr)
}

// varToJSON converts a VarTrace to its JSON representation.
func varToJSON(v VarTrace) jsonVarTrace {
	jv := jsonVarTrace{
		Name:      v.Name,
		Origin:    v.Origin.String(),
		Type:      v.Type,
		Value:     v.Value,
		IsRef:     v.IsRef,
		RefName:   v.RefName,
		IsDynamic: v.IsDynamic,
		ShCmd:     v.ShCmd,
	}
	if v.ValueID != 0 {
		jv.ValueID = fmt.Sprintf("0x%x", v.ValueID)
	}
	if v.ShadowsVar != nil {
		jv.Shadows = &jsonShadowInfo{
			Name:   v.ShadowsVar.Name,
			Value:  v.ShadowsVar.Value,
			Origin: v.ShadowsVar.Origin.String(),
		}
	}
	if v.IsDynamic && fmt.Sprintf("%v", v.Value) == "" {
		jv.Warning = "dynamic variable not evaluated (sh: command not executed in transparent mode)"
	}
	return jv
}
