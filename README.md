# Snowman

An **experimental** static site generator for SPARQL backends.

Copyright (c) 2020 Albin Larsson. Snowman is made available under the GNU Lesser General Public License.

## Installation

For the time being you will need to build Snowman from source:

```bash
git clone https://github.com/glaciers-in-archives/snowman
cd snowman
go build -o snowman
```

This will generate a binary that you can place in the root of your new project.

## Usage

For the time being, the easiest way to get started is by copying the Wikidata example and modifying it for your own needs. The Wikidata example will generate a website listing Douglas Adams various works. Run `snowman build` to generate the site.

### From scratch

This is a tutorial. You can at anytime run `snowman --help` for a full list of options.

#### Defining target endpoint

`snowman.yaml` should be located at the root of your project. It defines the URL of your SPARQL endpoint as well as optional HTTP headers and custom metadata.

```yaml
sparql_client:
  endpoint: "https://query.wikidata.org/sparql"
  http_headers: 
    User-Agent: "Snowman build example. https://github.com/glaciers-in-archives/snowman"
```

#### Defining queries

SPARQL queries provide data to views but because a single query can be used for multiple views and even partial rendering all your SPARQL files should be located in the `queries` directory(or child directories) of your project. Let's put this in `queries/works.rq`.

```sparql
SELECT ?qid ?title ?workLabel WHERE {
  BIND("Douglas Adams" AS ?title) .
  ?work wdt:P50 wd:Q42 .
  
  BIND(REPLACE(STR(?work), "http://www.wikidata.org/entity/", "") AS ?qid)

  SERVICE wikibase:label {
    bd:serviceParam wikibase:language "[AUTO_LANGUAGE],en" .
    ?work rdfs:label ?workLabel .
  }
}
```

#### Defining view templates

Snowman templates are [Go templates](https://golang.org/pkg/html/template/), a template can access a single SPARQL result or an entire resultset.

Let's start with an example demonstrating how to access data in a view template intended to access an entire resultset. Note that one needs to use the `index` and `range` keywords to access data. Let's put the following template in `templates/index.html`.

```html
<h1>Works by {{ (index . 0).title }}</h1>
<ul>
    {{ range . }}
    <li><a href="works/{{ .qid }}.html">{{ .workLabel }}</a></li>
    {{ end }}
</ul>
```

Snowman can also use each result in a SPARQL resultset to create a file for each result. If a view has been configured for this only a given result will be accessible from within a template. Put the following template in `templates/work.html`.

```html
<h1>{{ .workLabel }}</h1>
```

#### Turning templates and queries into views

By design, both templates and queries can be used across various views. For example, one can use the single query defined above in both of our templates. The following view will use the mentioned query and template to generate a file named `index.html` in your site's root. Views are placed in the `views` directory, name the following `index.yaml`.

```yaml
output: "index.html"
query: "works.rq"
template: "index.html"
```

While the above view takes all the results from the works query and forwards them to the template we can also generate a file from each result. We do this by wrapping the SPARQL variable we want to use in the resulting filename with double curly brackets in the `output` option. Note that the variable therefor needs to be unique.

The following view should generate a file for each result and use the `qid` SPARQL variable as the filename. You can name this view `work.yaml`.

```yaml
output: "works/{{qid}}.html"
query: "works.rq"
template: "work.html"
```

Now you can generate the site by running `snowman build`. Your static site should appear in the `site` directory in your project's root. `snowman clean` deletes the `site` directory so that you can regenerate the site when you have made changes to your code.

#### Static files

Static files are placed in the `static` directory and will be copied to the root of your built site. For example, the file `static/css/buttons.css` would be copied to `site/css/buttons.css`.

#### Layouts and child templates

Child templates and layouts are just regular Go templates that use the `define`, `block`, and `template` statements. To make layouts and child templates discoverable to Snowman they should be placed anywhere under `views`. You can see both layouts and child templates in the examples provided in the examples directory.

#### Static files with templates

If you want to use layouts and templates within a static file you need to create a view and a template for it but in the view configuration you exclude the `query` option.

#### Built in template functions

Note most functions do take strings and not RDF terms as arguments. You can access a string representation of an RDF term through rdfTerm.String.

##### Now

Snowman exposes the [time.Now](https://golang.org/pkg/time/#Now) function in all templates it can be used as follows:

```
{{ now.Format "2006-01-02" }}
```

```
{{ now.UTC.Year }}
```

For documentation on how to format dates see [the official Go documentation](https://golang.org/pkg/time/#pkg-constants).

##### Split

Snowman exposes the [strings.Split](https://golang.org/pkg/strings/#Split) function in all templates. The following example illustrates how to split a comma-separated in a range statement:

```
{{ range split .list_of_values "," }}
  {{ . }}
{{ end }}
```

##### Join

Snowman exposes a ´join´ function which can take a separator and any number of strings and merge them. The following example illustrates how to join three strings together with and without a separator:

```
{{ join "," "comma" "separated" }}
```

```
{{ join "" "Hello" " " "World" }}
```

##### Replace

Snowman exposes the [strings.Replace](https://golang.org/pkg/strings/#Replace) function in all templates. The following example illustrates how to replace a part of a string:

```
{{ replace . "https://en.wikipedia.org/wiki/" "" 1 }}
```

##### Env

`env` allows you to access environment variables from within your templates. `env` returns the value of an environment as a string.

```
{{ env "PATH" }}
```

##### Ucase, lcase, and tcase

Snowman provides `ucase`, `lcase`, and `tcase` for changing strings into uppercase, lowercase, and title-case.

```
{{ lcase .YourStringVariable }}
```

##### Query

Snowman provides a `query` function which allows one to issue a SPARQL query or a parameterized SPARQL query during rendering. The function takes two inputs, first the name of the query(without the `.rq` file extension) and then an optional string value to inject into the SPARQL query. The location for the injected string is set with `{{.}}`.

```
{{ $sparql_result := query "name_of_parameterized_query" $var }}
{{ $another_resultset := query "name_of_query" }}
```

##### Config

Snowman exposes your site's configuration through the function `config`. The following example illustrates how to retrieve your SPARQL endpoint:

```
{{ $yourVariable := config }}
{{ $yourVariable.Endpoint }}
```

##### Metadata

The `metadata` function is a shortcut for accessing the metadata defined in your site's configuration.

```
{{ $yourVariable := metadata }}
```

##### Safe HTML

The `safe_html` function allows you to render a HTML string as it is without the default escaping performed in unsafe templates. **Note that you should never trust third party HTML.**

```
{{ safe_html "<p>This renders as HTML</p>" }}
```

##### Add

The `add` function sums integer values, it can take any number of arguments beyond two.

```
{{ add 5 6 7 }}
```

##### Sub

The `sub` function subtract integer values, it must take two arguments.

```
{{ sub 10 5 }}
```

##### Div

The `div` function divides integer values, it must take two arguments.

```
{{ div 10 2 }}
```

##### Mul

The `mul` function multiplies integer values, it can take any number of arguments beyond two.

```
{{ mul 5 6 7 }}
```

##### Mod

The `mod` function returns the modulus of two given values.

```
{{ div 5 2 }}
```

##### Rand

The `rand` function returns a random integer between the two given values.

```
{{ rand 5 10 }}
```

##### Add1

The `add1` function increments the given integer by 1.

```
{{ add1 $your_intreger }}
```

### Working with cache

#### Default behaviour

By default, Snowman will only issue SPARQL queries when the result of a query is not found in the cache. To ignore the cache nor update it you can use the `cache` flag when running the `build` command to set a cache strategy:

```bash
snowman build --cache never
```

#### Inspect cache

Snowman allows you to inspect the cached data for a particular query or parameterized query using the `cache` command. The cache command takes one-two arguments, first the path of your query and optionally the argument used in a parameterized query.

```bash
snowman cache list-of-icecream.rq

snowman cache icecream.rq "your parameter"
```

#### Invalidate cache

Especially when you build very large sites or use expensive SPARQL queries it can be useful to invalidate specific portions of the cache. You can do so using the `cache` command. Specify the query or parameterized query for which you want to invalidate the cache and add the flag `invalidate`.

```bash
snowman cache list-of-icecream.rq --invalidate

snowman cache icecream.rq "your parameter" --invalidate
```