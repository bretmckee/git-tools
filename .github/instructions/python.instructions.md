---
description: "Python coding conventions and guidelines"
applyTo: "**/*.py"
---

# Python Coding Conventions

## Python Instructions

- Ensure functions have descriptive names and always use modern (python 3.13+)
  type hints.
- Use pydantic for type validation.
- Break down complex functions into smaller, more manageable functions.

## General Instructions

- Always prioritize readability and clarity.
- For algorithm-related code, include explanations of the approach used.
- Write code with good maintainability practices, including comments on why
  certain design decisions were made.
- Use consistent naming conventions and follow language-specific best practices.
- Write concise, efficient, and idiomatic code that is also easily
  understandable.
- Whenever code is added or modified, ensure that all tests pass and that the
  code adheres to the specified style guidelines using `ruff` and as described
  here.

## Tooling

- Use `uv` (https://uv.run/) for managing virtual environments and dependencies.
  - This means you will always have to use `uv run` to execute scripts, tests
    and commands.
- Use the (src
  layout)[https://packaging.python.org/en/latest/discussions/src-layout-vs-flat-layout/]
  for project structure.
- Use `pytest` for testing.
- Use `ruff` for code formatting, import sorting, and linting.

## Code Style and Formatting

- Follow the **PEP 8** style guide for Python.
- Place function and class docstrings immediately after the `def` or `class`
  keyword.
- Use blank lines to separate functions, classes, and code blocks where
  appropriate.
- Imports must be at the top of the file, never inside functions.
- Imports must be arranged by isort.
- Follow the recommendations for these sections from the Google style guide:

## Imports

Inspired by the Google Python Style Guide section 2.2.4:

- Use import only for packages and modules.
- Do no import package or module contents (e.g. functions/classes/etc. directly)
  except for testing and typing.
- Use from x import y where x is the package prefix and y is the module name
  with no prefix.
- Use from x import y as z in only in one of the following circumstances:
  - Two modules named y are to be imported.
  - y conflicts with a top-level name defined in the current module.
  - y conflicts with a common parameter name that is part of the public API
    (e.g., features).
  - y is an inconveniently long name.
  - y is too generic in the context of your code (e.g., from storage.file_system
    import options as fs_options).
  - If import y as z is a standard abbreviation (e.g., import numpy as np).

### Examples of Imports

#### Good

```python
  # Good: Imports a module
  from a.b import c
  ...
  c.DoSomething(input, output)
```

#### Bad

```python
  # Bad: Imports something other than a module or package
  from a.b import c
  ...
  c(input, output)
```

- 3.4.1 Trailing commas in sequences of items?

  ```

  Trailing commas in sequences of items are recommended only when the closing container token ], ), or } does not appear on the same line as the final element, as well as for tuples with a single element. The presence of a trailing comma is also used as a hint to our Python code auto-formatter Black or Pyink to direct it to auto-format the container of items to one item per line when the , after the final element is present.

  ```

### Naming Conventions

- Name modules, packages, methods, functions, global variables, instance
  variables, function parameters, local variables with snake_case.
- Name classes and exceptions using pascal case (camel case with the first
  letter capitalized).
- Name global constants in all upper caps snake case.

These names are all self describing (i.e. module_name shows how to name modules)

```python
module_name, package_name, ClassName, method_name, ExceptionName, function_name, GLOBAL_CONSTANT_NAME, global_var_name, instance_var_name, function_parameter_name, local_var_name, query_proper_noun_for_thing, send_acronym_via_https.
```

### Comments and Docstrings

- Write clear and concise comments for each function.
- Provide docstrings following PEP 257 conventions.
- For libraries or external dependencies, mention their usage and purpose in
  comments.
- Follow the recommendations for these sections from the Google style guide:

  ```
  3.8.1 Docstrings

  Python uses docstrings to document code. A docstring is a string that is the first statement in a package, module, class or function. These strings can be extracted automatically through the __doc__ member of the object and are used by pydoc. (Try running pydoc on your module to see how it looks.) Always use the three-double-quote """ format for docstrings (per PEP 257). A docstring should be organized as a summary line (one physical line not exceeding 80 characters) terminated by a period, question mark, or exclamation point. When writing more (encouraged), this must be followed by a blank line, followed by the rest of the docstring starting at the same cursor position as the first quote of the first line. There are more formatting guidelines for docstrings below.

  3.8.2 Modules

  Every file should contain license boilerplate. Choose the appropriate boilerplate for the license used by the project (for example, Apache 2.0, BSD, LGPL, GPL).

  Files should start with a docstring describing the contents and usage of the module.

  3.8.2.1 Test modules

  Module-level docstrings for test files are not required. They should be included only when there is additional information that can be provided.

  Examples include some specifics on how the test should be run, an explanation of an unusual setup pattern, dependency on the external environment, and so on.

  Docstrings that do not provide any new information should not be used. (e.g. """Tests for foo.bar.""")

  3.8.3 Functions and Methods

  In this section, “function” means a method, function, generator, or property.

  A docstring is mandatory for every function that has one or more of the following properties:

  being part of the public API
  nontrivial size
  non-obvious logic
  A docstring should give enough information to write a call to the function without reading the function’s code. The docstring should describe the function’s calling syntax and its semantics, but generally not its implementation details, unless those details are relevant to how the function is to be used. For example, a function that mutates one of its arguments as a side effect should note that in its docstring. Otherwise, subtle but important details of a function’s implementation that are not relevant to the caller are better expressed as comments alongside the code than within the function’s docstring.

  The docstring may be descriptive-style ("""Fetches rows from a Bigtable.""") or imperative-style ("""Fetch rows from a Bigtable."""), but the style should be consistent within a file. The docstring for a @property data descriptor should use the same style as the docstring for an attribute or a function argument ("""The Bigtable path.""", rather than """Returns the Bigtable path.""").

  Certain aspects of a function should be documented in special sections, listed below. Each section begins with a heading line, which ends with a colon. All sections other than the heading should maintain a hanging indent of two or four spaces (be consistent within a file). These sections can be omitted in cases where the function’s name and signature are informative enough that it can be aptly described using a one-line docstring.

  Args:
  List each parameter by name. A description should follow the name, and be separated by a colon followed by either a space or newline. If the description is too long to fit on a single 80-character line, use a hanging indent of 2 or 4 spaces more than the parameter name (be consistent with the rest of the docstrings in the file). The description should include required type(s) if the code does not contain a corresponding type annotation. If a function accepts *foo (variable length argument lists) and/or \*\*bar (arbitrary keyword arguments), they should be listed as *foo and \*\*bar.
  Returns: (or Yields: for generators)
  Describe the semantics of the return value, including any type information that the type annotation does not provide. If the function only returns None, this section is not required. It may also be omitted if the docstring starts with “Return”, “Returns”, “Yield”, or “Yields” (e.g. """Returns row from Bigtable as a tuple of strings.""") and the opening sentence is sufficient to describe the return value. Do not imitate older ‘NumPy style’ (example), which frequently documented a tuple return value as if it were multiple return values with individual names (never mentioning the tuple). Instead, describe such a return value as: “Returns: A tuple (mat_a, mat_b), where mat_a is …, and …”. The auxiliary names in the docstring need not necessarily correspond to any internal names used in the function body (as those are not part of the API). If the function uses yield (is a generator), the Yields: section should document the object returned by next(), instead of the generator object itself that the call evaluates to.
  Raises:
  List all exceptions that are relevant to the interface followed by a description. Use a similar exception name + colon + space or newline and hanging indent style as described in Args:. You should not document exceptions that get raised if the API specified in the docstring is violated (because this would paradoxically make behavior under violation of the API part of the API).

  3.8.3.1 Overridden Methods

  A method that overrides a method from a base class does not need a docstring if it is explicitly decorated with @override (from typing_extensions or typing modules), unless the overriding method’s behavior materially refines the base method’s contract, or details need to be provided (e.g., documenting additional side effects), in which case a docstring with at least those differences is required on the overriding method.

  3.8.4 Classes

  Classes should have a docstring below the class definition describing the class. Public attributes, excluding properties, should be documented here in an Attributes section and follow the same formatting as a function’s Args section.

  All class docstrings should start with a one-line summary that describes what the class instance represents. This implies that subclasses of Exception should also describe what the exception represents, and not the context in which it might occur. The class docstring should not repeat unnecessary information, such as that the class is a class.

  3.8.5 Block and Inline Comments

  The final place to have comments is in tricky parts of the code. If you’re going to have to explain it at the next code review, you should comment it now. Complicated operations get a few lines of comments before the operations commence. Non-obvious ones get comments at the end of the line.
  ```

### Docstring Anti-Patterns

- Do not mention users or callers in doc strings. The purpose of doc strings is
  to explain the code. The callers and users can change over time, rendering the
  docstring obsolete.

  ```python
  // Bad: Never name users or callers in doc strings
  def sync_get_data():
  """
  Provides data for Celery and other async systems.
  """
  ...
  ```

## Edge Cases and Testing

- Handle edge cases and write clear exception handling.
- Always include test cases for critical paths of the application.
- Account for common edge cases like empty inputs, invalid data types, and large
  datasets.
- Include comments for edge cases and the expected behavior in those cases.
- Write unit tests for functions and document them with docstrings explaining
  the test cases.

## Example of Proper Documentation

```python
def calculate_area(radius: float) -> float:
  """
  Calculate the area of a circle given the radius.

  Parameters:
  radius (float): The radius of the circle.

  Returns:
  float: The area of the circle, calculated as π * radius^2.
  """
  import math
  return math.pi * radius ** 2
```
