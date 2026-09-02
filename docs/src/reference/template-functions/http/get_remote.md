# `get_remote`

The `get_remote` function fetches the content of a remote URL and returns it as a string.

```
{{ get_remote "https://example.com" }}
```

Requests can be slowed down per URL pattern through `url_rules`, and responses telling Snowman that it's being rate limited are retried, see the [configuration reference](../../configuration.md#url_rules).
