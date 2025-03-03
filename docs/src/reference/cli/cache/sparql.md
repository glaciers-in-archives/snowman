# `sparql`

The `sparql` command is used to inspect and clear the SPARQL cache.

## `sparql inspect`

The `inspect` subcommand is used to inspect the SPARQL cache. It allows you to view the contents of the cache.

Its first argument is the file name or path of the query you want to inspect. 

```sh
snowman cache sparql inspect my-query.rq
```

If your query is parameterized, you can pass the parameter values as additional arguments. Passing a parameterized query without parameters will yeild the number of items for that query.

```sh
snowman cache sparql inspect my-parameterized-query.rq my-param-value
```

You can select only unused queries by using the `--unused` flag. When used, the command does not expect any arguments.

```sh
snowman cache sparql inspect --unused
```

## `sparql clear`

The `clear` subcommand is used to clear the SPARQL cache. It allows you to clear the cache for a specific query, all queries, only unused queries, or a specific parameterized query.

It follows the same syntax as the `inspect` subcommand.

Clear the cache for a query.

```sh
snowman cache sparql clear my-query.rq
```

Clear a specific parameterized query.

```sh
snowman cache sparql clear my-parameterized-query.rq my-param-value
```

Clear the cache for all queries.

```sh
snowman cache sparql clear
```

Clear the cache for only unused queries.

```sh
snowman cache sparql clear --unused
```

## Flags

 - `--unused` (`-u`) Selects only unused queries. Used without arguments.
 - `--snowman-directory` (`-d`) The path to your Snowman directory. (default ".snowman")
 - `--help` (`-h`) Shows help for the command.
 - `--verbose` (`-v`) Enables verbose output.
 - `--timeit` (`-t`) Print the time it took to build the site.

