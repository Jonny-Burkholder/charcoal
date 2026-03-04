# charcoal
Charcoal filters better

## Why use charcoal?

Building query logic is a drag, for a lot of reasons. There's so many ways to approach it that it can become paralysis by analysis. Many of the existing solutions provide lots of querying power, but are either not RESTful, are unreadable by users, or both.

That's where charcoal comes in.

Charcoal combines RESTful, human-readable Charcoal Query Language with powerful, database-agnostic tokenization that can be added to virtually any data backend by swapping in the appropriate integration. And if the integration doesn't exist, the user can easily create their own code to parse the tokens into whatever data backend they are using.

And on the backend, charcoal replaces the tedium of manually crafting filters by reducing it to a single function call. Simply pass in your struct or JSON, and charcoal parses it into a filter object that can ingest the user query and validate that it fits with the data you've provided.

## How To Use Charcoal

### CQL - Charcoal Query Language

This is the first step to the magic of charcoal filtering: a user-friendly query language that is easy for humans to read, and doesn't require putting an http body into a GET request (yuck!). The basics of it go like this: different clauses are comma separated. If things need to go together, they go in parentheses. Most operators that you're familiar with are allowed, and sorting and pagination are also supported. This is easier to see by looking at a query. Say you have a database of books for users to access, and they want to get a list of books by a certain author and sort them from most recent to oldest. That query would look something like this:

`my-api.com/books?filters=author="Jim Butcher"&sort=date:desc`

Prety simple, right? If the user is the sort that likes whitespace, that's supported too:

`my-api.com/books?filters=author eq "Steven Erikson", title in ["Midnight Tides", "The God is not Willing"]&pagination=true&per_page=10`

Notice how different forms of operators are allowed as well? We can do more complex things too:

`?filters=author eq Tolkien, title = "The Silmarillion" OR (published_by != Bantam, Title != Beowulf)&sort=rating:desc,price:asc`

It's pretty easy to see what this does and how it works, and I'll be posting more documentation soon on using the query language itself. But we're about to get into where the *real* magic happens.

### The backend

This is where it gets really fun for you as a developer. No more fiddling with manual filters, having to comb through bugs and breakages any time your data layer changes. Charcoal handles all of that for you. Here's how you'd create filters using charcoal:

```
type User struct {
  Name string
  Age uint
  Password string `charcoal:skip`
  Address address
}

type address struct {
  LineOne string
  Zip uint
}

// create the filters with a single function call:
filters, _ := charcoal.New(User{})
```

That's really it! All you need to do. Charcoal will automatically go through and create filters, converting field names to snake_case (unless otherwise specified in the config), and using dot notation for nested fields. Then we can filter:

```
query := "filters=name = Bob, occupation = Tomato"

result := filter.Activate(query)
```

You may have noticed we used a field that wasn't present in the original struct. Not to worry, charcoal will give us a very specific error telling the user exactly how to fix the mistake:

```
unable to parse query string
invalid filter expression
field not found: occupation
```

Just like that, with only a few lines of code and no manual filter work, your user knows exactly what they can and cannot filter by. Charcoal will do this for every error it finds in a single query, so that the user doesn't have to send different versions of the same request multiple times to the API, tweaking it a little bit each time. Instead, they can fix every error all at once. Pretty neat!

### Integrations

Now that you have the filters, and the user query, all you have to do is convert it into a form that your preferred data backend can read. This can be done through one of the several built-in integrations, or you can design your own to suit your needs. Here's how the built-in integrations work:

`sql, err := filter.Activate(query).ToMySQL()`

The sql will be a string that you can pass right to your mysql repository functions. For custom integrations, you'll simply pass in your result from the `filter.Activate` function into the custom integration:

```
import (
  "github.com/Jonny-Burkholder/charcoal"
  "github.com/my-custom-charcoal-integrations/obscuredb"
)

filter, _ := charcoal.New(myDataType)
integration := obscuredb.New()

result, _ := filter.Activate(userQuery)

transaction, _ := integration.ToObscureDBTransaction(result)

```


So there you have it! Filtering, made easy, for everyone.

This repo is still a work in progress, but I'm working hard to get it ready for prime-time.
