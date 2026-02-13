# Evaluation Steps Redesign — Tasks Overview

**Specification**: `docs/transparent-mode/specs/11-EVALUATION-STEPS-REDESIGN.md`

**See detailed task descriptions in:**

- [Features](./TASKS-FEATURES.md) - Implementation tasks

## Quick Status Overview

| Category | Total | Done | In Progress | TODO |
| -------- | ----- | ---- | ----------- | ---- |
| Features | 6     | 0    | 0           | 6    |

## Task Sets

### Set 1: Data Model & Analyzer

**Priority**: High
**Description**: Replace the flat `DetailedSteps` model with action-grouped `EvalActions` and rewrite the analyzer to produce correct, source-line-aware evaluation steps.

| Order | Task ID | Title                                         | Status  |
| ----- | ------- | --------------------------------------------- | ------- |
| 1     | T001    | Add EvalAction data model                     | 🔲 TODO |
| 2     | T002    | Implement AnalyzeEvalActions in pipe_analyzer | 🔲 TODO |

### Set 2: Rendering

**Priority**: High
**Description**: Update both human-readable and JSON renderers to use the new action-grouped model.

| Order | Task ID | Title                                          | Status  |
| ----- | ------- | ---------------------------------------------- | ------- |
| 1     | T003    | Update human-readable renderer for EvalActions | 🔲 TODO |
| 2     | T004    | Update JSON renderer for EvalActions           | 🔲 TODO |
| 3     | T005    | Update whitespace visibility for EvalActions   | 🔲 TODO |

### Set 3: Integration & Testing

**Priority**: High
**Description**: Wire up the new analyzer, update tests, regenerate golden files.

| Order | Task ID | Title                                          | Status  |
| ----- | ------- | ---------------------------------------------- | ------- |
| 1     | T006    | Wire up, update tests, regenerate golden files | 🔲 TODO |

## Task Summary

| ID   | Category | Title                                          | Priority | Status  | Dependencies |
| ---- | -------- | ---------------------------------------------- | -------- | ------- | ------------ |
| T001 | Feature  | Add EvalAction data model                      | 🟢 High  | 🔲 TODO | -            |
| T002 | Feature  | Implement AnalyzeEvalActions in pipe_analyzer  | 🟢 High  | 🔲 TODO | T001         |
| T003 | Feature  | Update human-readable renderer for EvalActions | 🟢 High  | 🔲 TODO | T001         |
| T004 | Feature  | Update JSON renderer for EvalActions           | 🟢 High  | 🔲 TODO | T001         |
| T005 | Feature  | Update whitespace visibility for EvalActions   | 🟡 Med   | 🔲 TODO | T001         |
| T006 | Feature  | Wire up, update tests, regenerate golden files | 🟢 High  | 🔲 TODO | T002-T005    |
