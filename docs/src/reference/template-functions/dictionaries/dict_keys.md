# `dict_keys`

The `dict_keys` function returns a list of all keys from one or more dictionaries.

## Syntax

```
{{ dict_keys dictionary1 dictionary2 ... }}
```

## Example

### Single dictionary

```
{{ $person := dict_create "name" "Bob" "age" 25 "city" "Oslo" }}
<ul>
{{ range dict_keys $person }}
  <li>{{ . }}</li>
{{ end }}
</ul>
```

Output:
```html
<ul>

  <li>name</li>

  <li>age</li>

  <li>city</li>

</ul>
```

### Multiple dictionaries

```
{{ $dict1 := dict_create "a" 1 "b" 2 }}
{{ $dict2 := dict_create "c" 3 "d" 4 }}
Keys: {{ join (dict_keys $dict1 $dict2) ", " }}
```

Output:
```
Keys: a, b, c, d
```

## Notes

- Returns keys in no particular order (dictionaries are unordered)
- When multiple dictionaries are provided, all keys from all dictionaries are returned
- Duplicate keys from different dictionaries will appear multiple times
