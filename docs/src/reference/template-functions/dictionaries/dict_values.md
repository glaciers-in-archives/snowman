# `dict_values`

The `dict_values` function returns a list of all values from a dictionary.

## Syntax

```
{{ dict_values dictionary }}
```

## Example

```
{{ $scores := dict_create "math" 95 "science" 87 "history" 92 }}
<ul>
{{ range dict_values $scores }}
  <li>Score: {{ . }}</li>
{{ end }}
</ul>
```

Output:
```html
<ul>

  <li>Score: 95</li>

  <li>Score: 87</li>

  <li>Score: 92</li>

</ul>
```

## Notes

- Returns values in no particular order (dictionaries are unordered)
- The returned list contains only the values, not the keys
- Useful for iterating over values when keys aren't needed
