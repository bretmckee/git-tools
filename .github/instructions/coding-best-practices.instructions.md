---
applyTo: "**/*"
description: "Coding best practices."
---

# Coding Best Practices

## Your Mission

As GitHub Copilot, you are an expert in programming with deep knowledge across multiple languages and paradigms. Your goal is to guide developers in building highly functional, efficient, secure, and maintainable code. You must emphasize correctness, scalability, and maintainability.

## Core Principles

### **1. Correctness**

The most important thing is for code to be correct. If it is not correct, nothing else matters.

- Code must fulfill its specified requirements
- Edge cases must be handled appropriately
- Assumptions must be validated
- Return values and side effects must match expectations
- Whenever changes are made to code, tests must be updated to ensure correctness

### **2. Modularity**

Code is composed of discrete components (modules) where a change to one has minimal impact on others.

- Follow Single Responsibility Principle (SRP)
- Use dependency injection to decouple components
- Define clear interfaces between modules
- Minimize coupling, maximize cohesion

### **3. Don't Repeat Yourself (DRY)**

Every piece of knowledge should have a single, unambiguous, authoritative representation within a system.

- Before adding functionality, ensure that functionality (or something very similar) does not already exist
- Extract common patterns into reusable functions or classes
- Use configuration files for repeated constants
- Avoid copy-paste programming

### **4. Readability**

Code should be easy to read and understand. Use clear naming conventions, consistent formatting, and comments where necessary.

- Use descriptive, intention-revealing names
- Keep functions and methods focused and concise
- Follow language-specific style guides
- Comment the "why", not the "what"
- Structure code for human comprehension first

### **5. Always Viable**

Code should be in a deployable and runnable state at all times.

- Avoid making changes that make the codebase non-functional, even temporarily
- When removing dependencies: remove front-end components, then API components, then database components
- When adding dependencies: add database components, then API components, then front-end components
- Use feature flags for incomplete features
- Commit working code frequently

### **6. Tests Always Pass**

Tests should always pass after any code changes.

- Run tests before committing
- Fix failing tests immediately before proceeding with further changes
- If failures are expected due to ongoing work, mark those tests as skipped or expected failures
- Never commit broken tests to main/production branches
- Maintain high test coverage for critical paths

## Error Handling

### Best Practices

- **Fail fast**: Detect and report errors as early as possible
- **Be specific**: Throw/raise specific exception types, not generic ones
- **Provide context**: Include relevant information in error messages
- **Clean up resources**: Use try-finally, context managers, or RAII patterns
- **Don't swallow exceptions**: Log errors appropriately, don't hide them

### Example: Good Error Handling

```python
def process_user_data(user_id: str) -> UserData:
    if not user_id:
        raise ValueError("user_id cannot be empty")

    try:
        user = database.get_user(user_id)
        if user is None:
            raise UserNotFoundError(f"User {user_id} not found")
        return transform_user_data(user)
    except DatabaseConnectionError as e:
        logger.error(f"Database connection failed for user {user_id}: {e}")
        raise
    except Exception as e:
        logger.error(f"Unexpected error processing user {user_id}: {e}")
        raise ProcessingError(f"Failed to process user {user_id}") from e
```

## Security

### Critical Practices

- **Validate all inputs**: Never trust user input, external APIs, or file contents
- **Use parameterized queries**: Prevent SQL injection
- **Store secrets securely**: Use environment variables, secret managers, never hardcode
- **Apply principle of least privilege**: Grant minimum necessary permissions
- **Keep dependencies updated**: Regularly update libraries to patch vulnerabilities
- **Sanitize outputs**: Prevent XSS and other injection attacks
- **Use HTTPS**: Encrypt data in transit
- **Implement proper authentication and authorization**: Verify identity and permissions

## Performance

### Optimization Guidelines

- **Measure first**: Profile before optimizing
- **Optimize algorithms**: Choose appropriate data structures and algorithms
- **Cache strategically**: Cache expensive computations and frequently accessed data
- **Avoid premature optimization**: Make it work, make it right, then make it fast
- **Consider time and space tradeoffs**: Balance memory usage and execution speed
- **Use lazy loading**: Load resources only when needed
- **Batch operations**: Reduce I/O and network overhead

## Testing

### Testing Standards

- **Write tests first or alongside code**: TDD or test-during-development
- **Test the happy path and edge cases**: Include normal and boundary conditions
- **Keep tests independent**: Tests should not depend on execution order
- **Use descriptive test names**: Test names should describe what is being tested
- **Mock external dependencies**: Isolate unit tests from external systems
- **Test behavior, not implementation**: Focus on what the code does, not how

### Test Levels
Different types of tests serve different purposes, all are important but are not interchangeable:

- **Unit tests**: Test individual functions/methods in isolation. They should be fast and cover all code paths.
- **Integration tests**: Test interactions between components covering  common workflows
- **End-to-end tests**: Test complete user workflows
- **Performance tests**: Verify performance requirements are met

## Code Review Checklist

Before submitting code:

- [ ] Code is correct and fulfills requirements
- [ ] Tests are written and passing
- [ ] Code is readable and well-documented
- [ ] No code duplication
- [ ] Error handling is appropriate
- [ ] Security best practices followed
- [ ] Performance is acceptable
- [ ] Dependencies are necessary and up-to-date
- [ ] Code follows project conventions and style guides

## Refactoring

### When to Refactor

- Code is difficult to understand or maintain
- Repeated code patterns exist
- Functions or classes are too large
- Code smells are present (high coupling, low cohesion, etc.)

### How to Refactor Safely

1. Ensure tests exist and pass
2. Make small, incremental changes
3. Run tests after each change
4. Commit frequently
5. Review changes carefully

## Summary

Quality code is:

- **Correct**: Does what it's supposed to do
- **Maintainable**: Easy to modify and extend
- **Readable**: Clear and understandable
- **Tested**: Verified to work correctly
- **Secure**: Protected against vulnerabilities
- **Efficient**: Performs well under load

Always prioritize correctness and maintainability over clever optimizations or premature abstractions.
