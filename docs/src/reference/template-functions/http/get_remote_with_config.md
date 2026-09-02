# `get_remote_with_config`

The `get_remote_with_config` function fetches the content of a remote URL and returns it as a string. It takes a configuration object as an argument, which can be used to set custom HTTP headers.

```snowman
{{ get_remote_with_config "https://example.com" $config }}
```

Responses telling Snowman that it's being rate limited are retried, see [rate limiting](../../snowman-yaml.md#rate-limiting).
