# Snowman

*A static site generator for SPARQL backends.*

![go version badge](https://img.shields.io/github/go-mod/go-version/glaciers-in-archives/snowman) [![codeclimate badge](https://img.shields.io/codeclimate/maintainability/glaciers-in-archives/snowman)](https://codeclimate.com/github/glaciers-in-archives/snowman/maintainability) ![license badge](https://img.shields.io/github/license/glaciers-in-archives/snowman)

Snowman is a static site generator for SPARQL backends. Snowman is designed to allow RDF-based projects to use SPARQL in the user-facing parts of their stack, even at scale. Snowman powers projects rendering simple SKOS vocabularies as well as projects rendering entire knowledge bases.

## Installation

For the time being, you will need to build Snowman from source:

```bash
git clone https://github.com/glaciers-in-archives/snowman
cd snowman
go build -o snowman
```

This will generate a binary that you can place in the root of your new project.

## Usage

For the time being, the easiest way to get started is by copying the Wikidata example and modifying it for your own needs. The Wikidata example will generate a website listing Douglas Adams various works. Run `snowman build` to generate the site. The `snowman server` command can then be used to serve the site with Snowman's built-in development server.

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

#### Connecting templates and queries with views

By design, both templates and queries can be used across various views. For example, one can use the single query defined above in both of our templates. The following view will use the mentioned query and template to generate a file named `index.html` in your site's root. Views are defined in a file named `views.yaml` which should be placed in your project's root folder.

```yaml
views:
  - output: "index.html"
    query: "works.rq"
    template: "index.html"
```

While the above view takes all the results from the works query and forwards them to the template, we can also generate a file from each result. We do this by wrapping the SPARQL variable we want to use in the resulting filename with double curly brackets in the `output` option. Note that the variable, therefore, needs to be unique.

The following view should generate a file for each result and use the `qid` SPARQL variable as the filename. You should append the following YAML to `views.yaml`

```yaml
  - output: "works/{{qid}}.html"
    query: "works.rq"
    template: "work.html"
```

Now you can generate the site by running `snowman build`. Your static site should appear in the `site` directory in your project's root. To run the site you can use the `snowman server` command.

#### Static files

Static files are placed in the `static` directory and will be copied to the root of your built site. For example, the file `static/css/buttons.css` would be copied to `site/css/buttons.css`.

#### Child templates

While child templates are regular Go templates, they are invoked with Snowman's `include` or `include_text` functions with the full path to a template rather than a Go template name.

`include` will expect HTML templates while `include_text` will treat the rendered content like text and might escape it if the parent template is an HTML template.

#### Layouts

Layouts in Snowman are regular Go templates that are defined with `define` and `block` statements and used with the `template` statement. Layout files must, however, be placed under `templates/layouts` to be discovered by Snowman.
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

For documentation on how to format dates, see [the official Go documentation](https://golang.org/pkg/time/#pkg-constants).

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

Snowman provides a `query` function which allows one to issue a SPARQL query or a parameterized SPARQL query during rendering. The function takes two inputs, first the name of the query and then an optional string value to inject into the SPARQL query. The location for the injected string is set with `{{.}}`.

```
{{ $sparql_result := query "name_of_parameterized_query.rq" $var }}
{{ $another_resultset := query "name_of_query.rq" }}
```

##### Config

Snowman exposes your site's configuration through the function `config`. The following example illustrates how to retrieve your SPARQL endpoint:

```
{{ $yourVariable := config }}
{{ $yourVariable.Endpoint }}
```

##### Safe HTML

The `safe_html` function allows you to render a HTML string as it is without the default escaping performed in unsafe templates. **Note that you should never trust third party HTML.**

```
{{ safe_html "<p>This renders as HTML</p>" }}
```

##### URI

The `uri` function takes a string and tries to cast it to a URI, if it fails it will produce an error.

```
{{ uri "https://schema.org/Person" }}
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

### Timing your builds

Sometimes when you work on large sites it can be useful to time your builds to mesuare the impcat of various changes. All Snowman commands therefore got a flag named `timeit`. "Time it" will once the command finishes executing print its execuation time. While this is mostly useful for mesuaring built times all Snowman commands support it.

## License

Copyright (c) 2020 Albin Larsson. Snowman is made available under the GNU Lesser General Public License.
