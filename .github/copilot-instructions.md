# Copilot Instructions for Charcoal

## Project Overview

Charcoal is a database-agnostic API query filtering library written in Go. It parses query strings and Go structs into an abstract token tree that backend adapters (MySQL, Postgres, Mongo, GraphQL, etc.) can translate into native queries. The token language is intentionally data-store agnostic — the same tokens work regardless of whether the backing store is SQL, NoSQL, a REST API, or a spreadsheet.

## Architecture

```
tokens  ←  filter  ←  query  ←  adapters
```

- **`tokens`**: Pure data contract — struct definitions only, no logic. Shared by all other packages.
- **`filter`**: Introspects Go structs and JSON to produce a `Fields` map (field name → type). Defines operator constants and mappings.
- **`query`**: Parses URL query strings + a `Fields` map into a `tokens.Tokens` tree.
- **`adapters/`**: Backend-specific translators that consume tokens. Anyone should be able to build an adapter inside or outside the project using only public API.

No circular dependencies. New packages should respect this dependency flow.

## Go Proverbs & Guiding Principles

Follow the [Go Proverbs](https://go-proverbs.github.io/):

- **The bigger the interface, the weaker the abstraction.** Keep interfaces small. The codebase currently uses concrete types — only introduce interfaces when a real abstraction boundary exists (e.g., the adapter layer).
- **Make the zero value useful.** Design structs so their zero value is a valid, sensible default.
- **A little copying is better than a little dependency.** Don't pull in external dependencies for small utilities. The project has zero external dependencies — keep it that way unless absolutely necessary.
- **Clear is better than clever.** Write straightforward code. Avoid unnecessary abstractions, metaprogramming, or "clever" tricks.
- **Don't just check errors, handle them gracefully.** Use error accumulation (`errors.Join`) when parsing to return all errors in one round-trip rather than failing on the first.
- **Don't panic.** Reserve `panic` for truly unrecoverable programmer errors (invalid config at init time). Never panic on user input.
- **Errors are values.** Use sentinel errors (`var ErrFoo = errors.New(...)`) and typed errors (custom types implementing `Error()`) to enable `errors.Is()` matching.

## Design Constraints

- **No recursion.** This is an explicit design principle. Use iterative approaches (queues, stacks, index-based loops) for tree traversal and nested struct inspection.
- **No external dependencies.** Standard library only.
- **Data-store agnosticism.** Core packages (`filter`, `query`, `tokens`) must never contain backend-specific logic. All backend awareness belongs in adapters.
- **Adapters use public API only.** Adapters must be buildable using only exported Charcoal functionality, so third parties can create their own.

## Code Style

### Naming

- **Packages**: Short, lowercase, single-word (`filter`, `query`, `tokens`).
- **Exported types**: PascalCase (`Filter`, `Fields`, `Config`, `Clause`).
- **Unexported types**: camelCase (`dataKind`, `runeSpace`).
- **Constants**: Typed `uint8` with `iota` for enums (`OpEq`, `TypeNumber`, `JoinOp`).
- **Files**: snake_case (`filter_test.go`, `test_types.go`).
- **Maintain consistency with existing names**, even unconventional ones (e.g., `CannonicalStartRunes`).

### Error Handling

- Define sentinel errors as package-level `var` blocks with `errors.New()`.
- Use typed errors (custom types implementing `Error()`) when errors need to carry context (e.g., `FieldNotFoundError`, `TypeMismatchError`).
- Accumulate multiple errors with `errors.Join()` rather than failing fast — return a complete error report.
- Wrap accumulated errors with a top-level sentinel (e.g., `errors.Join(ErrParsingQuery, parseErr)`).
- Use `errors.Is()` for error comparison in tests.

### Types

- Prefer value types over pointer types. `Filter`, `Config`, `Clause`, `Tokens` are all value types.
- Constructor functions (e.g., `New()`) return by value, not pointer.

### Documentation

- Every package gets a `doc.go` with a package-level `/* ... */` comment block.
- Write doc comments on exported functions and types.
- Comments can be informal and have personality — this is a stylistic choice of the project.

## Testing

### Table-Driven Tests — Always

Every test function must use table-driven tests. No exceptions.

```go
type fooTestCase struct {
    name     string
    input    string
    expected string
    err      error
}

var fooTestCases = []fooTestCase{
    {
        name:     "descriptive name",
        input:    "...",
        expected: "...",
    },
}

func TestFoo(t *testing.T) {
    for _, tc := range fooTestCases {
        t.Run(tc.name, func(t *testing.T) {
            // ...
        })
    }
}
```

### Test Conventions

- **Dedicated struct type per test function**: Declare a named struct type for each test's cases (not anonymous structs).
- **Package-level test slices**: Declare test case slices as package-level `var` — not inside the test function.
- **`t.Run` subtests**: Always use `t.Run(tc.name, ...)` for each case.
- **`t.Fatalf` for error checks** (early exit on unexpected errors), **`t.Errorf` for value comparisons** (continue checking other fields).
- **Custom equality helpers** over `reflect.DeepEqual` for complex types — write dedicated comparison functions (e.g., `clausesAreEqual()`). Use `reflect.DeepEqual` only for simple/flat types.
- **No test frameworks.** Use the standard `testing` package only.

### Test Organization

- **Same-package tests** for unit testing unexported functions (`package query` in `parse_test.go`).
- **Separate `tests/` subpackages** for black-box integration tests that only exercise the exported API (`internal/filter/tests/`, `internal/query/tests/`).
- **Test helper types** go in dedicated files (`test_types.go`).

## What NOT to Do

- Do not add external dependencies.
- Do not use recursion.
- Do not put backend-specific logic in core packages.
- Do not use anonymous structs for test cases.
- Do not use `reflect.DeepEqual` for complex nested types — write explicit comparison helpers.
- Do not use `interface{}` / `any` when a concrete type will do.
- Do not create interfaces preemptively — wait until there are multiple concrete implementations.
