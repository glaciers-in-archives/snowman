# `dict_pluck`

The `dict_pluck` function extracts values for a specific key from multiple dictionaries, returning them as a list.

## Syntax

```
{{ dict_pluck "key" dictionary1 dictionary2 ... }}
```

## Example

```
{{ $person1 := dict_create "name" "Alice" "age" 30 "city" "Stockholm" }}
{{ $person2 := dict_create "name" "Bob" "age" 25 "city" "Oslo" }}
{{ $person3 := dict_create "name" "Charlie" "age" 35 "city" "Copenhagen" }}

Names: {{ join (dict_pluck "name" $person1 $person2 $person3) ", " }}
Ages: {{ join (dict_pluck "age" $person1 $person2 $person3) ", " }}
```

Output:
```
Names: Alice, Bob, Charlie
Ages: 30, 25, 35
```

## Example with iteration

```
{{ $users := list }}
{{ $users = append $users (dict_create "id" 1 "username" "alice" "active" true) }}
{{ $users = append $users (dict_create "id" 2 "username" "bob" "active" false) }}
{{ $users = append $users (dict_create "id" 3 "username" "charlie" "active" true) }}

<ul>
{{ range dict_pluck "username" $users }}
  <li>{{ . }}</li>
{{ end }}
</ul>
```

## Notes

- The first argument is the key to extract
- Remaining arguments are the dictionaries to extract from
- Only includes values from dictionaries that have the specified key
- If a dictionary doesn't have the key, it's skipped (not included in the result)
- Useful for extracting a column of data from a list of dictionaries
