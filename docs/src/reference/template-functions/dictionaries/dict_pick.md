# `dict_pick`

The `dict_pick` function creates a new dictionary containing only the specified keys from the original dictionary.

## Syntax

```
{{ dict_pick dictionary "key1" "key2" ... }}
```

## Example

```
{{ $user := dict_create "username" "alice" "password" "secret" "email" "alice@example.com" "role" "admin" "lastLogin" "2025-12-25" }}
{{ $publicData := dict_pick $user "username" "email" "role" }}

Public user data:
{{ range $key, $value := $publicData }}
  {{ $key }}: {{ $value }}
{{ end }}
```

Output:
```
Public user data:

  username: alice

  email: alice@example.com

  role: admin

```

## Notes

- Creates a new dictionary; the original is unchanged
- If a specified key doesn't exist in the original dictionary, it's silently ignored
- Useful for filtering sensitive data or creating subsets of larger dictionaries
- The opposite of [`dict_omit`](dict_omit.md)
