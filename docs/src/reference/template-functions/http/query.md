# `query`

The `query` function allows for the issuing of SPARQL queries or parameterized SPARQL queries from within templates. The function takes one or more parameters. The first is the name of the query, and the following parameters, optionally, is are strings to inject into the query. The given injection strings will replace instances of `{{.}}` in their given order.

```
{{ query "name_of_parameterized_query.rq" "param" }}
```

```
{{ query "name_of_query.rq" }}
```
