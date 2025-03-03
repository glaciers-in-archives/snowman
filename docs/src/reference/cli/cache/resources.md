# `resources`

The `resources` subcommand of the `cache` command is used to manage the cache store for non-SPARQL HTTP requests.

## `resources inspect`

The `inspect` subcommand is used to inspect the resources cache. It allows you to view the contents of the cache by passing the URL of the resource you want to inspect.

```sh
snowman cache resources inspect https://example.com/resource
```

You can also list the number of items in the cache by not passing any arguments.

```sh
snowman cache resources inspect
```

In addition, you can select only unused resources by using the `--unused` flag. When used, the command does not expect any arguments.

```sh
snowman cache resources inspect --unused
```

## `resources clear`

The `clear` subcommand is used to clear the resources cache. It allows you to clear the cache for a specific resource, all resources, or only unused resources.

It follows the same syntax as the `inspect` subcommand.

Clear the cache for a resource.

```sh
snowman cache resources clear https://example.com/resource
```

Clear the cache for all resources.

```sh
snowman cache resources clear
```

Clear the cache for only unused resources.

```sh
snowman cache resources clear --unused
```

## Flags

 - `--unused` (`-u`) Selects only unused queries. Used without arguments.
 - `--snowman-directory` (`-d`) The path to your Snowman directory. (default ".snowman")
 - `--help` (`-h`) Shows help for the command.
 - `--verbose` (`-v`) Enables verbose output.
 - `--timeit` (`-t`) Print the time it took to build the site.