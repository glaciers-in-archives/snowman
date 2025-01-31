# Multilangual RDF labels

When a RDF literal has a associated language you can access it by prefixing your template variable with `.Lang`.

```html
<span lang="{{ .name.Lang }}">{{ .name }}</span>
```

