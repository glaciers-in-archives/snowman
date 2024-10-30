# Variables

Variables can be defined inside of templates by using a keyword prefixed with `$`:

```
{{ $my_variable := "this is a string" }}
```

To change the value of a existing variable:

```
{{ $my_variable = "this is a new string" }}
```

Variables initiated by a control-structure(such as a `range` statement) is accessed by the special variable `.`:

```
{{ range $my_list }}
    Current list item: {{ . }}
{{ end }}
```

Fields can be assesed by suffixing your variable name with `.` followed by the field name:

```
{{ $my_other_variable.A_field }}
```

A variable’s scope extends to the `end` action of the control structure in which it is declared, or to the end of the template if there is no such control structure.

