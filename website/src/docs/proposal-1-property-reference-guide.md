# Proposal 1: Property-Centric Reference Guide

## Overview

This proposal suggests creating a **Property-Centric Reference Architecture** that reorganizes documentation around individual Taskfile properties (like `vars`, `cmds`, `deps`, etc.) rather than organizing by topics or guides.

## Problem Statement

Currently, when developers want to understand a specific property like `vars:`, they must:

1. Search through the 2,386-line guide.md
2. Check the schema reference for technical specs
3. Look through examples scattered across different sections
4. Remember where different use cases are documented

For example, `vars:` can be defined in:
- Global/root level (`vars:` at Taskfile root)
- Task level (`tasks.mytask.vars:`)
- Include level (`includes.namespace.vars:`)
- When calling tasks (`task: sometask, vars: {}`)
- In dependencies (`deps[].vars:`)

This information is currently spread across:
- guide.md lines 1105-1230 (Variables section)
- guide.md line 431 (Vars of included Taskfiles)
- reference/schema.md lines 83-110 (root vars)
- reference/schema.md lines 302-313 (include vars)
- Multiple examples throughout both files

## Proposed Solution

### New Directory Structure

```
website/src/docs/
├── guide.md (simplified, narrative-focused)
├── getting-started.md
├── installation.md
├── properties/           # NEW: Property reference directory
│   ├── index.md         # Overview of all properties
│   ├── vars.md          # Complete vars reference
│   ├── cmds.md          # Complete cmds reference
│   ├── deps.md          # Complete deps reference
│   ├── sources.md
│   ├── generates.md
│   ├── preconditions.md
│   ├── status.md
│   ├── includes.md
│   ├── env.md
│   ├── dotenv.md
│   ├── run.md
│   ├── method.md
│   ├── output.md
│   └── ... (one file per major property)
├── reference/
│   ├── cli.md
│   ├── schema.md        # Technical JSON schema only
│   ├── templating.md
│   └── environment.md
└── ... (other existing files)
```

### Example: properties/vars.md Structure

Each property file follows a consistent template:

1. **Quick Reference Table** - Shows all contexts where property can be used
2. **Overview** - What the property does
3. **Where Can You Define It?** - All locations with examples
4. **Property Types/Options** - Different ways to use it
5. **Priority/Precedence** - When multiple definitions exist
6. **Common Patterns** - Real-world use cases
7. **Validation** - Requirements and constraints
8. **See Also** - Related properties and references

## Benefits

### 1. Findability
Developers can directly navigate to `properties/vars.md` when they need information about vars. No more searching through 2000+ lines of guide.

### 2. Completeness  
All information about a property in ONE place:
- All contexts where it can be used
- All types and variations
- Priority rules
- Validation options
- Real examples

### 3. Consistent Structure
Every property document follows the same pattern, making it predictable and easy to scan.

### 4. Context Awareness
Clearly shows WHERE and HOW a property can be used, answering the "can I use this here?" question immediately.

### 5. Cross-referencing
Easy to link between related properties (e.g., vars ↔ env ↔ dotenv ↔ cmds).

### 6. Maintainability
Updates to a property go in one file instead of scattered across multiple documents.

## Implementation Plan

### Phase 1: Setup (2 hours)
- Create `website/src/docs/properties/` directory
- Create index.md template
- Set up navigation in .vitepress/config.ts

### Phase 2: Core Properties (16 hours)
Extract and reorganize content for top 8 most-used properties:
- vars (shown in example above)
- cmds
- deps  
- sources
- generates
- preconditions
- status
- includes

### Phase 3: Index Page (2 hours)
Create comprehensive index with:
- Searchable table of all properties
- Quick links
- Property categories

### Phase 4: Cross-referencing (4 hours)
- Add links from guide.md to property docs
- Update schema.md to reference property docs
- Add "See Also" sections

### Phase 5: Remaining Properties (8 hours)
Fill in remaining 12-15 properties

### Phase 6: Testing & Polish (4 hours)
- Review for completeness
- Test all links
- Get feedback

**Total**: ~36 hours

## Migration Strategy

### Backward Compatibility
- Keep guide.md as narrative/tutorial focused
- Keep reference/schema.md for technical JSON schema
- No breaking changes to existing URLs
- Add redirects if URLs need to change

### Content Transformation
1. Extract property information from guide.md
2. Extract schema details from reference/schema.md
3. Combine into unified property documents
4. Add cross-references and examples
5. Update original docs with links to new structure

### User Journey
**Before**: User searches → finds guide → searches within 2386 lines → finds section → might miss other contexts

**After**: User searches OR navigates to Properties → finds vars.md → sees ALL contexts, types, examples in one place

## Example Quick Reference (from vars.md)

```markdown
## Quick Reference

| Context | Syntax | Scope | Priority |
|---------|--------|-------|----------|
| Global | `vars:` at root | All tasks | 6 |
| Task | `tasks.<name>.vars:` | Single task | 3 |
| Include | `includes.<ns>.vars:` | Included Taskfile | 5 |
| Calling | `task: name, vars: {}` | Single invocation | 2 |
| Dependency | `deps[].vars:` | Dep execution | 2 |
| CLI | `task name VAR=val` | Single run | 1 (highest) |

**Priority**: 1 = highest (overrides all), 7 = lowest
```

## Success Metrics

1. **Time to find information** - Track how long it takes users to find property details
2. **Documentation completeness** - Measure coverage of all property contexts
3. **User satisfaction** - Survey users on new structure
4. **Bounce rate** - Lower bounce rate on documentation pages
5. **Cross-references used** - Track clicks on "See Also" links

## Risks & Mitigation

### Risk: Duplication of content
**Mitigation**: Keep guide.md high-level/narrative, property docs reference-focused

### Risk: Maintenance burden  
**Mitigation**: Consistent template makes updates predictable, single source per property

### Risk: User confusion
**Mitigation**: Clear navigation, prominent links from old structure to new

### Risk: Breaking existing bookmarks
**Mitigation**: Keep existing URLs, add redirects only if necessary

## Alternative Considered

We could enhance the existing schema.md instead, but:
- Schema.md is too technical (JSON schema focused)
- Doesn't show practical usage patterns
- Hard to navigate (840 lines)
- Doesn't explain context/priority clearly

The property-centric approach provides a better user experience for day-to-day Taskfile development.
