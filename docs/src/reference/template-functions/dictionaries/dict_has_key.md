# `dict_has_key`

The `dict_has_key` function checks whether a dictionary contains a specific key.

## Syntax

```
{{ dict_has_key dictionary "key" }}
```

## Example

```
{{ $data := dict_create "title" "Hello World" "author" "Alice" }}

{{ if dict_has_key $data "title" }}
  <h1>{{ dict_get $data "title" }}</h1>
{{ end }}

{{ if dict_has_key $data "subtitle" }}
  <h2>{{ dict_get $data "subtitle" }}</h2>
{{ else }}
  <h2>No subtitle</h2>
{{ end }}
```

Output:
```html
  <h1>Hello World</h1>

  <h2>No subtitle</h2>
```

## Notes

- Returns `true` if the key exists, `false` otherwise
- Useful for conditional rendering based on data availability
- The key's value doesn't matter—even empty string or `false` values return `true` if the key exists
