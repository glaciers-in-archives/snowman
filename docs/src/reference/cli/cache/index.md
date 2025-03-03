# `cache`

The `cache` command is a powerful tool used to manage the two different cache stores held by Snowman. One for HTTP requests made by the SPARQL engine and one for non-SPARQL HTTP requests.  Snowman will by default cache HTTP responses to avoid making the same request multiple times even across multiple builds.

The `cache` command has two subcommands: `sparql` and `resources`.