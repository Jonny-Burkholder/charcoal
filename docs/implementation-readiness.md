# Charcoal Implementation Readiness

Assessment date: 2026-10-12

Repository state: current working tree at the time of assessment

Status: pre-release

## Executive summary

Charcoal has two materially different completion levels:

- The **core CQL implementation**—turning a URL query string into a validated,
  backend-independent token tree—is approximately **80% complete**.
- The **end-to-end library experience**—installing Charcoal, using its public
  API, and obtaining a correct backend query—is approximately **48% complete**.

These are engineering estimates, not percentages of files or lines written.
Each is based on the weighted rubric below. A reasonable confidence interval is
plus or minus five percentage points because the project does not yet define a
release contract or acceptance suite.

The parser is the most mature part of the repository. It supports filter
clauses, aliases for common operators, nested boolean groups, sorting,
pagination fields, type checking, and aggregated errors. The token model and
MySQL fragment builder are also substantial.

The project is not yet usable through the API advertised in the README.
`ToSql` calls an adapter that is never initialized, other built-in adapters are
stubs, JSON schema extraction is empty, and examples teach stale or broken
flows. There is also a correctness mismatch between parsed `in` values and the
MySQL adapter.

### Specification delta

The CQL draft was revised after this assessment to standardize operator
spellings and infer pagination from `page` or `per_page` without a separate
boolean. The implementation does not yet match those semantics. The resulting
work is tracked in [`cql-bugs.md`](cql-bugs.md); the scores below continue to
measure the current code and have not been recalculated.

## Scope definitions

### Core language

For this assessment, the core ends at:

```text
URL query string + field registry -> validated tokens or structured errors
```

Relevant code is primarily in
[`internal/query`](../internal/query),
[`internal/filter/operators.go`](../internal/filter/operators.go), and
[`internal/tokens`](../internal/tokens).

This scope does not include deriving the field registry, compiling tokens for a
database, or exposing a stable public package.

### End-to-end usability

The end-to-end scope is:

```text
install module
  -> derive an allowed schema
  -> parse a user query
  -> compile safe backend fragments
  -> execute them from a documented example
```

This includes schema extraction, the public API, adapters, packaging,
documentation, examples, and automated quality gates.

## Core language score: 79%, rounded to 80%

The weighted score is 78.9 out of 100.

### URL parameter extraction and orchestration: 9.0 / 10

Implemented:

- Decodes URL query parameters.
- Coordinates filters, sort, and pagination.
- Returns an empty token tree if any component fails.
- Aggregates independent errors.

Remaining:

- Only the first repeated parameter is used, with no documented policy.
- Resource limits are absent.

Evidence:
[`internal/query/parse.go`](../internal/query/parse.go).

### Filter tokenization: 12.8 / 15

Implemented:

- Splits outer clauses while respecting quotes, parentheses, and brackets.
- Reports unbalanced delimiters.
- Preserves commas in groups and list values.

Remaining:

- The clause tokenizer handles single-quoted whitespace but not equivalent
  double-quoted whitespace.
- Escaping and quoted commas inside list values are not modeled consistently.
- Lexing is spread across several handwritten scanners with different rules.

Evidence:
[`internal/query/parse.go`](../internal/query/parse.go) and
[`internal/query/split.go`](../internal/query/split.go).

### Clause parsing, operators, and types: 15.6 / 20

Implemented:

- Equality, comparison, membership, pattern, and null operators.
- Symbolic and keyword aliases.
- String, number, boolean, and bytes validation.
- Field allowlist checks.

Remaining:

- Compact operator discovery iterates a Go map. Overlapping operators such as
  `>` and `>=` can be selected nondeterministically.
- The declared multiple-operator error is unused.
- Bytes have no meaningful validation.
- List normalization is not an explicit token type.
- Dates, UUIDs, enums, and application validators are absent.

Evidence:
[`internal/query/parse.go`](../internal/query/parse.go),
[`internal/filter/operators.go`](../internal/filter/operators.go), and
[`internal/query/errors.go`](../internal/query/errors.go).

### Boolean groups: 10.5 / 15

Implemented:

- `AND`, `OR`, comma-as-AND, parentheses, and nested token children.
- Bare `OR` and `AND` expressions are recognized.

Remaining:

- Mixed join operators at one level are not rejected; the last separator wins.
- No conventional precedence is defined.
- The downstream SQL tree walk needs integration coverage for deeply nested
  combinations.

Evidence:
[`internal/query/split.go`](../internal/query/split.go) and
[`internal/query/parse.go`](../internal/query/parse.go).

### Sorting: 8.5 / 10

Implemented:

- Multiple sort fields.
- Short and long ascending/descending spellings.
- Unknown-field and invalid-direction errors.

Remaining:

- Sort field casing differs from filter field casing.
- Repeated fields and conflicting directions are not rejected.

Evidence:
[`internal/query/sort.go`](../internal/query/sort.go).

### Pagination: 6.0 / 10

Implemented:

- Boolean enablement.
- Integer page and page-size parsing.
- Cursor capture.

Remaining:

- No positive-range or maximum-page-size validation.
- Cursor behavior is unspecified and unused by the MySQL adapter.
- Page defaults are implicit rather than part of a documented contract.

Evidence:
[`internal/query/pagination.go`](../internal/query/pagination.go).

### Token model: 5.0 / 5

The token structs represent filters, nested groups, sort clauses, and
pagination without coupling the parser to a backend.

The model is sufficient for the implemented language, although list values
would be safer as structured values rather than bracketed strings.

Evidence:
[`internal/tokens`](../internal/tokens).

### Errors: 8.0 / 10

Implemented:

- Specific errors for fields, operators, values, sort, and pagination.
- Joined errors let a client fix multiple mistakes in one pass.

Remaining:

- Error codes and machine-readable locations are absent.
- Some callers can still encounter panics outside the parser.
- Exact wrapping and matching behavior lacks public integration tests.

Evidence:
[`internal/query/errors.go`](../internal/query/errors.go) and
[`internal/query/parse.go`](../internal/query/parse.go).

### Core automated tests: 3.5 / 5

There is substantial unit coverage in
[`internal/query/parse_test.go`](../internal/query/parse_test.go) and
[`internal/query/split_test.go`](../internal/query/split_test.go).

The score is reduced because successful full-query cases are commented out and
the black-box suite in
[`internal/query/tests/parse_test.go`](../internal/query/tests/parse_test.go)
contains no cases.

## End-to-end score: 48%

The weighted score is 48.2 out of 100.

### Parser and token core: 27.7 / 35

This carries the 79% core score into the product rubric. It is the strongest
part of the project.

### Schema derivation: 6.8 / 15

Working:

- Struct reflection supports exported scalar fields, nested dot paths, and
  configurable pointer depth.
- Fields receive coarse string, number, boolean, or bytes types.

Blocking gaps:

- [`internal/filter/json.go`](../internal/filter/json.go) returns `nil, nil`.
- Struct tags and snake_case conversion promised in the README are absent.
- Unsupported field types panic.
- Invalid configuration can panic instead of returning an error.
- Field-name behavior does not align with filter and sort normalization.

### Backend adapters: 8.3 / 15

Working:

- The MySQL builder emits parameterized `WHERE`, `ORDER BY`, and
  `LIMIT`/`OFFSET` fragments.
- It supports field-to-column mapping and configured joins.
- Adapter unit tests cover many hand-built token cases.

Blocking gaps:

- Parsed membership values retain brackets, while the MySQL builder splits the
  raw value on commas. A parsed `[a, b]` can therefore bind `"[a"` and `"b]"`.
- Cursor pagination is ignored.
- Nested parser-to-SQL behavior is not tested end to end.
- Mongo, Elastic, Redis, and GraphQL built-ins are not implemented.

Evidence:
[`adapters/sql/build.go`](../adapters/sql/build.go) and
[`charcoal/builtin.go`](../charcoal/builtin.go).

### Public API and packaging: 1.5 / 15

Blocking gaps:

- `ToSql` parses successfully and then invokes a nil embedded adapter, causing
  a runtime panic.
- `ToMongo`, `ToElastic`, `ToRedis`, and `ToGraphQL` return empty strings and
  nil errors.
- `Filter` returns an unexported concrete type.
- `Activate` exposes a type under `internal/tokens`, which external modules
  cannot import for adapter development.
- The module path is `charcoal`, while README examples import the GitHub path.
- The public API has no tests.

Evidence:
[`charcoal/charcoal.go`](../charcoal/charcoal.go),
[`charcoal/builtin.go`](../charcoal/builtin.go), and
[`go.mod`](../go.mod).

### Documentation and examples: 2.5 / 10

Working:

- The README explains the product intent and shows readable CQL.
- Parser and SQL packages have implementation-level documentation.

Blocking gaps:

- The README uses APIs that do not exist (`charcoal.New`, `ToSQL`).
- It promises field conversion and tags that are not implemented.
- The simple example has an empty `main` and passes a slice where a struct is
  expected.
- The full example depends on local services and calls incomplete built-ins.
- There is no runnable minimal path from HTTP request to parameterized SQL.

Evidence:
[`README.md`](../README.md),
[`example/simple`](../example/simple), and
[`example/full`](../example/full).

### Quality and release readiness: 1.5 / 10

Blocking gaps:

- The full Go test suite has a failing black-box struct test.
- There is no continuous-integration workflow.
- There is no release/versioning policy or compatibility contract.
- There are no parser-to-adapter or public API integration tests.
- Query complexity and page-size limits are not enforced.

The GPL-3.0 license is valid but can limit adoption by proprietary services.
That is a product-positioning decision, not an implementation defect.

## What must be completed before Charcoal is usable

The following milestones define a focused, MySQL-first usable release. Other
backends can remain explicitly out of scope.

### 1. Make one public happy path work

Required outcome:

- A user can import the module, construct a filter from a struct, parse a CQL
  query, compile it with a documented MySQL configuration, and receive SQL
  fragments plus bound arguments without a panic.

Acceptance criteria:

- Set the canonical module path.
- Export stable user-facing filter, token, configuration, result, and error
  types, or hide tokens behind a complete public compile method.
- Replace the nil adapter path in `ToSql`.
- Return the structured MySQL result, including `Args`; returning only a SQL
  string is insufficient for safe execution.
- Add a public-package integration test for this exact path.

### 2. Fix correctness defects in the core-to-adapter contract

Acceptance criteria:

- Represent membership values as a structured list or normalize brackets
  before building SQL.
- Parse overlapping operators deterministically, preferably by tokenizing or
  longest-match order.
- Reject mixed boolean joins at the same group level or implement and document
  precedence.
- Use one quoting and escaping implementation for every parser stage.
- Use one field-normalization rule for filtering and sorting.
- Add end-to-end cases for nested groups, `in`, `nin`, quoted values, null
  tests, sorting, and pagination.

### 3. Define the schema contract

Acceptance criteria:

- Choose and document canonical field names: exact Go names, lower camel case,
  snake_case, or explicit tags.
- Implement the chosen rename and skip tags.
- Return errors rather than panic for unsupported types and invalid config.
- Either implement JSON-derived schemas with tests or remove JSON support from
  the first release contract.

### 4. Bound and finish pagination

Acceptance criteria:

- Validate `page >= 1` and `per_page` within a configurable safe range.
- Define defaults.
- Infer pagination from either parameter and remove the `pagination` boolean.
- Either implement cursor semantics through the MySQL adapter or reject/remove
  `cursor` for the first release.

### 5. Publish accurate documentation and a runnable example

Acceptance criteria:

- Make the root README use only real exported APIs.
- Link the README to [`cql-rfc.md`](cql-rfc.md).
- Provide a minimal example that runs without MongoDB or an unfinished
  adapter.
- Show URL encoding, adapter configuration, complete SQL assembly, bound
  arguments, error-to-HTTP-response handling, and several valid requests.
- Label Mongo, Elastic, Redis, and GraphQL as planned rather than built in.

### 6. Establish a release quality gate

Acceptance criteria:

- `go test ./...` passes from a clean checkout.
- Continuous integration runs formatting, vetting, and tests.
- Black-box CQL tests exercise the documented public grammar.
- Add fuzz tests for delimiter balancing and operator parsing.
- Enforce maximum clause count, nesting depth, list length, page size, and
  input length.
- Tag an initial pre-release and state the compatibility policy.

## Suggested release boundary

A realistic first usable release is **MySQL-only**:

- struct-derived schemas;
- the portable CQL profile in [`cql-rfc.md`](cql-rfc.md);
- page-based pagination;
- one stable public compile API returning fragments and arguments;
- passing parser-to-MySQL integration tests;
- one runnable example.

JSON schemas, cursor pagination, and Mongo/Elastic/Redis/GraphQL adapters SHOULD
be deferred unless they are required for that release. Keeping unfinished
methods in the public API makes the product appear more complete while making
failures harder for users to diagnose.

## Reassessment rule

Recalculate these percentages after each milestone using the same weights.
Do not raise a category merely because code exists; raise it when the behavior
is documented, reachable through the supported public API, and protected by
tests.
