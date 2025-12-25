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
