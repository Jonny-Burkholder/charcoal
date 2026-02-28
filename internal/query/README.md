# How the Charcoal Query Language works

Charcoal supports a wide arrange of operators and complexity, and is mean to be a RESTful implementation of filters with the power of a system like GraphQL. Here are a few examples of how the language works:

Query fields are based on commas for separation. This is comparable to a MySQL "AND" statement, though charcoal operators don't couple tightly to any one data backend. Take this filter string for instance:

`filters=name eq bob, age = 30, occupation = "driver"`

this is roughly equivalent in mysql to

`name = bob AND age = 30 AND occupation = driver`

Notice that we can use both "=" and "eq" for the operator, and we can choose whether to wrap values in quotes. However, if a value contains spaces or operator characters, it must be wrapped in quotes:

`occupation = "uber driver"`

OR is more explicit:

`(name = bob OR occupation = driver)`

because of the relative ambiguity of OR statements, they must be wrapped in parenthesis to avoid confusion.

Complex queries can also be nested:

`name = bob, (age > 30 OR (occupation = driver, hair_color = brown))`


