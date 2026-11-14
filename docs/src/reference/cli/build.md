# `build`

The `build` command is used to build a Snowman project. It reads the configuration(`snowman.yaml`) file and the views file(`views.yaml`), then it fetches the data from a SPARQL endpoint, and finally it renders the templates. The output is written to the `site` directory.

```sh
snowman build
```

## Partial builds

On large sites you don't always want to rebuild everything. To build only the pages of a single view, pass the view's `output` as it is written in `views.yaml`:

```sh
snowman build index.html
snowman build "entities/{{id}}.html"
```

For a view which outputs one file per SPARQL result, you can add the value of its path argument to build only that page. The following renders `site/entities/Q42.html` and nothing else:

```sh
snowman build "entities/{{id}}.html" Q42
```

A partial build leaves the rest of the `site` directory in place: other pages and static files are neither rebuilt nor removed. This also means that a page isn't removed when its row is no longer returned by the view's query; run a full build to clean up.

Partial builds use the cache in the same way as full builds, so with the default cache strategy a page is re-rendered from cached query results. To refresh the data of a page, clear the cache of the queries it uses before building it:

```sh
snowman cache sparql clear entity.rq Q42
snowman build "entities/{{id}}.html" Q42
```

## Flags

 - `--cache` (`-c`) Sets the cache strategy. "available" will use cached SPARQL responses when available and fallback to making queries. "never" will ignore existing cache and will not update or set new cache. (default "available")
 - `--config` (`-f`) The path to the configuration file. (default "snowman.yaml")
 - `--snowman-directory` (`-d`) The path to your Snowman directory. (default ".snowman")
 - `--help` (`-h`) Shows help for the command.
 - `--verbose` (`-v`) Enables verbose output.
 - `--static` (`-s`) Only update static files, do not fetch data or render templates.
 - `--timeit` (`-t`) Print the time it took to build the site.

