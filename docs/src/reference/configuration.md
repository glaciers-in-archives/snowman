# Configuration

Snowman is configured through a `snowman.yaml` file in the project's root
directory. An alternative location can be passed to `snowman build` with the
`--config`/`-f` flag.

## `sparql_client`

Configures the connection to the SPARQL endpoint.

- `endpoint` — the URL of the SPARQL endpoint.
- `http_headers` — optional map of HTTP headers sent with every SPARQL request.

Requests to the endpoint can be slowed down through [`url_rules`](#url_rules),
which is keyed by URL patterns matching the endpoint.

```yaml
sparql_client:
  endpoint: "https://query.wikidata.org/sparql"
  http_headers:
    User-Agent: "example Snowman (https://github.com/glaciers-in-archives/snowman)"
```

## `url_rules`

Optional map of settings applied to the requests Snowman makes, both the SPARQL
queries sent to `sparql_client.endpoint` and the resources fetched by the
`get_remote` and `get_remote_with_config` template functions.

Each key is a URL pattern matched against the host name and path of a request,
where `*` matches any sequence of characters and `?` a single one. A pattern
without a path, such as `query.wikidata.org` or `*.wikimedia.org`, matches
every URL on the hosts it names. Neither the scheme nor the port take part in
the matching, and patterns are matched case insensitively.

- `rate_limit` — a minimum interval between the matching requests, expressed as
  a [Go duration string](https://pkg.go.dev/time#ParseDuration) (e.g. `500ms`,
  `1s`, `2s`). Requests that no pattern matches are not slowed down, and cached
  responses are served without waiting.

```yaml
sparql_client:
  endpoint: "https://query.wikidata.org/sparql"
url_rules:
  "query.wikidata.org":
    rate_limit: "1s"
  "query.wikidata.org/sparql":
    rate_limit: "100ms"
  "*.wikimedia.org":
    rate_limit: "500ms"
  "*":
    rate_limit: "2s"
```

Requests matching the same pattern share its rate limit, so `*.wikimedia.org`
above limits the requests made to the service as a whole rather than to each of
its hosts. Where several patterns match a request only the most specific one
applies: a pattern without wildcards wins over any pattern with them, and a
longer pattern wins over a shorter one. The example above allows a SPARQL query
every 100ms, any other request to `query.wikidata.org` every second, a request
to any other Wikimedia host every 500ms, and anything else every two seconds.

Should a server rate limit Snowman anyway, by answering HTTP 429 (Too Many
Requests) or HTTP 503 with a `Retry-After` header, Snowman waits and retries
the request instead of failing the build. That happens on its own and needs no
configuration, it respects the delay the server asks for and otherwise waits a
little longer for each retry:

```
Rate limited (HTTP 429) by query.wikidata.org, waiting 30s before retry 1 of 3.
```

A build still fails if a server keeps rate limiting Snowman after a few
retries. Giving the matching URLs a `rate_limit` is the way to keep that from
happening again.

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
