# `dict_get`

The `dict_get` function retrieves a value from a dictionary by its key.

## Syntax

```
{{ dict_get dictionary "key" }}
```

## Example

```
{{ $config := dict_create "theme" "dark" "language" "en" "timezone" "UTC" }}
Theme: {{ dict_get $config "theme" }}
Language: {{ dict_get $config "language" }}
Missing: {{ dict_get $config "nonexistent" }}
```

Output:
```
Theme: dark
Language: en
Missing:
```

## Notes

- Returns an empty string if the key doesn't exist
- Use [`dict_has_key`](dict_has_key.md) to check for key existence before retrieving
- For nested dictionaries, use [`dict_find`](dict_find.md) instead
