# `dict_omit`

The `dict_omit` function creates a new dictionary excluding the specified keys from the original dictionary.

## Syntax

```
{{ dict_omit dictionary "key1" "key2" ... }}
```

## Example

```
{{ $response := dict_create "id" 123 "name" "Product" "price" 29.99 "internal_code" "X-2025" "warehouse_id" 5 }}
{{ $apiResponse := dict_omit $response "internal_code" "warehouse_id" }}

API Response:
{{ range $key, $value := $apiResponse }}
  {{ $key }}: {{ $value }}
{{ end }}
```

Output:
```
API Response:

  id: 123

  name: Product

  price: 29.99

```

## Notes

- Creates a new dictionary; the original is unchanged
- If a specified key doesn't exist in the original dictionary, it's silently ignored
- Useful for removing internal fields, sensitive data, or temporary values
- The opposite of [`dict_pick`](dict_pick.md)
