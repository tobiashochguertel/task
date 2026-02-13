package task

import (
	"testing"

	"github.com/Masterminds/semver/v3"

	"github.com/go-task/task/v3/taskfile/ast"
)

func TestTaskToMapBasicProperties(t *testing.T) {
	t.Parallel()
	silent := true
	task := &ast.Task{
		Task:        "build",
		Desc:        "Build the project",
		Summary:     "Builds everything",
		Aliases:     []string{"b", "compile"},
		Dir:         "./src",
		Method:      "checksum",
		Silent:      &silent,
		Interactive: false,
		Internal:    true,
		IgnoreError: true,
		Run:         "once",
		Watch:       true,
		Namespace:   "ns",
		Label:       "my-build",
		Prefix:      "[build]",
		If:          "true",
		Failfast:    true,
	}

	m := taskToMap(task)

	checks := map[string]any{
		"Name":        "build",
		"Desc":        "Build the project",
		"Summary":     "Builds everything",
		"Dir":         "./src",
		"Method":      "checksum",
		"Silent":      true,
		"Interactive": false,
		"Internal":    true,
		"IgnoreError": true,
		"Run":         "once",
		"Watch":       true,
		"Namespace":   "ns",
		"Label":       "my-build",
		"Prefix":      "[build]",
		"If":          "true",
		"Failfast":    true,
	}

	for key, want := range checks {
		got, ok := m[key]
		if !ok {
			t.Errorf("missing key %q in task map", key)
			continue
		}
		if got != want {
			t.Errorf("key %q = %v (%T), want %v (%T)", key, got, got, want, want)
		}
	}

	// Aliases
	aliases, ok := m["Aliases"].([]string)
	if !ok {
		t.Fatal("Aliases should be []string")
	}
	if len(aliases) != 2 || aliases[0] != "b" || aliases[1] != "compile" {
		t.Errorf("Aliases = %v, want [b compile]", aliases)
	}
}

func TestTaskToMapCmds(t *testing.T) {
	t.Parallel()
	task := &ast.Task{
		Task: "test",
		Cmds: []*ast.Cmd{
			{Cmd: "echo hello"},
			{Task: "subtask"},
			nil, // should be skipped
		},
	}

	m := taskToMap(task)
	cmds, ok := m["Cmds"].([]map[string]any)
	if !ok {
		t.Fatal("Cmds should be []map[string]any")
	}
	if len(cmds) != 2 {
		t.Fatalf("expected 2 cmds (nil skipped), got %d", len(cmds))
	}
	if cmds[0]["Cmd"] != "echo hello" {
		t.Errorf("cmd 0 Cmd = %q, want %q", cmds[0]["Cmd"], "echo hello")
	}
	if cmds[1]["Task"] != "subtask" {
		t.Errorf("cmd 1 Task = %q, want %q", cmds[1]["Task"], "subtask")
	}
}

func TestTaskToMapDeps(t *testing.T) {
	t.Parallel()
	task := &ast.Task{
		Task: "deploy",
		Deps: []*ast.Dep{
			{Task: "build"},
			{Task: "test", Silent: true},
			nil, // should be skipped
		},
	}

	m := taskToMap(task)
	deps, ok := m["Deps"].([]map[string]any)
	if !ok {
		t.Fatal("Deps should be []map[string]any")
	}
	if len(deps) != 2 {
		t.Fatalf("expected 2 deps (nil skipped), got %d", len(deps))
	}
	if deps[0]["Task"] != "build" {
		t.Errorf("dep 0 Task = %q, want build", deps[0]["Task"])
	}
	if deps[1]["Silent"] != true {
		t.Errorf("dep 1 Silent = %v, want true", deps[1]["Silent"])
	}
}

func TestTaskToMapSources(t *testing.T) {
	t.Parallel()
	task := &ast.Task{
		Task: "check",
		Sources: []*ast.Glob{
			{Glob: "*.go"},
			{Glob: "vendor/**", Negate: true},
		},
	}

	m := taskToMap(task)
	sources, ok := m["Sources"].([]string)
	if !ok {
		t.Fatal("Sources should be []string")
	}
	if len(sources) != 2 {
		t.Fatalf("expected 2 sources, got %d", len(sources))
	}
	if sources[0] != "*.go" {
		t.Errorf("source 0 = %q, want *.go", sources[0])
	}
	if sources[1] != "!vendor/**" {
		t.Errorf("source 1 = %q, want !vendor/**", sources[1])
	}
}

func TestTaskToMapVarsAndEnv(t *testing.T) {
	t.Parallel()
	vars := ast.NewVars()
	vars.Set("FOO", ast.Var{Value: "bar"})
	envVars := ast.NewVars()
	envVars.Set("PATH", ast.Var{Value: "/usr/bin"})

	task := &ast.Task{
		Task: "env-test",
		Vars: vars,
		Env:  envVars,
	}

	m := taskToMap(task)

	taskVars, ok := m["Vars"].(map[string]any)
	if !ok {
		t.Fatal("Vars should be map[string]any")
	}
	if taskVars["FOO"] != "bar" {
		t.Errorf("Vars[FOO] = %v, want bar", taskVars["FOO"])
	}

	taskEnv, ok := m["Env"].(map[string]any)
	if !ok {
		t.Fatal("Env should be map[string]any")
	}
	if taskEnv["PATH"] != "/usr/bin" {
		t.Errorf("Env[PATH] = %v, want /usr/bin", taskEnv["PATH"])
	}
}

func TestTaskToMapLocation(t *testing.T) {
	t.Parallel()
	task := &ast.Task{
		Task: "located",
		Location: &ast.Location{
			Line:     10,
			Column:   3,
			Taskfile: "Taskfile.yml",
		},
	}

	m := taskToMap(task)
	loc, ok := m["Location"].(map[string]any)
	if !ok {
		t.Fatal("Location should be map[string]any")
	}
	if loc["Line"] != 10 {
		t.Errorf("Location.Line = %v, want 10", loc["Line"])
	}
	if loc["Taskfile"] != "Taskfile.yml" {
		t.Errorf("Location.Taskfile = %v, want Taskfile.yml", loc["Taskfile"])
	}
}

func TestTaskToMapNilTask(t *testing.T) {
	t.Parallel()
	m := taskToMap(nil)
	if len(m) != 0 {
		t.Errorf("nil task should produce empty map, got %d keys", len(m))
	}
}

func TestTaskfileToMapBasicProperties(t *testing.T) {
	t.Parallel()
	v := semver.MustParse("3.0.0")
	tasks := ast.NewTasks()
	silent := true
	tasks.Set("build", &ast.Task{Task: "build", Desc: "Build", Silent: &silent})
	tasks.Set("test", &ast.Task{Task: "test", Desc: "Test"})

	vars := ast.NewVars()
	vars.Set("VERSION", ast.Var{Value: "1.0"})

	envVars := ast.NewVars()
	envVars.Set("GO111MODULE", ast.Var{Value: "on"})

	tf := &ast.Taskfile{
		Location: "/path/to/Taskfile.yml",
		Version:  v,
		Method:   "checksum",
		Silent:   false,
		Run:      "always",
		Tasks:    tasks,
		Vars:     vars,
		Env:      envVars,
		Dotenv:   []string{".env"},
	}

	m := taskfileToMap(tf)

	if m["Version"] != "3.0.0" {
		t.Errorf("Version = %v, want 3.0.0", m["Version"])
	}
	if m["Location"] != "/path/to/Taskfile.yml" {
		t.Errorf("Location = %v, want /path/to/Taskfile.yml", m["Location"])
	}
	if m["Method"] != "checksum" {
		t.Errorf("Method = %v, want checksum", m["Method"])
	}
	if m["Silent"] != false {
		t.Errorf("Silent = %v, want false", m["Silent"])
	}
	if m["Run"] != "always" {
		t.Errorf("Run = %v, want always", m["Run"])
	}

	// Dotenv
	dotenv, ok := m["Dotenv"].([]string)
	if !ok || len(dotenv) != 1 || dotenv[0] != ".env" {
		t.Errorf("Dotenv = %v, want [.env]", m["Dotenv"])
	}

	// Vars
	tfVars, ok := m["Vars"].(map[string]any)
	if !ok {
		t.Fatal("Vars should be map[string]any")
	}
	if tfVars["VERSION"] != "1.0" {
		t.Errorf("Vars[VERSION] = %v, want 1.0", tfVars["VERSION"])
	}

	// Env
	tfEnv, ok := m["Env"].(map[string]any)
	if !ok {
		t.Fatal("Env should be map[string]any")
	}
	if tfEnv["GO111MODULE"] != "on" {
		t.Errorf("Env[GO111MODULE] = %v, want on", tfEnv["GO111MODULE"])
	}
}

func TestTaskfileToMapTasks(t *testing.T) {
	t.Parallel()
	tasks := ast.NewTasks()
	tasks.Set("build", &ast.Task{Task: "build", Desc: "Build task"})
	tasks.Set("test", &ast.Task{Task: "test", Desc: "Test task"})
	tasks.Set("deploy", &ast.Task{Task: "deploy", Desc: "Deploy task"})

	tf := &ast.Taskfile{
		Version: semver.MustParse("3"),
		Tasks:   tasks,
		Vars:    ast.NewVars(),
		Env:     ast.NewVars(),
	}

	m := taskfileToMap(tf)

	// TaskNames should be a list of all task names
	taskNames, ok := m["TaskNames"].([]string)
	if !ok {
		t.Fatal("TaskNames should be []string")
	}
	if len(taskNames) != 3 {
		t.Fatalf("expected 3 task names, got %d", len(taskNames))
	}

	// Tasks should be a map of name→task info
	taskMap, ok := m["Tasks"].(map[string]any)
	if !ok {
		t.Fatal("Tasks should be map[string]any")
	}
	if len(taskMap) != 3 {
		t.Fatalf("expected 3 tasks in map, got %d", len(taskMap))
	}

	buildInfo, ok := taskMap["build"].(map[string]any)
	if !ok {
		t.Fatal("Tasks[build] should be map[string]any")
	}
	if buildInfo["Desc"] != "Build task" {
		t.Errorf("Tasks[build].Desc = %v, want Build task", buildInfo["Desc"])
	}
}

func TestTaskfileToMapNilTaskfile(t *testing.T) {
	t.Parallel()
	m := taskfileToMap(nil)
	if len(m) != 0 {
		t.Errorf("nil taskfile should produce empty map, got %d keys", len(m))
	}
}

func TestTaskfileToMapNilVersion(t *testing.T) {
	t.Parallel()
	tf := &ast.Taskfile{
		Tasks: ast.NewTasks(),
		Vars:  ast.NewVars(),
		Env:   ast.NewVars(),
	}
	m := taskfileToMap(tf)
	if m["Version"] != "" {
		t.Errorf("nil version should produce empty string, got %v", m["Version"])
	}
}

func TestGlobsToStrings(t *testing.T) {
	t.Parallel()
	globs := []*ast.Glob{
		{Glob: "*.go"},
		{Glob: "vendor/**", Negate: true},
		nil,
	}
	result := globsToStrings(globs)
	if len(result) != 2 {
		t.Fatalf("expected 2 results (nil skipped), got %d", len(result))
	}
	if result[0] != "*.go" {
		t.Errorf("result[0] = %q, want *.go", result[0])
	}
	if result[1] != "!vendor/**" {
		t.Errorf("result[1] = %q, want !vendor/**", result[1])
	}
}

func TestSafeStringSlice(t *testing.T) {
	t.Parallel()
	// nil should return empty slice, not nil
	result := safeStringSlice(nil)
	if result == nil {
		t.Error("safeStringSlice(nil) should return non-nil empty slice")
	}
	if len(result) != 0 {
		t.Errorf("safeStringSlice(nil) length = %d, want 0", len(result))
	}

	// non-nil should return a clone
	input := []string{"a", "b"}
	result = safeStringSlice(input)
	if len(result) != 2 || result[0] != "a" || result[1] != "b" {
		t.Errorf("safeStringSlice should clone input, got %v", result)
	}
}

func TestVarsToMap(t *testing.T) {
	t.Parallel()
	vars := ast.NewVars()
	vars.Set("STATIC", ast.Var{Value: "hello"})
	sh := "echo world"
	vars.Set("DYNAMIC", ast.Var{Sh: &sh})

	m := varsToMap(vars)
	if m["STATIC"] != "hello" {
		t.Errorf("STATIC = %v, want hello", m["STATIC"])
	}
	if m["DYNAMIC"] != "(sh: echo world)" {
		t.Errorf("DYNAMIC = %v, want (sh: echo world)", m["DYNAMIC"])
	}
}

func TestVarsToMapNil(t *testing.T) {
	t.Parallel()
	m := varsToMap(nil)
	if m == nil || len(m) != 0 {
		t.Errorf("varsToMap(nil) should return empty non-nil map, got %v", m)
	}
}

func TestCmdToMap(t *testing.T) {
	t.Parallel()
	cmd := &ast.Cmd{
		Cmd:         "go build",
		Task:        "",
		If:          "true",
		Silent:      true,
		IgnoreError: false,
		Defer:       false,
	}
	m := cmdToMap(cmd)
	if m["Cmd"] != "go build" {
		t.Errorf("Cmd = %v, want go build", m["Cmd"])
	}
	if m["Silent"] != true {
		t.Errorf("Silent = %v, want true", m["Silent"])
	}
}

func TestCmdToMapNil(t *testing.T) {
	t.Parallel()
	m := cmdToMap(nil)
	if len(m) != 0 {
		t.Errorf("nil cmd should produce empty map, got %d keys", len(m))
	}
}

func TestDepToMap(t *testing.T) {
	t.Parallel()
	dep := &ast.Dep{Task: "build", Silent: true}
	m := depToMap(dep)
	if m["Task"] != "build" {
		t.Errorf("Task = %v, want build", m["Task"])
	}
	if m["Silent"] != true {
		t.Errorf("Silent = %v, want true", m["Silent"])
	}
}

func TestDepToMapNil(t *testing.T) {
	t.Parallel()
	m := depToMap(nil)
	if len(m) != 0 {
		t.Errorf("nil dep should produce empty map, got %d keys", len(m))
	}
}
