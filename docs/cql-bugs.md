# CQL Implementation Issues

Status: Draft issue backlog

This document records differences between the CQL specification and the
current implementation. Each section is structured so it can become a GitHub
issue with minimal editing.

The language contract is defined by [`cql-rfc.md`](cql-rfc.md). Entries here
describe implementation work, not alternative language semantics.

## Parse compact operators deterministically

Current behavior:

Compact clauses scan `filter.OperatorMap` using Go map iteration. When one
operator prefixes another, such as `>` and `>=`, either match can win.

Expected behavior:

Operator recognition MUST be deterministic. Tokenization or longest-match
ordering should produce the same result on every run. Spaced and compact forms
should share the same recognition rules.

Evidence:

- [`internal/query/parse.go`](../internal/query/parse.go)
- [`internal/filter/operators.go`](../internal/filter/operators.go)

## Preserve double-quoted values in spaced clauses

Current behavior:

The outer filter splitter recognizes single and double quotes, but
`splitFilterExpression` toggles quote state only for single quotes. A clause
such as `occupation = "delivery driver"` can split the value at its space.

Expected behavior:

Single-quoted and double-quoted values MUST follow the same tokenization and
escaping rules at every parser stage.

Evidence:

- [`internal/query/parse.go`](../internal/query/parse.go)
- [`internal/query/split.go`](../internal/query/split.go)

## Reject mixed boolean separators at one level

Current behavior:

`splitTopLevel` accepts commas, `AND`, and `OR` at the same nesting level. The
last separator encountered selects the join operation for the entire group.

Expected behavior:

A group that mixes boolean separators at one level MUST return an actionable
parse error. Nested parentheses are required when both `AND` and `OR` are
needed.

Evidence:

- [`internal/query/split.go`](../internal/query/split.go)
- [`internal/query/parse.go`](../internal/query/parse.go)

## Normalize filter and sort field names consistently

Current behavior:

Filter fields are trimmed and lowercased. Sort fields are trimmed but retain
their original case, so the same field spelling can succeed in `filters` and
fail in `sort`.

Expected behavior:

Filtering and sorting MUST use one field-name normalization and lookup rule.
That rule should also agree with struct-derived field names.

Evidence:

- [`internal/query/parse.go`](../internal/query/parse.go)
- [`internal/query/sort.go`](../internal/query/sort.go)
- [`internal/filter/struct.go`](../internal/filter/struct.go)

## Preserve membership lists across parser and MySQL adapter

Current behavior:

The parser stores an `in` or `nin` value with its surrounding brackets. The
MySQL adapter splits that raw string on commas without removing the brackets.
For example, `[active, pending]` can bind as `"[active"` and `"pending]"`.

Expected behavior:

Membership values SHOULD be represented as structured token values. At
minimum, the parser and adapter MUST agree on one normalized representation,
and brackets must never enter bound values.

Evidence:

- [`internal/query/parse.go`](../internal/query/parse.go)
- [`adapters/sql/build.go`](../adapters/sql/build.go)
- [`internal/query/parse_test.go`](../internal/query/parse_test.go)
- [`adapters/sql/build_test.go`](../adapters/sql/build_test.go)

## Remove or reject the unfinished cursor parameter

Current behavior:

The parser accepts and stores `cursor`, but the MySQL adapter ignores it.

Expected behavior:

Because cursor pagination is not part of CQL 0.1, the parser SHOULD reject the
parameter or the public API SHOULD stop advertising and storing it. Cursor
support can be introduced later with defined ordering, encoding, and adapter
semantics.

Evidence:

- [`internal/query/parse.go`](../internal/query/parse.go)
- [`internal/query/pagination.go`](../internal/query/pagination.go)
- [`internal/tokens/pagination.go`](../internal/tokens/pagination.go)
- [`adapters/sql/build.go`](../adapters/sql/build.go)

## Align struct field behavior with documentation

Current behavior:

The root README promises snake_case conversion and struct tags for skipping
fields. Struct reflection uses Go field names and does not inspect those tags.

Expected behavior:

Choose and implement a canonical external field-name rule plus documented
rename and skip tags. Update the README and tests to match the chosen
contract.

Evidence:

- [`README.md`](../README.md)
- [`internal/filter/struct.go`](../internal/filter/struct.go)
- [`example/simple/model.go`](../example/simple/model.go)

## Reconcile bare boolean groups in the older query guide

Current behavior:

The older query guide says `OR` must be wrapped in parentheses. The parser and
the CQL specification accept a bare boolean expression, although the
specification recommends parentheses for clarity.

Expected behavior:

Update or replace the older guide so it does not contradict the normative CQL
document.

Evidence:

- [`internal/query/README.md`](../internal/query/README.md)
- [`internal/query/parse.go`](../internal/query/parse.go)
- [`docs/cql-rfc.md`](cql-rfc.md)

## Implement the canonical negative and null spellings

Current behavior:

The parser recognizes `nin` but does not recognize the canonical symbolic
negative-membership operator `!in`. It also does not recognize the canonical
negative-null spelling `is !null`.

Expected behavior:

The parser MUST recognize the operator spellings specified by CQL 0.1:
`!in`/`nin`, `!~`/`!like`, `is null`, and `is !null`/`not null`.

Evidence:

- [`internal/filter/operators.go`](../internal/filter/operators.go)
- [`internal/query/parse.go`](../internal/query/parse.go)
- [`docs/cql-rfc.md`](cql-rfc.md)

## Infer pagination from page parameters

Current behavior:

The parser processes `page` and `per_page` only when a nonempty `pagination`
parameter is also present. A request such as `page=2&per_page=20` produces no
pagination token.

Expected behavior:

The presence of `page` or `per_page` MUST enable pagination. The
`pagination` boolean should no longer be required or exposed by the CQL
contract.

Evidence:

- [`internal/query/parse.go`](../internal/query/parse.go)
- [`internal/query/pagination.go`](../internal/query/pagination.go)
- [`internal/tokens/pagination.go`](../internal/tokens/pagination.go)

## Validate page and page-size ranges

Current behavior:

`page` and `per_page` must parse as integers, but zero and negative values are
accepted. The MySQL adapter silently emits no limit for a nonpositive page
size and treats a page below two as the first page.

Expected behavior:

Both parameters MUST be positive when supplied. Implementations SHOULD enforce
a configurable maximum `per_page` value and return a validation error rather
than silently changing invalid input.

Evidence:

- [`internal/query/pagination.go`](../internal/query/pagination.go)
- [`adapters/sql/build.go`](../adapters/sql/build.go)
