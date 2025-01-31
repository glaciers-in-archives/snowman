# `views.yaml`

`views.yaml` connects you queries with your templates and assignes an output path to each pair. You can think of it as Snowman's router or controller.


Snowman have two different types of views. One which outputs a single file and forwards all of the query results to the template.

```yaml
views:
  - output: "index.html"
    query: "index.rq"
    template: "index.html"
```

The second outputs one file per SPARQL result(row) and takes one of the returned SPARQL-variables as a path argument. The following would therefore need a query which returns a `?id` variable for each row. Only the row is forwarded to the template.

```yaml
views:
  - output: "entities/{{id}}.html"
    query: "entities.rq"
    template: "entity.html"
```

## Non-HTML templates

By default Snowman uses HTML-aware templates which escapes `<`, `>`, etc as well as CSS and JavaScript. To disable this behaviour for a given view you can set the `unsafe` option.

You should never use this for HTML templates, instead you should use the `safe_html` template function to manage unsafe injections on a case-to-case basics.

```yaml
views:
  - output: "sitemap.xml"
    query: "entities.rq"
    template: "sitemap.xml"
    unsafe: true
```

