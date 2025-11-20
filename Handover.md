# Handover: Taskfile Documentation Refactoring

## Status Overview
**Date:** 2025-11-20
**Current State:** In Progress (Proposal 3 Implementation)

## Work Done
1. **Repository Setup:**
   - Forked `go-task/task` to private repository.
   - Configured `Taskfile.yml` in `task/website/` for local development.
   - Added `website:dev`, `website:stop`, `website:restart`, `website:logs` tasks.
   - Configured server to run on `0.0.0.0:3003` for network access.

2. **Documentation Analysis:**
   - Analyzed existing documentation structure.
   - Identified issues with discoverability (e.g., `vars` definition scattered).
   - Created 3 proposals for restructuring:
     - Proposal 1: Reference-First Approach.
     - Proposal 2: Context-Based Cookbook.
     - Proposal 3: Hybrid "Core Concepts + Usage Patterns" (Selected).

3. **Implementation (Proposal 3):**
   - Created new directory structure in `website/src/docs/`:
     - `core-concepts/`
     - `usage/`
     - `reference/`
     - `guides/`
   - Moved existing files to appropriate locations (e.g., `installation.md` -> `guides/installation.md`).
   - Created `core-concepts.md` with initial content.
   - Updated `website/.vitepress/config.mts` to reflect the new sidebar structure including "Proposals" section.

## Current Issues
- **Vue Compiler Error:**
  - File: `website/src/docs/proposal-2-context-based-cookbook.md`
  - Error: `[plugin:vite:vue] Error parsing JavaScript expression: Unexpected token (1:1)`
  - Location: Around line 272/274 (`### Pattern: Build Pipeline with Artifacts`).
  - Suspected cause: Inline code or content being misinterpreted as Vue interpolation.

## Todo
1. **Fix Build Error:**
   - Investigate and fix the Vue compiler error in `proposal-2-context-based-cookbook.md`.
   - Verify that the website builds and renders correctly without errors.

2. **Continue Proposal 3 Implementation:**
   - **Core Concepts:** Expand `core-concepts.md` to cover `vars`, `tasks`, `includes`, `deps` in depth.
   - **Reference:** Create/Move reference documentation (Schema, CLI usage) to `reference/`.
   - **Usage Patterns:** Create `usage/` documents for common patterns (e.g., Docker, Monorepo).
   - **Migration:** Ensure all content from original `docs/` is preserved and correctly categorized.

3. **Verification:**
   - Check all internal links in the documentation.
   - Verify sidebar navigation in `config.mts`.

## Operations
- **Start Server:**
  ```bash
  cd task
  task website:dev
  ```
  (Runs in a tmux session named `task-website`)

- **Stop Server:**
  ```bash
  task website:stop
  ```

- **Restart Server:**
  ```bash
  task website:restart
  ```

- **View Logs:**
  ```bash
  task website:logs
  ```

- **Access:**
  - URL: `http://localhost:3003` (or your local IP)
