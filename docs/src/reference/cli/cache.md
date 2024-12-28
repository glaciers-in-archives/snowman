# `cache`

The `cache` command is a powerful tool used to manage the query cache of a Snowman project. By defualt, Snowman will cache SPARQL responses to avoid making the same query multiple times even across multiple builds. This command allows you to inspect and clear all or subsets of the cache.

```sh
snowman cache [optional-query] [optional-query-arguments]
```

## Flags

 - `--invalidate` (`-i`) Invalidates the cache for the given scope. If no query is given, the entire cache is invalidated. If a query but no query arguments are given, the cache for all queries using the given query file is invalidated.
 - `--unused` Sets the scope to all queries not used during the last build.
 - `--snowman-directory` (`-d`) The path to your Snowman directory. (default ".snowman")
 - `--help` (`-h`) Shows help for the command.
 - `--verbose` (`-v`) Enables verbose output.
 - `--timeit` (`-t`) Shows the time it took to execute the command.
