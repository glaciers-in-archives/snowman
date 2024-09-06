# `include_text`

The ìnclude_text` function can be used to insert another text template at its position during rendering. The function takes a mandatory argument, the path to the template which should be included, as well as any number of additional arguments which will be passed to the included template.

```
{{ include_text "includes/description.txt" $description }}
```
