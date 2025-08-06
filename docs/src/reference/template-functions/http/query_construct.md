# `query_construct`

The `query_construct` function allows for the issuing of SPARQL CONSTRUCT queries or parameterized SPARQL CONSTRUCT queries from within templates. SPARQL CONSTRUCT queries return JSON-LD that can be processed further using other template functions.

The function takes one or more parameters. The first is the name of the query, and the following parameters, optionally, is are strings to inject into the query. The given injection strings will replace instances of `{{.}}` in their given order.

```html
{{ $results1 := query_construct "name_of_query.rq" }}
{{ $results2 := query_construct "name_of_parameterized_query.rq" "param" }}
```

```html
{{ $results := query_construct "skos_concepts.rq" .collection }}
{{ range $results }}
<article>
  <h1>URI: {{ index . "@id" }}</h1>
  <ul>
  {{ range (index . "http://www.w3.org/2004/02/skos/core#prefLabel") }}
    <li lang="{{ index . "@lang" }}">{{ index . "@value" }}</li>
  {{ end }}
  </ul>
</article>
{{ end }}
```
