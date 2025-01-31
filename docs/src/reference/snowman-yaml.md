# `snowman.yaml`

`snowman.yaml` contains the core configuartion needed to build your project such as the SPARQL endpoint you are targeting and project specific metadata.

A typical `snowman.yaml` looks like this, however only the SPARQL client and its endpoint is mandatory:

```yaml
sparql_client:
  endpoint: "https://query.wikidata.org/sparql"
  http_headers:
    User-Agent: "example Snowman (https://github.com/glaciers-in-archives/snowman)"
metadata:
  production_setting: "a config value"
```

Note that while Snowman will look for `snowman.yaml` by default you can point to other files when building your project:

```bash
snowman build --config=production-snowman.yaml
```

This is useful if you need to build your proejct in various environments such as development, CI, and production.

