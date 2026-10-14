# Charcoal Query Language

Status: Draft

Version: 0.1

Audience: API consumers and service authors

## Abstract

Charcoal Query Language (CQL) expresses filtering, sorting, and pagination in
an HTTP URL query string. A service defines the fields clients may query and
their types. Charcoal validates a request against that field set and produces
a backend-independent token tree.

This document defines the draft language. The implementation is expected to
converge on this specification; known differences are tracked separately.

## Conformance language

The words **MUST**, **MUST NOT**, **SHOULD**, and **MAY** describe requirements
for clients and servers that claim conformance with this draft.

## Query parameters

CQL uses these URL query parameters:

- `filters`: the filter expression.
- `sort`: a comma-separated list of sort clauses.
- `per_page`: the requested page size.
- `page`: the one-based page number.

A request MAY omit any parameter. A client MUST NOT repeat a parameter.

The examples in this document show the decoded query for readability:

```text
filters=author = 'Jim Butcher'&sort=published_at:desc
```

When constructing a URL, clients MUST percent-encode reserved characters.
For example:

```text
filters=author+%3D+%27Jim+Butcher%27&sort=published_at%3Adesc
```

In URL query strings, `+` decodes to a space. A literal plus sign therefore
MUST be encoded as `%2B`.

## Data model and field names

The server supplies a field registry before parsing a request. Each entry has
a name and one of four types:

- string
- number
- boolean
- bytes

The registry is the allowlist for both filtering and sorting. An unknown field
is an error.

Services MUST publish their exact queryable field names. The same field-name
normalization and lookup rules MUST apply to filtering and sorting. Services
SHOULD use lowercase external field names for portability.

## Filter expressions

### Recommended grammar

The following EBNF defines the stable client-facing subset:

```ebnf
filter       = expression, { ",", expression } ;
expression   = clause | group ;
group        = "(", expression, join, expression,
               { join, expression }, ")" ;
join         = " AND " | " OR " ;
clause       = field, whitespace, operator,
               [ whitespace, value ] ;
field        = identifier, { identifier | "." } ;
operator     = equality-operator
             | comparison-operator
             | membership-operator
             | pattern-operator
             | null-operator ;
value        = bare-value | quoted-value | list-value ;
list-value   = "[", value, { ",", value }, "]" ;
```

An outer comma is an `AND`. For example:

```text
filters=name = bob, age >= 30
```

means:

```text
name = bob AND age >= 30
```

Whitespace around commas is optional.

### Clauses

A normal clause contains exactly one field, one operator, and one value:

```text
age >= 30
name = bob
occupation = 'delivery driver'
```

Null operators do not take a value:

```text
deleted_at is null
email not null
```

Membership operators take one bracket-delimited list:

```text
status in [active, pending]
role nin [guest, blocked]
```

A membership list MUST contain at least one value.

### Operators

Each operator has at most one symbolic spelling and one textual spelling.
Operators are case-insensitive after normalization.

- Equal: `=` or `eq`
- Not equal: `!=` or `ne`
- Greater than: `>` or `gt`
- Greater than or equal: `>=` or `gte`
- Less than: `<` or `lt`
- Less than or equal: `<=` or `lte`
- Membership: `in`
- Negative membership: `!in` or `nin`
- Pattern match: `~` or `like`
- Negative pattern match: `!~` or `!like`
- Null: `is null`
- Not null: `is !null` or `not null`

`!` is the preferred symbolic negation prefix. Membership has no separate
positive symbol because `in` is already concise and URL-safe. Null checks are
the exception to the general two-spelling rule: the `is` prefix keeps their
meaning explicit while `not null` remains the readable textual form.

The meaning of `like` is adapter-specific. The MySQL adapter maps it to SQL
`LIKE` and does not add wildcard characters.

### Values and quoting

Outer matching single or double quotes are removed from scalar values.
Escaped single and double quote sequences (`\'` and `\"`) are unescaped.
Leading and trailing whitespace is removed.

Clients MAY use either matching single or double quotes for values containing
spaces:

```text
occupation = 'delivery driver'
occupation = "delivery driver"
```

Clients SHOULD separate fields, operators, and values with whitespace:

```text
age >= 30
```

List elements are separated by commas. Quoted commas inside a list are not a
supported escape mechanism; services SHOULD treat list elements as simple
scalar values.

### Type validation

The parser validates values against the server's field registry:

- strings accept any parsed scalar value;
- numbers must parse as floating-point numbers;
- booleans use Go boolean syntax, including `true` and `false`;
- bytes currently accept any value.

Every member of an `in` or `nin` list is checked against the field type.
CQL does not currently have date, time, UUID, enum, or null literal types.
Services needing those types must model them as strings or perform additional
validation after parsing.

## Boolean expressions and grouping

`AND` and `OR` keywords are case-insensitive and MUST be separated from
neighboring text by spaces. Parentheses create nested groups:

```text
filters=name = bob, (age > 30 OR (occupation = driver AND hair_color = brown))
```

The top-level comma joins the two outer expressions with `AND`. Inside a
group, all separators at one nesting level SHOULD be the same:

```text
(status = active OR status = pending)
(age >= 18 AND verified = true)
```

Clients MUST use nested parentheses instead of mixing `AND`, `OR`, and commas
at the same level:

```text
((status = active OR status = pending) AND verified = true)
```

The current parser accepts mixed separators, but the last separator it sees
selects the join operation for the whole group. That behavior is ambiguous and
is not part of this specification.

The parser also accepts a bare boolean expression without outer parentheses:

```text
name = bob OR name = robert
```

Clients SHOULD include parentheses. They make the relationship with outer
comma-separated expressions explicit and remain compatible with the older
CQL documentation.

CQL does not define conventional `AND`-before-`OR` precedence. Parentheses
MUST be used whenever both operations are needed.

## Sorting

The `sort` parameter is a comma-separated sequence of `field:direction`
clauses:

```text
sort=published_at:desc,title:asc
```

Accepted directions are:

- ascending: `asc`, `ascending`
- descending: `desc`, `descending`

Directions are case-insensitive. Field names follow the same normalization and
lookup rules as filter field names. Each sort clause must contain exactly one
colon.

The order of clauses is significant: earlier clauses have higher sort
priority.

## Pagination

The presence of `page` or `per_page` requests pagination:

```text
per_page=20&page=2
```

No separate pagination boolean is used. When neither parameter is present,
the backend implementation decides whether and how to apply a default result
limit. When either parameter is present:

- `per_page`, if present, MUST be a positive integer;
- `page`, if present, MUST be a positive, one-based integer;
- an omitted parameter uses the backend implementation's configured default.

For example, `page=3` uses the configured default page size, while
`per_page=50` requests the first page unless the implementation configures a
different default page.

Implementations SHOULD define and enforce a maximum `per_page` value. Cursor
pagination is not part of CQL 0.1.

## Error behavior

Parsing is all-or-nothing. If any filter, sort, or pagination component is
invalid, Charcoal returns no token tree and an error rooted at:

```text
unable to parse query string
```

The parser attempts to collect independent errors before returning. A single
response can therefore report multiple unknown fields, invalid operators,
type mismatches, malformed groups, sort errors, and pagination errors.

Typical failures include:

```text
filters=height > 180
# field not found: height

filters=age > old
# type mismatch for field 'age'

sort=age:sideways
# invalid sort direction: 'sideways'
```

Unbalanced parentheses, brackets, or quotes invalidate the complete filter
parameter.

## Complete examples

Filter and sort:

```text
filters=author = 'Jim Butcher'&sort=published_at:desc
```

Membership and page pagination:

```text
filters=author eq 'Steven Erikson', title in [Midnight Tides, The God is not Willing]&per_page=10&page=1
```

Nested boolean logic:

```text
filters=author = Tolkien, (title = 'The Silmarillion' OR (publisher != Bantam AND title != Beowulf))&sort=rating:desc,price:asc
```

Null test:

```text
filters=deleted_at is null
```

## Security considerations

CQL is not SQL and clients cannot supply column names outside the server's
field registry. Adapters are responsible for mapping registered fields to
backend identifiers and binding values as parameters.

Applications MUST NOT concatenate CQL values into SQL or another backend
language. The included MySQL adapter uses placeholders for values, but table,
column, and join configuration remains trusted server-side input.

Services SHOULD enforce limits on URL length, group depth, number of clauses,
list length, page size, and query execution cost. The current parser does not
enforce those limits.

## Implementation status

Known differences between this specification and the current implementation
are tracked in [`cql-bugs.md`](cql-bugs.md). Those differences do not change
the CQL 0.1 language contract defined here.
