# 02 — Architecture Overview

## High-Level Design

Transparent Mode is a **compile-time-only diagnostic layer** that intercepts the variable resolution and template rendering pipeline, collects trace data, and renders a structured report. It does **not** execute any commands.

## Architecture Diagram

```mermaid
graph TD
    CLI["CLI: --transparent flag"]
    FLAGS["internal/flags/flags.go"]
    EXEC["Executor (executor.go)"]
    SETUP["Setup (setup.go)"]
    COMP["Compiler (compiler.go)"]
    TMPL["Templater (internal/templater/)"]
    
    TRACER["NEW: internal/transparent/tracer.go"]
    RENDERER["NEW: internal/transparent/renderer.go"]
    REPORT["Diagnostic Report (stderr)"]
    
    CLI --> FLAGS
    FLAGS --> EXEC
    EXEC --> SETUP
    SETUP --> COMP
    COMP --> TMPL
    
    COMP -.->|"hook: OnVarResolved()"| TRACER
    TMPL -.->|"hook: OnTemplateEval()"| TRACER
    
    TRACER --> RENDERER
    RENDERER --> REPORT

    style TRACER fill:#2d5,stroke:#333,color:#000
    style RENDERER fill:#2d5,stroke:#333,color:#000
    style REPORT fill:#2d5,stroke:#333,color:#000
```

## Component Diagram

```mermaid
graph LR
    subgraph "Existing Code (minimal changes)"
        A[flags.go] -->|"Transparent bool"| B[executor.go]
        B -->|"e.Transparent"| C[compiler.go]
        C -->|"tracer param"| D[templater.go]
    end

    subgraph "New Package: internal/transparent"
        E[tracer.go] -->|"collects"| F[model.go]
        F -->|"consumed by"| G[renderer.go]
    end

    C -.->|"VarTrace events"| E
    D -.->|"TemplateTrace events"| E
    G -->|"formatted output"| H[stderr]

    style E fill:#2d5,stroke:#333,color:#000
    style F fill:#2d5,stroke:#333,color:#000
    style G fill:#2d5,stroke:#333,color:#000
```

## Data Flow

```mermaid
sequenceDiagram
    participant CLI
    participant Executor
    participant Compiler
    participant Templater
    participant Tracer
    participant Renderer

    CLI->>Executor: --transparent flag
    Executor->>Executor: Setup() (normal)
    Executor->>Compiler: getVariables(task, call)
    
    loop For each variable scope
        Compiler->>Tracer: RecordVar(name, value, origin, scope)
    end

    Compiler->>Templater: Replace(template, cache)
    
    loop For each template string
        Templater->>Tracer: RecordTemplate(input, output, funcCalls)
    end

    Executor->>Renderer: Render(tracer.Traces())
    Renderer->>CLI: Print diagnostic report
    Note over CLI: Exit (no task execution)
```

## Design Principles

| Principle | How Applied |
|-----------|-------------|
| **S** — Single Responsibility | `Tracer` only collects; `Renderer` only formats; existing code only emits events |
| **O** — Open/Closed | New package `internal/transparent/` — no modification of core data structures |
| **L** — Liskov Substitution | Tracer is nil-safe (no-op when transparent mode off) |
| **I** — Interface Segregation | Small `Trace` interface, not a monolithic debugger |
| **D** — Dependency Inversion | Compiler/Templater depend on a `TraceCollector` interface, not the concrete tracer |

## Key Decision: Nil-Safe Tracer Pattern

The tracer is injected into `Compiler` and `templater.Cache` as an **optional pointer**. All methods on the tracer are nil-receiver safe:

```go
// All methods are no-ops when t is nil
func (t *Tracer) RecordVar(name string, v VarTrace) {
    if t == nil { return }
    // ...
}
```

This means **zero performance impact** when transparent mode is off — no interface dispatch, no allocation, just a nil pointer check.
