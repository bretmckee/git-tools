# Global Copilot Instructions

## General Guidelines

- When generating code or making suggestions, prioritize minimizing the number
  of changes and the impact on existing code. Focus on small, targeted
  modifications rather than large refactorings or extensive additions, unless
  explicitly requested.

## Programming Practices

For any code generated, follow these guidelines:

## General Instructions

- Always prioritize readability and clarity.
- For algorithm-related code, include explanations of the approach used.
- Write code with good maintainability practices, including comments on why certain design decisions were made.
- Handle edge cases and write clear exception handling.
- For libraries or external dependencies, mention their usage and purpose in comments.
- Use consistent naming conventions and follow language-specific best practices.
- Write concise, efficient, and idiomatic code that is also easily
  understandable.

### Comments

- Strive for self-documenting code; prefer clear variable names, function names, and code structure to comments
- Write comments only when necessary to explain complex logic, business rules,
  or non-obvious behavior.
- Do not write vacuous comments that simply explain what a line of code does unless it
  is not obvious.
- Write comments in complete sentences in English by default
- Document why, not what, unless the what is complex
- Avoid using emoji in code and comments

### Edge Cases and Testing

- Always include test cases for critical paths of the application.
- Account for common edge cases like empty inputs, invalid data types, and large datasets.
- Include comments for edge cases and the expected behavior in those cases.
- Write unit tests for functions and document them with docstrings explaining the test cases.
- Whenever you write a test, ensure that it passes before finalizing the code.

### Use Language-Specific Guidelines

- Follow the specific instructions provided in the language-specific instruction
  files located in the `.github/instructions` directory. Note that they may not be
  read automatically for some operations because their applyTo patterns may not match.
  For example, when creating a python project, since there are no existing python files,
  the instructions in `.github/instructions/python-instructions.md` will not be
  applied automatically. In such cases, refer to the relevant language-specific
  instructions manually to ensure compliance with the guidelines.

### Helpful Hints

- This project uses uv for python dependency and virtual environment management. Use `uv`
  commands to add, remove, and update dependencies as needed and 'uv run' to run python files.
- This project uses modern docker, so the command to run docker compose is
  'docker compose' (not 'docker-compose')
