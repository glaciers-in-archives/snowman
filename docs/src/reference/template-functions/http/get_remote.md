# `get_remote`

The `get_remote` function fetches the content of a remote URL and returns it as a string.

```
{{ get_remote "https://example.com" }}
```

Responses telling Snowman that it's being rate limited are retried, see [rate limiting](../../snowman-yaml.md#rate-limiting).
