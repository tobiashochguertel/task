# Documentation Restructuring Proposals

This directory contains three comprehensive proposals for restructuring the Taskfile.dev documentation to improve findability, usability, and developer experience.

## Background

The current Taskfile documentation faces several challenges:

1. **Information Scatter**: Documentation for properties like `vars:` is spread across multiple files
   - guide.md (2,386 lines)
   - reference/schema.md (840 lines)
   - Multiple sections in different contexts

2. **Context Confusion**: Users don't know where they can use specific properties
   - `vars:` can be defined at root level, task level, include level, call-time, and in dependencies
   - This information is scattered across different sections

3. **No Clear Path**: Different user types (beginners, daily users, power users) all use the same documentation

4. **Hard to Find**: Developers often search repeatedly for the same information

## The Problem: A Real Example

When a developer wants to understand `vars:`, they currently need to:

1. Read guide.md lines 1105-1230 (Variables section)
2. Check guide.md line 431 (Vars of included Taskfiles)  
3. Look at reference/schema.md lines 83-110 (root vars)
4. Find reference/schema.md lines 302-313 (include vars)
5. Search for examples throughout both files
6. Remember priority order from scattered mentions

This same pattern applies to all major properties: `cmds`, `deps`, `sources`, `generates`, `preconditions`, `status`, etc.

## Three Proposals

### [Proposal 1: Property-Centric Reference Guide](./proposal-1-property-reference-guide.md)

**Philosophy**: Organize documentation by Taskfile properties

**Structure**: One comprehensive document per property (vars.md, cmds.md, deps.md, etc.)

**Best For**: 
- Developers who know what property they want to use
- Quick reference lookup
- Understanding all contexts where a property works

**Effort**: ~28-36 hours

**Key Innovation**: 
- Shows ALL contexts where each property can be used
- Clear priority/precedence tables
- Consistent structure across all property docs

### [Proposal 2: Context-Based Cookbook](./proposal-2-context-based-cookbook.md)

**Philosophy**: Organize documentation by developer intent ("I want to...")

**Structure**: 35+ recipes organized by categories (data-sharing, task-orchestration, optimization, etc.)

**Best For**:
- Developers who know what they want to achieve
- Practical, copy-paste solutions
- Learning by example

**Effort**: ~48 hours

**Key Innovation**:
- Task-oriented problem-solving
- Quick recipe tables
- Real-world patterns and troubleshooting

### [Proposal 3: Hybrid Three-Tier System](./proposal-3-hybrid-three-tier-system.md) ⭐ **Recommended**

**Philosophy**: Serve different user types with appropriate documentation tiers

**Structure**: Three complementary tiers:
1. **Learning Tier** - Tutorial-focused for beginners
2. **Cookbook Tier** - Problem-solution recipes for daily work
3. **Reference Tier** - Complete technical specifications for power users

**Best For**:
- All user types
- Progressive learning path
- Sustainable long-term documentation

**Effort**: ~96 hours (can be done incrementally)

**Key Innovation**:
- Combines benefits of Proposals 1 & 2
- Clear user journeys for different experience levels
- Separated concerns (learning vs. doing vs. reference)

## Comparison Matrix

| Feature | Proposal 1 | Proposal 2 | Proposal 3 |
|---------|-----------|-----------|-----------|
| **Organization** | By property | By use case | By tier (both!) |
| **Findability** | ⭐⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Beginner-Friendly** | ⭐⭐ | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Daily Use** | ⭐⭐⭐ | ⭐⭐⭐⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Complete Reference** | ⭐⭐⭐⭐⭐ | ⭐⭐ | ⭐⭐⭐⭐⭐ |
| **Implementation Effort** | Medium | Medium-High | High |
| **Long-term Maintenance** | Good | Good | Excellent |
| **Community Contribution** | Medium | High | High |

## Recommendation

**Proposal 3 (Hybrid Three-Tier System)** is recommended because:

1. ✅ Solves all identified problems
2. ✅ Serves all user types effectively
3. ✅ Provides clear learning paths
4. ✅ Maintains complete technical reference
5. ✅ Enables practical problem-solving
6. ✅ Sustainable for long-term growth
7. ✅ Can be implemented incrementally

## Implementation Approach

If Proposal 3 is chosen, implement in phases:

### Phase 1: Foundation (Week 1-2)
- Set up three-tier directory structure
- Create navigation
- Design templates

### Phase 2: Learning Tier (Week 3-4)
- Refactor guide.md into focused modules
- No disruption to existing docs

### Phase 3: Cookbook Tier (Week 5-8)
- Start with top 10 most-requested patterns
- Gather feedback, iterate
- Community can contribute additional recipes

### Phase 4: Reference Tier (Week 9-12)
- Build property references incrementally
- Start with most-used properties (vars, cmds, deps)
- Maintain backward compatibility

### Phase 5: Integration (Week 13)
- Complete cross-linking
- Final testing
- Launch

## Next Steps

1. **Review**: Read all three proposals in detail
2. **Decide**: Choose which approach (or combination) to pursue
3. **Plan**: Create detailed implementation timeline
4. **Execute**: Start with foundation and iterate

## Questions?

These proposals are starting points for discussion. Each can be adapted based on:
- Team capacity
- Community feedback
- Technical constraints
- Timeline requirements

## Files

- [`proposal-1-property-reference-guide.md`](./proposal-1-property-reference-guide.md) - Property-centric organization
- [`proposal-2-context-based-cookbook.md`](./proposal-2-context-based-cookbook.md) - Use-case driven recipes
- [`proposal-3-hybrid-three-tier-system.md`](./proposal-3-hybrid-three-tier-system.md) - Three-tier hybrid approach ⭐

---

**Created**: November 17, 2024  
**Status**: Proposals for review  
**Repository**: https://github.com/tobiashochguertel/task (fork)
