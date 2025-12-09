---
mode: agent
model: Claude Sonnet 4.5
---

<!-- markdownlint-disable-file -->

# Verify Coding Standards

Comprehensive verification and quality assurance for code changes.

## Objective

Systematically verify that all new and modified code meets project standards,
conventions, and quality requirements before finalizing changes.

## Preparation Phase

**MANDATORY before verification:**

1. **Identify scope of changes:**

   - Use `get_changed_files` to identify all modified, added, and deleted files.
     If that does not find any, use `git show --name-only HEAD`.
   - Review git diff to understand extent of changes
   - Identify which languages and frameworks are affected (Python, React,
     Docker, etc.)

2. **Load applicable coding standards:**

   - #file:../../.github/instructions/coding-best-practices.instructions.md for
     all code
   - #file:../../.github/instructions/python.instructions.md for Python files
   - #file:../../.github/instructions/reactjs.instructions.md for
     React/TypeScript files
   - #file:../../.github/instructions/self-explanatory-code-commenting.instructions.md
     for commenting style
   - #file:../../.github/instructions/containerization-docker-best-practices.instructions.md
     for Docker files

3. **Review current state:**
   - Check if changes file exists in `.copilot-tracking/changes/`
   - Read existing test files to understand testing patterns
   - Identify affected Docker containers and services

## Verification Process

Execute verification steps **in order**. Fix issues immediately when found
before proceeding.

### 1. Code Convention Compliance

**For each modified file:**

- [ ] **Update Comments and Documentation:**
  - Ensure that comments and docstrings are still accurate after modifications

**For each modified/new file:**

- [ ] **Python files (.py):**

  - Functions use type hints (Python 3.11+)
  - Imports follow Google Style Guide conventions (modules only, at top of file)
  - Naming: `snake_case` for functions/variables, `PascalCase` for classes,
    `UPPER_SNAKE_CASE` for constants
  - Docstrings present for all public functions/classes (PEP 257)
  - No obvious comments that state what code does (self-documenting code)
  - Uses `uv` for dependency management where applicable

- [ ] **React/TypeScript files (.tsx, .ts, .jsx, .js):**

  - TypeScript interfaces defined for props and state
  - Functional components with hooks (no class components)
  - Proper hook dependency arrays
  - Component names use PascalCase
  - Props destructured at function signature
  - Meaningful component and function names

- [ ] **Docker files:**
  - Multi-stage builds used where appropriate
  - Security best practices followed (non-root user, minimal base images)
  - Build optimization (layer caching, .dockerignore)
  - Health checks defined where applicable

### 2. Copyright and Legal Compliance

- [ ] **All new files have copyright notice:**
  - Check for AGPL-3.0 notice at top of each new file
  - Run `scripts/add-copyright` if any files missing copyright
  - Verify copyright year is current (2025)

### 3. Code Quality and Best Practices

- [ ] **Modularity and DRY:**

  - No duplicate code patterns
  - Functions follow Single Responsibility Principle
  - Common patterns extracted to reusable functions

- [ ] **Error Handling:**

  - Specific exception types used (not generic)
  - Error messages provide context
  - Resources cleaned up properly (try-finally, context managers)

- [ ] **Security:**
  - All user inputs validated
  - No hardcoded secrets or credentials
  - SQL injection prevention (parameterized queries)
  - Proper authentication/authorization checks

### 4. Testing Requirements

**Run and verify tests systematically:**

- [ ] **Unit tests exist for new/modified code:**

  - Business logic covered
  - Input validation tested
  - Edge cases handled (empty inputs, nulls, boundary conditions)
  - State changes verified
  - Return values validated
  - Error conditions tested

- [ ] **Test quality checks:**

  - Tests focus on behavior, not implementation
  - Test names clearly describe what is being tested
  - No vacuous tests (tests that just call code without assertions)
  - Mocks used appropriately for external dependencies

- [ ] **Coverage verification:**

  - Run coverage tools for affected code

  - Aim for minimum 80% coverage on new code.

  - To run coverage on individual functions, use dot notation not slashes (i.e.
    services.scheduler.generic_scheduler_daemon not
    services/scheduler/generic_scheduler_daemon)
  - **Document coverage numbers in changes file**

### 5. Build and Integration Verification

**Verify system-level integration:**

- [ ] **Docker containers build successfully:**

  - Identify affected containers from changes
  - Run `docker compose build [service-name]` for each affected service
  - Fix build errors before proceeding

- [ ] **Integration tests pass:**
  - Run `./scripts/run-integration-tests.sh`
  - All tests must pass
  - Fix any failures immediately

### 6. Final Lint and Error Check

**MUST be performed last** (after all other fixes):

- [ ] **Run linters and type checkers:**

  - Python: `uv run ruff check .` and `uv run ruff format --check .`
  - TypeScript: Check for compilation errors in IDE
  - Fix all errors and warnings

- [ ] **Verify no compile/lint errors:**
  - Use `get_errors` tool to check for remaining issues
  - Address all errors before completion
  - Re-run after fixes to ensure no new issues introduced

## Changes File Update Requirements

**If changes file exists** (`.copilot-tracking/changes/*.md`):

- [ ] **Append meaningful changes only:**
  - Add entries to Added/Modified/Removed sections
  - Use relative file paths
  - One-sentence summary per file
  - **Do NOT document verification activities** (e.g., "Fixed lint issues",
    "Added tests")
  - Only document functional/feature changes

## Completion Checklist

Before marking verification complete:

- [x] All code convention checks passed
- [x] Copyright notices added to all new files
- [x] Unit tests exist and pass for new/modified code
- [x] Test coverage meets 80% minimum (documented in changes file)
- [x] All affected Docker containers build successfully
- [x] Integration tests pass
- [x] No compile or lint errors remain
- [x] Changes file updated appropriately (if applicable)

## Quality Standards

This verification ensures:

- **Correctness**: Code functions as intended
- **Maintainability**: Code is readable and follows conventions
- **Testability**: Comprehensive test coverage exists
- **Security**: Best practices followed
- **Integration**: System-level compatibility verified
