# Proposal 3: Hybrid Three-Tier Documentation System

## Overview

This proposal combines the best of Proposals 1 and 2 into a **Three-Tier Documentation Architecture**:

1. **Learning Tier** (Tutorial/Guide) - How to get started and understand concepts
2. **Cookbook Tier** (Problem → Solution) - "I want to..." use cases
3. **Reference Tier** (Complete Specification) - Definitive property documentation

## Problem Statement

Different users need documentation for different purposes:

- **New users** need tutorials and conceptual understanding
- **Practitioners** need quick solutions to specific problems
- **Power users** need complete technical specifications

Current documentation tries to serve all these needs in two places (guide.md + reference/), resulting in:
- Guide that's too long and reference-like
- Reference that lacks practical context
- No clear path for different user types

## Proposed Solution

### Three-Tier Architecture

```
┌─────────────────────────────────────────────────────┐
│              TIER 1: LEARNING                       │
│  Tutorial-focused, conceptual, narrative            │
│  → getting-started.md, guide.md                     │
│     Target: New users, understanding concepts       │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│              TIER 2: COOKBOOK                       │
│  Task-focused, practical, copy-paste ready          │
│  → cookbook/* (30-40 recipes)                       │
│     Target: Daily users, solving specific problems  │
└─────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────┐
│              TIER 3: REFERENCE                      │
│  Complete, exhaustive, technical                    │
│  → reference/* + properties/*                       │
│     Target: Advanced users, edge cases, all options │
└─────────────────────────────────────────────────────┘
```

### Directory Structure

```
website/src/docs/
├── index.md                      # Landing page with tier navigation
├── installation.md
│
├── learning/                     # TIER 1: Learning
│   ├── getting-started.md       # Quick 5-minute intro
│   ├── core-concepts.md         # Tasks, deps, vars explained
│   ├── your-first-taskfile.md   # Step-by-step tutorial
│   └── best-practices.md        # Recommended patterns
│
├── cookbook/                     # TIER 2: Cookbook (Proposal 2)
│   ├── index.md                 # Searchable recipe index
│   ├── data-sharing/
│   ├── task-orchestration/
│   ├── optimization/
│   ├── configuration/
│   ├── validation/
│   └── advanced-patterns/
│
├── reference/                    # TIER 3: Reference
│   ├── properties/              # Property-centric docs (Proposal 1)
│   │   ├── index.md            # All properties at a glance
│   │   ├── vars.md             # Complete vars reference
│   │   ├── cmds.md
│   │   ├── deps.md
│   │   ├── sources.md
│   │   ├── generates.md
│   │   └── ... (20+ properties)
│   ├── cli.md                   # CLI reference
│   ├── schema.md                # JSON schema (technical)
│   ├── templating.md            # Go template reference
│   └── environment.md           # Environment variables
│
└── ... (other docs: FAQ, community, contributing)
```

## User Journeys

### Journey 1: Complete Beginner

```
1. installation.md → Install Task
2. learning/getting-started.md → First Taskfile in 5 minutes
3. learning/your-first-taskfile.md → Build real project setup
4. cookbook/... → Solve specific needs as they arise
5. reference/properties/... → Deep dive when curious
```

### Journey 2: Experienced Developer (New to Task)

```
1. learning/core-concepts.md → Understand Task model
2. cookbook/index.md → Browse common patterns
3. cookbook/[relevant-recipe] → Apply to their project
4. reference/properties/... → Look up specific details
```

### Journey 3: Daily Task User

```
1. cookbook/index.md → Search for use case
2. cookbook/[recipe] → Copy-paste solution
3. reference/properties/[X].md → Understand property deeply if needed
```

### Journey 4: Power User / Contributor

```
1. reference/properties/index.md → Browse all properties
2. reference/properties/[X].md → Read complete specification
3. reference/schema.md → Check JSON schema for edge cases
```

## Tier Details

### Tier 1: Learning (Enhanced guide.md)

**Current**: guide.md (2,386 lines) tries to be tutorial + reference

**New**: Split into focused learning modules:

```markdown
# learning/core-concepts.md (300-400 lines)

## What is a Task?
Simple explanation with minimal example

## How Tasks Relate
Dependencies, calling, ordering

## Variables in Task
Introduction to vars, simple examples

## Running Tasks
Basic CLI usage

## Next Steps
→ Link to cookbook for specific use cases
→ Link to reference for complete details
```

**Benefits**:
- Digestible chunks
- Progressive learning
- Clear next steps
- Not overwhelming

### Tier 2: Cookbook (From Proposal 2)

**Purpose**: Solve specific problems quickly

**Structure**: Each recipe follows template:
- Problem statement
- Quick recipe table
- Multiple solutions (simple → complex)
- Common patterns
- Troubleshooting
- See also links

**Example Categories**:
1. Data Sharing (8 recipes)
2. Task Orchestration (6 recipes)
3. Optimization (5 recipes)
4. Configuration (6 recipes)
5. Validation (4 recipes)
6. Advanced Patterns (6 recipes)

Total: ~35 recipes

### Tier 3: Reference (From Proposal 1)

**Purpose**: Complete, exhaustive documentation

**Structure**: Properties directory with one file per property

**Coverage**:
- All contexts where property can be used
- All types and variations
- Complete syntax reference
- Priority/precedence rules
- All options and flags
- Edge cases and limitations

## Cross-Tier Linking Strategy

### From Learning → Cookbook
```markdown
# In learning/core-concepts.md
Want to learn more about passing data between tasks?
→ See [Cookbook: Data Sharing](../cookbook/data-sharing/)
```

### From Cookbook → Reference
```markdown
# In cookbook/data-sharing/between-tasks.md
For complete `vars` documentation including all types and contexts:
→ See [Reference: vars](../../reference/properties/vars.md)
```

### From Reference → Cookbook
```markdown
# In reference/properties/vars.md
## Common Patterns
For practical recipes using vars:
- [Passing data between tasks](../../cookbook/data-sharing/between-tasks.md)
- [Multi-environment configuration](../../cookbook/configuration/multi-environment.md)
```

### From Reference → Learning
```markdown
# In reference/properties/deps.md
New to task dependencies? Start with:
→ [Learning: Core Concepts](../../learning/core-concepts.md#task-dependencies)
```

## Navigation Design

### Main Navigation
```
Docs
├── 📚 Learning
│   ├── Getting Started
│   ├── Core Concepts
│   ├── Your First Taskfile
│   └── Best Practices
├── 👨‍🍳 Cookbook
│   ├── Browse Recipes
│   ├── Data Sharing
│   ├── Task Orchestration
│   ├── Optimization
│   ├── Configuration
│   └── Advanced
├── 📖 Reference
│   ├── Properties A-Z
│   ├── CLI
│   ├── Schema
│   └── Templating
└── Other
    ├── FAQ
    ├── Installation
    └── Community
```

### Documentation Landing Page

```markdown
# Task Documentation

Choose your path:

## 🎓 Just Starting?
→ [Getting Started](./learning/getting-started.md) - Your first Taskfile in 5 minutes
→ [Core Concepts](./learning/core-concepts.md) - Understand how Task works

## 🎯 Need to Solve Something?
→ [Browse Cookbook](./cookbook/) - 35+ recipes for common tasks
→ [Search Recipes](./cookbook/index.md) - Find solutions quickly

## 📚 Looking for Details?
→ [Properties Reference](./reference/properties/) - Complete property documentation
→ [CLI Reference](./reference/cli.md) - Command-line options
→ [Schema](./reference/schema.md) - Technical specification

## 💡 Popular Topics
- [Passing data between tasks](./cookbook/data-sharing/between-tasks.md)
- [Running tasks only when files change](./cookbook/optimization/incremental-builds.md)
- [Multi-environment configuration](./cookbook/configuration/multi-environment.md)
- [Complete vars reference](./reference/properties/vars.md)
```

## Implementation Plan

### Phase 1: Foundation (8 hours)
- Create three-tier directory structure
- Design templates for each tier
- Set up navigation in VitePress config
- Create documentation landing page

### Phase 2: Learning Tier (12 hours)
- Refactor guide.md into 4 focused modules
- Create core-concepts.md (comprehensive but digestible)
- Write your-first-taskfile.md tutorial
- Extract best-practices.md

### Phase 3: Cookbook Tier (32 hours)
- Implement 35 recipes across 6 categories
- Create cookbook index with search
- Add quick recipe tables
- Write troubleshooting sections

### Phase 4: Reference Tier (24 hours)
- Create 20 property reference docs
- Extract content from current guide.md and schema.md
- Unify into comprehensive per-property docs
- Add cross-references

### Phase 5: Cross-Linking (8 hours)
- Add tier-to-tier navigation
- Create "Next Steps" sections
- Implement "See Also" links
- Add breadcrumbs

### Phase 6: Testing & Polish (12 hours)
- User testing with each journey
- Fix navigation issues
- Ensure consistency
- Get community feedback

**Total**: ~96 hours (3 developers × 32 hours OR 1 developer × 12 weeks part-time)

## Benefits

### For New Users
- Clear learning path
- Not overwhelmed by reference material
- Can progress from tutorial → cookbook → reference

### For Daily Users
- Quick access to solutions (cookbook)
- Can skip learning tier if experienced
- Reference available when needed

### For Power Users
- Complete technical documentation
- All edge cases covered
- Deep understanding of all options

### For Maintainers
- Clear separation of concerns
- Easier to update (change in one tier doesn't affect others)
- Community can contribute recipes without touching reference

## Migration Strategy

### Week 1-2: Foundation
- Set up structure
- Create templates
- No content disruption

### Week 3-4: Learning Tier
- Refactor guide.md
- Add redirects from old guide sections
- Test with users

### Week 5-8: Cookbook Tier
- Create recipes incrementally
- Start with top 10 most-requested patterns
- Gather feedback, adjust template

### Week 9-12: Reference Tier
- Build property references
- Maintain backward compatibility with current reference/
- Add cross-links

### Week 13: Integration
- Complete cross-linking
- Final testing
- Launch announcement

## Success Metrics

### By Tier

**Learning Tier**:
- Time to first working Taskfile
- Tutorial completion rate
- User confidence (survey)

**Cookbook Tier**:
- Recipe usage (page views)
- Copy-paste rate
- Time to solution
- Community recipe contributions

**Reference Tier**:
- Search → find rate
- Time spent on property pages
- Reduced "where do I find X" questions

### Overall
- Documentation NPS (Net Promoter Score)
- GitHub issue reduction for docs questions
- Community feedback
- External blog posts / tutorials referencing new structure

## Risk Mitigation

### Risk: Too much work
**Mitigation**: Incremental rollout, community can contribute recipes

### Risk: User confusion with 3 tiers
**Mitigation**: Clear landing page, obvious user journey navigation

### Risk: Content duplication
**Mitigation**: Each tier has distinct purpose:
- Learning: Concepts + simple examples
- Cookbook: Solutions + practical patterns
- Reference: Complete specification + all options

### Risk: Maintenance burden
**Mitigation**: 
- Templates ensure consistency
- Clear ownership per tier
- Automated link checking

## Why This is Better Than Proposals 1 or 2 Alone

| Aspect | Proposal 1 Only | Proposal 2 Only | Proposal 3 (Hybrid) |
|--------|----------------|----------------|---------------------|
| New user onboarding | Lacks tutorial | Lacks tutorial | ✅ Learning tier |
| Quick solutions | Must search reference | ✅ Cookbook | ✅ Cookbook |
| Complete documentation | ✅ Properties | Missing details | ✅ Reference tier |
| Progressive disclosure | No | Limited | ✅ All tiers |
| User journey | Unclear | Task-focused only | ✅ Multiple paths |
| Maintainability | Good | Good | ✅ Excellent (separated) |

## Comparison Chart

```
User Type        | Best Tier      | Secondary     | Occasional
-----------------|----------------|---------------|-------------
Beginner         | Learning (1)   | Cookbook (2)  | -
Daily User       | Cookbook (2)   | Reference (3) | Learning (1)
Power User       | Reference (3)  | Cookbook (2)  | -
Contributor      | Reference (3)  | Learning (1)  | Cookbook (2)
Troubleshooter   | Cookbook (2)   | Reference (3) | -
```

## Conclusion

Proposal 3 provides a complete documentation system that serves all user types while maintaining clear separation of concerns. It combines:

- **Structured learning** for beginners
- **Practical solutions** for daily work
- **Complete reference** for deep understanding

This creates a sustainable, scalable documentation architecture that grows with the project and serves the community effectively.
