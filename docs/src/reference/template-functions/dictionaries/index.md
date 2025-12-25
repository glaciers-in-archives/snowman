# Dictionaries

Snowman provides a set of functions for working with dictionaries (also known as maps or associative arrays). These functions allow you to create, manipulate, and query dictionary data structures in your templates.

Dictionaries are useful for organizing related data, building dynamic structures, and transforming data during template rendering.

## Available Functions

- [`dict_create`](dict_create.md) - Create a new dictionary from key-value pairs
- [`dict_get`](dict_get.md) - Get a value from a dictionary by key
- [`dict_set`](dict_set.md) - Set a key-value pair in a dictionary
- [`dict_unset`](dict_unset.md) - Remove a key from a dictionary
- [`dict_has_key`](dict_has_key.md) - Check if a dictionary contains a key
- [`dict_keys`](dict_keys.md) - Get all keys from one or more dictionaries
- [`dict_values`](dict_values.md) - Get all values from a dictionary
- [`dict_pick`](dict_pick.md) - Create a new dictionary with only specified keys
- [`dict_omit`](dict_omit.md) - Create a new dictionary excluding specified keys
- [`dict_pluck`](dict_pluck.md) - Extract values for a specific key from multiple dictionaries
- [`dict_find`](dict_find.md) - Navigate nested dictionaries safely with a default value
