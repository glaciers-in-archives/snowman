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

## Version Requirements

You can specify a required Snowman version in your `snowman.yaml` to ensure all team members are using a compatible version:

```yaml
snowman_version: ">=0.7.0"
sparql_client:
  endpoint: "https://query.wikidata.org/sparql"
```

The `snowman_version` field uses NPM-style semantic versioning syntax and supports various constraint formats:

- **Range constraint**: `">=0.7.0"` - Requires version 0.7.0 or higher
- **Caret constraint**: `"^0.7.0"` - Compatible with 0.7.x (allows patch and minor updates)
- **Tilde constraint**: `"~0.7.0"` - Compatible with 0.7.x (allows patch updates only)
- **Exact version**: `"0.7.1"` - Requires exactly version 0.7.1
- **Complex range**: `">=0.7.0 <0.8.0"` - Between 0.7.0 (inclusive) and 0.8.0 (exclusive)

If the Snowman version doesn't satisfy the requirement, the build will fail with a clear error message:

```
Error: Version mismatch. Your Snowman version (0.6.5) does not satisfy the project requirement (>=0.7.0)
```

This field is optional. If omitted, any Snowman version can attempt to build the project.

## Rate limiting

SPARQL endpoints and other remote resources often limit how many requests they accept over a given period of time. Instead of failing your build the moment an endpoint asks Snowman to slow down, Snowman waits and retries the request.

Snowman treats HTTP 429 (Too Many Requests) responses, and HTTP 503 responses carrying a `Retry-After` header, as rate limiting. When the response tells Snowman how long to wait using `Retry-After` that delay is respected, otherwise Snowman waits `initial_backoff` and doubles the wait for each following retry.

The behaviour can be tuned using the optional `rate_limit` field:

```yaml
sparql_client:
  endpoint: "https://query.wikidata.org/sparql"
rate_limit:
  max_retries: 3
  initial_backoff: "1s"
  max_wait: "60s"
```

- `max_retries` The number of times a rate limited request is retried. Set it to `0` to disable retrying and fail the build on the first rate limiting response. (default `3`)
- `initial_backoff` How long to wait before the first retry when the response doesn't include a `Retry-After` header, the wait is doubled for each following retry. (default `"1s"`)
- `max_wait` The longest Snowman waits before a single retry. Backoffs are capped by this value, while a `Retry-After` asking for a longer delay makes Snowman give up as it can't respect it. (default `"60s"`)

Durations are written as Go duration strings such as `"500ms"`, `"30s"` or `"5m"`.

These settings apply to both SPARQL queries and to remote resources fetched with [`get_remote`](template-functions/http/get_remote.md) and [`get_remote_with_config`](template-functions/http/get_remote_with_config.md). Snowman prints a line each time it waits so that a slow build doesn't look like a stuck one:

```
Rate limited (HTTP 429) by query.wikidata.org, waiting 30s before retry 1 of 3.
```

Note that a build still fails once the retries have been exhausted. Endpoints that rate limit you aggressively are best combined with the [SPARQL cache](../guides/managing-cache.md), which lets a new build reuse the responses an earlier build already retrieved.

## Metadata

The `metadata` field allows you to define project-specific configuration values that can be accessed in your templates. This is useful for storing site-wide settings like site titles, URLs, author information, or any other data you want to reuse across your templates.

```yaml
metadata:
  site_title: "My Awesome Knowledge Base"
  base_url: "https://example.org"
  author: "Tux"
  contact_email: "tux@example.org"
  analytics_id: "123456-1"
  social_links:
    - name: "Codeberg"
      url: "https://codeberg.org/..."
    - name: "Mastodon"
      url: "https://social.example.org/tux"
```

### Accessing metadata in templates

Metadata values can be accessed in your templates using the [`config` template function](template-functions/utilities/config.md):

```html
<title>{{ config.Metadata.site_title }}</title>
<meta name="author" content="{{ config.Metadata.author }}">
<a href="mailto:{{ config.Metadata.contact_email }}">Contact</a>
```

For nested values like the social links example above, you can iterate over them:

```html
<ul>
{{ range config.Metadata.social_links }}
  <li><a href="{{ .url }}">{{ .name }}</a></li>
{{ end }}
</ul>
```
