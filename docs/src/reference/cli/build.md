# `build`

The `build` command is used to build a Snowman project. It reads the configuration(`snowman.yaml`) file and the views file(`views.yaml`), then it fetches the data from a SPARQL endpoint, and finally it renders the templates. The output is written to the `site` directory.

```sh
snowman build
```

## Flags

 - `--cache` (`-c`) Sets the cache strategy. "available" will use cached SPARQL responses when available and fallback to making queries. "never" will ignore existing cache and will not update or set new cache. (default "available")
 - `--config` (`-f`) The path to the configuration file. (default "snowman.yaml")
 - `--help` (`-h`) Shows help for the command.
 - `--verbose` (`-v`) Enables verbose output.
 - `--static` (`-s`) Only update static files, do not fetch data or render templates.
 - `--timeit` (`-t`) Print the time it took to build the site.

