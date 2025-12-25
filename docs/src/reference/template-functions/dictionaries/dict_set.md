# `dict_set`

The `dict_set` function sets or updates a key-value pair in a dictionary and returns the modified dictionary.

## Syntax

```
{{ $dict := dict_set dictionary "key" value }}
```

## Example

```
{{ $settings := dict_create "volume" 50 "muted" false }}
{{ $settings = dict_set $settings "volume" 75 }}
{{ $settings = dict_set $settings "equalizer" "on" }}

Volume: {{ dict_get $settings "volume" }}
Muted: {{ dict_get $settings "muted" }}
Equalizer: {{ dict_get $settings "equalizer" }}
```

Output:
```
Volume: 75
Muted: false
Equalizer: on
```

## Notes

- If the key already exists, its value is updated
- If the key doesn't exist, it's added to the dictionary
- Returns the modified dictionary (allows chaining)
- Modifies the original dictionary in place
