# `dict_find`

The `dict_find` function safely navigates nested dictionaries, returning a default value if any key in the path doesn't exist.

## Syntax

```
{{ dict_find "key1" "key2" ... "keyN" defaultValue dictionary }}
```

## Example

### Basic nested access

```
{{ $config := dict_create "database" (dict_create "host" "localhost" "port" 5432) }}

Host: {{ dict_find "database" "host" "unknown" $config }}
Port: {{ dict_find "database" "port" 0 $config }}
User: {{ dict_find "database" "user" "default_user" $config }}
```

Output:
```
Host: localhost
Port: 5432
User: default_user
```

### Deeply nested structure

```
{{ $data := dict_create "user" (dict_create "profile" (dict_create "settings" (dict_create "theme" "dark"))) }}

Theme: {{ dict_find "user" "profile" "settings" "theme" "light" $data }}
Lang: {{ dict_find "user" "profile" "settings" "language" "en" $data }}
```

Output:
```
Theme: dark
Lang: en
```

## Notes

- Arguments are: path keys, default value, then the dictionary
- The last two arguments are always the default value and the dictionary
- All preceding arguments are the path of keys to traverse
- Returns the default value if any key in the path doesn't exist
- Safer than chained `dict_get` calls as it won't error on missing keys
- Particularly useful when working with data from external sources where structure isn't guaranteed
