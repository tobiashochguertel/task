# Proposal 2: Context-Based Cookbook Architecture

## Overview

This proposal suggests creating a **Context-Based Cookbook** that organizes documentation around developer intent and use cases ("I want to...") rather than properties or technical specifications.

## Problem Statement

Developers often know WHAT they want to achieve but don't know WHICH properties or combinations to use:

- "I want to pass data between tasks" → Need vars? deps? environment variables?
- "I want to run tasks only when files change" → Need sources? generates? status? method?
- "I want to share configuration across multiple Taskfiles" → Need includes? vars? dotenv?

Current documentation is organized by WHAT EXISTS (properties, features) not by WHAT USERS WANT TO DO (goals, outcomes).

## Proposed Solution

### New Directory Structure

```
website/src/docs/
├── guide.md (keep as is)
├── getting-started.md
├── cookbook/                    # NEW: Use-case driven docs
│   ├── index.md                # Searchable cookbook index
│   ├── data-sharing/
│   │   ├── index.md            # Overview of data sharing
│   │   ├── between-tasks.md    # Pass data between tasks
│   │   ├── from-cli.md         # Accept CLI arguments
│   │   ├── from-environment.md # Use environment variables
│   │   ├── from-files.md       # Load from .env, JSON, etc.
│   │   └── across-taskfiles.md # Share data in includes
│   ├── task-orchestration/
│   │   ├── index.md
│   │   ├── sequential.md       # Run tasks in order
│   │   ├── parallel.md         # Run tasks concurrently
│   │   ├── conditional.md      # Run based on conditions
│   │   ├── loops.md            # Iterate over items
│   │   └── dependencies.md     # Manage task deps
│   ├── optimization/
│   │   ├── index.md
│   │   ├── incremental-builds.md    # sources/generates
│   │   ├── caching.md               # Avoid redundant work
│   │   ├── parallel-execution.md    # Speed up builds
│   │   └── watch-mode.md            # Continuous development
│   ├── configuration/
│   │   ├── index.md
│   │   ├── multi-environment.md     # Dev/staging/prod
│   │   ├── secrets.md               # Handle sensitive data
│   │   ├── monorepos.md             # Workspace organization
│   │   └── shared-configs.md        # DRY principles
│   ├── validation/
│   │   ├── index.md
│   │   ├── required-inputs.md       # Ensure vars are set
│   │   ├── preconditions.md         # Check before running
│   │   ├── status-checks.md         # Up-to-date validation
│   │   └── error-handling.md        # Graceful failures
│   └── advanced-patterns/
│       ├── index.md
│       ├── dynamic-tasks.md         # Generate tasks
│       ├── matrix-builds.md         # Build matrices
│       ├── pipeline-stages.md       # CI/CD patterns
│       └── custom-output.md         # Format output
├── reference/                   # Keep technical reference
│   ├── cli.md
│   ├── schema.md
│   ├── templating.md
│   └── properties.md            # Quick property lookup
└── ... (other existing files)
```

### Example: cookbook/data-sharing/between-tasks.md

```markdown
# Passing Data Between Tasks

Learn how to pass data from one task to another in various scenarios.

## Quick Recipes

| Scenario | Solution | Jump to |
|----------|----------|---------|
| Simple value | Use `vars:` when calling | [→](#simple-values) |
| Complex data | Use `ref:` for type safety | [→](#complex-data-types) |
| Multiple tasks | Set task-level vars | [→](#reusable-values) |
| Computed values | Use `sh:` dynamic vars | [→](#computed-values) |
| External source | Read from files | [→](#from-files) |

## Problem: I Need to Pass Values

### Simple Values

**Use Case**: Pass a single value like a version number or environment name.

**Solution**: Use `vars:` when calling the task.

```yaml
version: '3'

tasks:
  build:
    cmds:
      - task: compile
        vars:
          VERSION: 1.2.3
      - task: package
        vars:
          VERSION: 1.2.3

  compile:
    cmds:
      - go build -ldflags="-X main.Version={{.VERSION}}"

  package:
    cmds:
      - tar -czf app-{{.VERSION}}.tar.gz bin/
```

**Why this works**: Variables passed via `vars:` are available in the called task.

**Gotcha**: String values only with this method. For complex types, see below.

---

### Complex Data Types

**Use Case**: Pass arrays, maps, or other structured data without string conversion.

**Solution**: Use `ref:` to preserve the data type.

```yaml
version: '3'

tasks:
  deploy-all:
    vars:
      SERVICES:
        ref: SERVICES  # Preserves array type
    cmds:
      - task: deploy-each
        vars:
          SERVICES:
            ref: SERVICES

  deploy-each:
    for: {var: SERVICES}  # Loops require array type
    cmd: kubectl apply -f {{.ITEM}}.yaml

vars:
  SERVICES: [api, worker, frontend]
```

**Why this works**: `ref:` doesn't convert to string, preserving the original type for operations like loops.

**When to use**: Anytime you need to loop, index, or manipulate structured data.

---

### Reusable Values

**Use Case**: Multiple tasks need the same computed or dynamic value.

**Solution**: Define vars at the task level or globally.

```yaml
version: '3'

vars:
  # Computed once, used everywhere
  GIT_COMMIT:
    sh: git rev-parse --short HEAD
  BUILD_TIME:
    sh: date -u +"%Y-%m-%dT%H:%M:%SZ"

tasks:
  build:
    cmds:
      - echo "Building commit {{.GIT_COMMIT}} at {{.BUILD_TIME}}"
      - task: compile
      - task: test

  compile:
    cmds:
      - go build -ldflags="-X main.Commit={{.GIT_COMMIT}}"

  test:
    cmds:
      - echo "Testing {{.GIT_COMMIT}}"
```

**Why this works**: Global vars are available to all tasks and computed once per task run.

---

### Computed Values

**Use Case**: Generate a value based on command output for use in other tasks.

**Solution**: Use dynamic variables with `sh:`.

```yaml
version: '3'

tasks:
  deploy:
    vars:
      # Compute at runtime
      IMAGE_TAG:
        sh: git describe --tags --always
      REGISTRY:
        sh: aws ecr describe-repositories --repository-names myapp --query 'repositories[0].repositoryUri' --output text
    cmds:
      - task: push-image
        vars:
          IMAGE: "{{.REGISTRY}}:{{.IMAGE_TAG}}"

  push-image:
    cmds:
      - docker push {{.IMAGE}}
```

**Why this works**: `sh:` executes commands and assigns output to variables.

**Performance**: Dynamic vars are cached per task execution.

---

### From Files

**Use Case**: Read configuration from JSON, YAML, or environment files.

**Solution 1 - Environment files**: Use `dotenv:`

```yaml
version: '3'

dotenv:
  - .env
  - .env.{{.ENVIRONMENT}}

tasks:
  deploy:
    cmds:
      - echo "Deploying to {{.API_URL}}"
```

**.env.production**:
```
API_URL=https://api.prod.example.com
DB_HOST=prod-db.example.com
```

**Solution 2 - JSON/YAML**: Parse with template functions

```yaml
version: '3'

tasks:
  deploy:
    vars:
      CONFIG:
        sh: cat config.json
      API_URL:
        sh: cat config.json | jq -r '.api.url'
    cmds:
      - curl {{.API_URL}}/deploy
```

---

## Common Patterns

### Pattern: Build Pipeline with Artifacts

Pass build artifacts location between stages:

```yaml
version: '3'

vars:
  BUILD_DIR: ./dist
  ARTIFACT_NAME: app-{{.VERSION}}.tar.gz

tasks:
  pipeline:
    cmds:
      - task: build
      - task: test
      - task: package
      - task: upload

  build:
    cmds:
      - mkdir -p {{.BUILD_DIR}}
      - go build -o {{.BUILD_DIR}}/app

  package:
    cmds:
      - tar -czf {{.BUILD_DIR}}/{{.ARTIFACT_NAME}} -C {{.BUILD_DIR}} app

  upload:
    cmds:
      - aws s3 cp {{.BUILD_DIR}}/{{.ARTIFACT_NAME}} s3://artifacts/
```

### Pattern: Environment-Specific Configuration

Pass environment context through task chain:

```yaml
version: '3'

tasks:
  deploy:
    vars:
      ENV: '{{.ENV | default "dev"}}'
    cmds:
      - task: validate-env
        vars: {ENV: "{{.ENV}}"}
      - task: build
        vars: {ENV: "{{.ENV}}"}
      - task: push
        vars: {ENV: "{{.ENV}}"}

  validate-env:
    requires:
      vars:
        - name: ENV
          enum: [dev, staging, production]
    cmds:
      - echo "Deploying to {{.ENV}}"
```

## Troubleshooting

### Problem: Variable Not Available

**Symptom**: `{{.VAR}}` expands to empty string

**Causes**:
1. Variable not defined in scope
2. Variable defined in different task
3. Typo in variable name

**Solution**: Check variable priority order:
1. CLI args (highest)
2. Call-time vars
3. Task vars
4. Global vars
5. Environment vars (lowest)

### Problem: Type Mismatch in Loop

**Symptom**: `for` loop doesn't work, type error

**Cause**: Variable was string-templated instead of referenced

**Solution**: Use `ref:` instead of template:

```yaml
# ❌ Wrong - converts to string
vars:
  MY_ITEMS: "{{.ITEMS}}"

# ✅ Correct - preserves array type
vars:
  MY_ITEMS:
    ref: ITEMS
```

## See Also

- [Passing Data Across Taskfiles](./across-taskfiles.md)
- [Accepting CLI Arguments](./from-cli.md)
- [Loading from Environment](./from-environment.md)
- [Schema Reference: vars](../../reference/schema.md#vars)
- [Template Reference](../../reference/templating.md)
```

## Benefits

### 1. Task-Oriented
Documentation organized by what users want to accomplish, not by technical features.

### 2. Immediate Solutions
Quick recipes at the top of each page provide instant answers.

### 3. Context-Rich Examples
Real-world scenarios show not just HOW but WHEN and WHY to use features.

### 4. Progressive Disclosure
Start with simple examples, expand to complex patterns.

### 5. Troubleshooting Included
Common problems and solutions in context.

### 6. Cross-Cutting Concerns
Shows how multiple properties work together to solve real problems.

## Implementation Plan

### Phase 1: Structure (4 hours)
- Create cookbook directory structure
- Design recipe template
- Plan 30-40 cookbook entries

### Phase 2: Core Recipes (24 hours)
Write 20 most common use cases:
- Data sharing (6 recipes)
- Task orchestration (6 recipes)
- Optimization (4 recipes)
- Configuration (4 recipes)

### Phase 3: Advanced Patterns (12 hours)
10 advanced recipes for power users

### Phase 4: Index & Navigation (4 hours)
- Create searchable cookbook index
- Add "Related Recipes" links
- Update main navigation

### Phase 5: Integration (4 hours)
- Link from guide.md to recipes
- Add recipe callouts in reference docs

**Total**: ~48 hours

## Migration Strategy

### Complement, Don't Replace
- Keep guide.md for learning flow
- Keep reference/ for technical specs
- Add cookbook/ for practical solutions

### User Pathways
1. **New users**: getting-started.md → guide.md
2. **Task-focused**: cookbook/ → specific recipe
3. **Reference lookup**: reference/ → schema/properties

### Content Sources
- Extract patterns from guide.md examples
- Create recipes from common GitHub issues
- Add community-contributed patterns

## Success Metrics

1. **Time to solution** - How fast users find working code
2. **Copy-paste rate** - How often users copy examples
3. **Return visits** - Users coming back to cookbook
4. **Issue reduction** - Fewer "how do I..." questions
5. **Recipe contributions** - Community adds patterns

## Comparison with Proposal 1

| Aspect | Proposal 1 (Properties) | Proposal 2 (Cookbook) |
|--------|------------------------|---------------------|
| Organization | By property | By use case |
| Best for | Reference lookup | Problem solving |
| Examples | Comprehensive per property | Task-focused combinations |
| Learning curve | Requires knowing properties | Starts with intent |
| Maintenance | Per-property updates | Per-recipe updates |
| Completeness | All property options | Common scenarios |

**Recommendation**: These proposals are complementary! Proposal 2 helps users solve problems, Proposal 1 helps understand individual properties deeply.

## Example Use Cases

### "I want to run tests only if source code changed"
→ cookbook/optimization/incremental-builds.md

### "I want to deploy to different environments"
→ cookbook/configuration/multi-environment.md

### "I want to pass JSON data between tasks"
→ cookbook/data-sharing/between-tasks.md (Complex Data Types section)

### "I want to avoid running the same task twice"
→ cookbook/optimization/caching.md

## Risk Mitigation

### Risk: Recipe duplication
**Mitigation**: Each recipe focuses on one scenario, links to related recipes

### Risk: Outdated examples
**Mitigation**: Automated testing of code examples (future enhancement)

### Risk: Overwhelming navigation
**Mitigation**: Clear categorization, powerful search/index

### Risk: Gaps in coverage
**Mitigation**: Start with top 20 use cases, expand based on feedback
