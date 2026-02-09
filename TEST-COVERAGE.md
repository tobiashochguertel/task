# Test Coverage for Variable Descriptions Feature

## Test Pyramid

This document describes the comprehensive test coverage for the variable descriptions feature across all levels of the test pyramid.

### 1. Unit Tests (Bottom of Pyramid)

#### taskfile/ast/var_test.go
- **TestVarWithDescription**: Tests parsing variables with descriptions combined with sh/ref/map
- **TestVarsWithDescriptions**: Tests parsing multiple variables with descriptions
- **TestVarDescriptionInheritance**: Tests basic description inheritance
- **TestVarDescriptionInheritanceMultipleLevels**: Tests multi-level inheritance chains
- **TestVarDescriptionWithEmptyString**: Edge case for empty description strings
- **TestVarOnlyDescriptionField**: Variables with only a description field
- **TestVarDescriptionBackwardCompatibility**: Ensures old Taskfiles without descriptions work
- **TestVarStaticValueWithDescription**: Tests the `value` field with descriptions (string, number, boolean, object)
- **TestVarMapWithDescription**: Tests the `map` field still works with descriptions
- **TestVarMutuallyExclusiveFields**: Validates that sh/ref/map/value are mutually exclusive
- **TestVarDescOnlyIsAllowed**: Confirms description-only variables are valid

**Total: 11 test functions, ~40 test cases**

#### taskfile/ast/vars_test.go (NEW)
- **TestVarsMerge**: Tests basic variable merging without descriptions
- **TestVarsMergeWithDescriptions**: Tests description inheritance during merge
  - Child inherits parent description
  - Child overrides parent description
  - Multi-level inheritance (global → include → task)
  - Preserve non-empty descriptions
  - Description-only vars
- **TestVarsMergeEdgeCases**: Edge cases
  - Empty vars merge
  - Merge with nil
  - Multiple vars with different description states

**Total: 3 test functions, ~15 test cases**

### 2. Integration Tests (Middle of Pyramid)

#### task_test.go - Integration Tests

**TestVariableDescriptions**
Tests basic variable description functionality in real Taskfiles:
- Shows variables with descriptions render correctly
- Task-level variable description overrides
- Task-level variable description inheritance
- Uses testdata/var-desc-integration/Taskfile.yml

**TestVariableDescriptionsIncludes**
Tests variable descriptions across included Taskfiles:
- Global variables work with descriptions
- Included Taskfile variables have their own descriptions
- Included tasks can inherit global variable descriptions
- Uses testdata/var-desc-includes/ with nested Taskfiles

**TestListVariablesCommand**
Tests the `--list-vars` CLI flag:
- List variables in table format with descriptions
- List variables in JSON format with descriptions
- Verifies static variables include values in JSON
- Verifies dynamic variables (sh) don't expose values
- Uses testdata/var-desc-integration/Taskfile.yml

**Total: 3 test functions, ~10 test scenarios**

### 3. End-to-End Tests (Top of Pyramid)

#### task_test.go - E2E Tests

**TestVariableDescriptionsE2E**
Real-world deployment pipeline simulation:
- Full deployment workflow (build → docker-build → deploy)
- Variable inheritance through task dependencies
- List all pipeline variables with descriptions
- JSON output includes all variable metadata
- Cleanup after execution
- Uses testdata/var-desc-e2e/Taskfile.yml (realistic CI/CD scenario)

**Total: 1 test function, 4 comprehensive scenarios**

## Test Data Files

### Unit Test Data
- `testdata/var-descriptions/Taskfile.yml` - Basic variable descriptions
- `testdata/var-descriptions/Taskfile-inheritance.yml` - Inheritance scenarios
- `testdata/var-descriptions/Taskfile-static-values.yml` - Static values with `value` field

### Integration Test Data
- `testdata/var-desc-integration/Taskfile.yml` - Integration test scenarios
- `testdata/var-desc-includes/Taskfile.yml` - Main Taskfile with includes
- `testdata/var-desc-includes/included/Taskfile.yml` - Included Taskfile

### E2E Test Data
- `testdata/var-desc-e2e/Taskfile.yml` - Real-world CI/CD pipeline

## Coverage Summary

| Test Level    | Test Files | Test Functions | Test Cases | Coverage Area |
|---------------|------------|----------------|------------|---------------|
| Unit          | 2          | 14             | ~55        | Parsing, merging, inheritance logic |
| Integration   | 1          | 3              | ~10        | CLI commands, includes, real Taskfiles |
| E2E           | 1          | 1              | 4          | Complete workflows, user scenarios |
| **TOTAL**     | **4**      | **18**         | **~69**    | **Full feature coverage** |

## Test Scenarios Covered

### ✅ Basic Functionality
- Variable parsing with descriptions
- All variable types (sh, ref, map, value)
- Static and dynamic variables
- Simple and complex (object) values

### ✅ Description Inheritance
- Global → Task inheritance
- Global → Include → Task multi-level inheritance
- Explicit override of inherited descriptions
- Empty description preserves parent description

### ✅ CLI Integration
- `--list-vars` table output
- `--list-vars --json` output
- JSON structure includes variables with descriptions
- Static vs dynamic variable value exposure

### ✅ Advanced Features
- Included Taskfiles with their own variables
- Variable descriptions across task dependencies
- Task calls with variable overrides
- Multiple levels of variable scoping

### ✅ Edge Cases
- Variables with only descriptions (no value)
- Empty description strings
- Backward compatibility (variables without descriptions)
- Mutual exclusivity validation (sh/ref/map/value)

### ✅ Real-World Usage
- Complete CI/CD pipeline
- Multiple variable types in one Taskfile
- Complex variable templating
- Task dependency chains with variable inheritance

## Running the Tests

```bash
# Run all unit tests
go test ./taskfile/ast/... -v

# Run all integration and E2E tests
go test ./... -v -run "TestVariable"

# Run specific test
go test ./... -v -run "TestVariableDescriptionsE2E"

# Run with coverage
go test ./taskfile/ast/... -cover
go test ./... -run "TestVariable" -cover
```

## Test Quality Metrics

- **Comprehensive**: Covers all feature aspects
- **Isolated**: Unit tests don't depend on external state
- **Fast**: Unit tests execute in milliseconds
- **Maintainable**: Clear test names and structure
- **Realistic**: E2E tests mirror actual user workflows
- **Documented**: Each test has clear purpose and assertions

## Future Test Considerations

1. Performance tests for large Taskfiles with many variables
2. Concurrency tests for parallel task execution with variable inheritance
3. Fuzzing tests for variable description parsing
4. Property-based testing for merge operations
5. Integration with shell completion scripts
