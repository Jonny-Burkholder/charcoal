# Design Principles

These are, for now, in no particular order. Will organize later

 - Integrations/adapters should be able to operate based on publicly exposed Charcoal functionality. In other words, anyone should be able to build an integration inside or outside of the project

  - No recursion

  - There will be several methods for generating filters. Some of these will probably best be served by an intermediate adapter. For now, the two main ways of generating filters will be parsing a go struct directly or parsing a JSON payload

  - Users can define an action to take on a filter, but there should be a default action for each backend

  - Switching out data backends should have minimal impact on Charcoal. Some special cases will need updating by the user - for instance, relational databases will sometimes require join statements, whereas Mongo will not. But in "default" cases, the user should be able to simply plug in a new backend and not notice the change

  - The token language should be data-store agnostic. If the data is in snowflake, salesforce, a REST API, MYSQL, NOSQL, or Excel, each integration should be able to take in the same tokens and perform the necessary functionality.

  - The adapters themselves must be responsible for understanding the data relationships of the data layer. If an SQL query requires a join statement, it is the responsibility of the adapter to know that column a and column b are in different tables, and relate based on foreign key ab.

  - There should be a "do once" quality to parsing the filters and the data relationship. A user should have the option of generating the data a single time at compile time or runtime, and either store it or embed it into the binary