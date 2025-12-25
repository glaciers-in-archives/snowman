# `dict_unset`

The `dict_unset` function removes a key from a dictionary and returns the modified dictionary.

## Syntax

```
{{ $dict := dict_unset dictionary "key" }}
```

## Example

```
{{ $user := dict_create "username" "alice" "password" "secret123" "email" "alice@example.com" }}
{{ $user = dict_unset $user "password" }}

{{ range dict_keys $user }}
  {{ . }}
{{ end }}
```

Output:
```
username
email
```

## Notes

- If the key doesn't exist, the dictionary is unchanged
- Returns the modified dictionary
- Modifies the original dictionary in place
- Useful for removing sensitive data or temporary fields
