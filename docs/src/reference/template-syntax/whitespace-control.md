# Whitespace control

Snowman templates renders everything between code blocks and comments including whitespace and line feeds. As a result the output can look messy. To control the extra white whitespace in the output you can use trim markers.

Considering the following example.

```
<ul>
{{ range .foos }}
    <li>list item</li>
{{ end }}
</ul>
```

The `{{ range .foos }}` and `{{ end }}` blocks are sourrounded line breaks which will be rendered in the output. To remove the extra whitespace you can use trim markers.

```
<ul>
{{- range .foos -}}
    <li>list item</li>
{{- end -}}
</ul>
```

The `-` is optional and can be placed on either side of the block. The following example will render in the same way as the previous example. 

```
<ul>{{ range .foos -}}
    <li>list item</li>
{{- end }}</ul>
```

You can also use trim markers with comments:

{{- /* this is a comment with trim markers */ -}}
