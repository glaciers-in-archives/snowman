# Configuration

Snowman is configured through a `snowman.yaml` file in the project's root
directory. An alternative location can be passed to `snowman build` with the
`--config`/`-f` flag.

## `sparql_client`

Configures the connection to the SPARQL endpoint.

- `endpoint` — the URL of the SPARQL endpoint.
- `http_headers` — optional map of HTTP headers sent with every SPARQL request.
- `rate_limit` — optional minimum interval between SPARQL requests, expressed
  as a [Go duration string](https://pkg.go.dev/time#ParseDuration)
  (e.g. `500ms`, `1s`, `2s`). When set, Snowman waits at least this long
  between consecutive requests to the endpoint. Cached responses are served
  without waiting.

```yaml
sparql_client:
  endpoint: "https://query.wikidata.org/sparql"
  http_headers:
    User-Agent: "example Snowman (https://github.com/glaciers-in-archives/snowman)"
  rate_limit: "1s"
```

## `http_rate_limits`

Optional map configuring rate limiting for non-SPARQL HTTP requests made by
the `get_remote` and `get_remote_with_config` template functions. Each key is
a host name and each value is a minimum interval between requests to that host,
expressed as a [Go duration string](https://pkg.go.dev/time#ParseDuration).

Requests to hosts not listed here are not rate limited, and cached resources
are served without waiting.

```yaml
http_rate_limits:
  "api.example.com": "500ms"
  "images.example.org": "2s"
```

## `snowman_version`

Optional [semver constraint](https://github.com/Masterminds/semver#checking-version-constraints)
specifying which Snowman versions the project supports. The build fails if the
running Snowman version does not satisfy the constraint.

```yaml
snowman_version: ">=0.7.0"
```

## `metadata`

Optional map of arbitrary values made available to templates through the
`metadata` template function.

```yaml
metadata:
  a_config_key: "a config value"
```
