# Working with multiple environments

If you need different snowman.yaml configurations for different environments you can use the `config` build flag to build your project using configurations other than the default `snowman.yaml`.

```bash
snowman build --config=production-snowman.yaml
```

You can also read environment variables using the built-in template function `env`.

