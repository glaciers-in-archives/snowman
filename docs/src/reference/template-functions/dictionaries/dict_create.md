# `dict_create`

The `dict_create` function creates a new dictionary from key-value pairs.

## Syntax

```
{{ dict_create "key1" "value1" "key2" "value2" ... }}
```

## Example

```
{{ $person := dict_create "name" "Alice" "age" 30 "city" "Stockholm" }}
<p>Name: {{ dict_get $person "name" }}</p>
<p>Age: {{ dict_get $person "age" }}</p>
<p>City: {{ dict_get $person "city" }}</p>
```

Output:
```html
<p>Name: Alice</p>
<p>Age: 30</p>
<p>City: Stockholm</p>
```

## Notes

- Keys must be strings
- Values can be any type (strings, numbers, booleans, arrays, or nested dictionaries)
- If an odd number of arguments is provided, the last key will have an empty string value
