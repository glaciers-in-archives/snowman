# `current_view`

The `current_view` function return the configuration of the view being rendered.

```
{{ current_view }}
```

Note that the`curre`t_view` function isn't available when used inside of templates included using the `include` or `include_tex` functions. You can however pass the value of `current_view` as an argument to the included template:

```
{{ include "includes/child-template-aware-of-the-view.html" $current_view }}
```
