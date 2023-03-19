![Snowman - A static site generator for SPARQL backends.](assets/snowman-header.svg)

![go version badge](https://img.shields.io/github/go-mod/go-version/glaciers-in-archives/snowman) [![codeclimate badge](https://img.shields.io/codeclimate/maintainability/glaciers-in-archives/snowman)](https://codeclimate.com/github/glaciers-in-archives/snowman/maintainability) ![license badge](https://img.shields.io/github/license/glaciers-in-archives/snowman)

Snowman is designed to allow RDF-based projects to use SPARQL in the user-facing parts of their stack, even at scale. Snowman powers projects rendering simple SKOS vocabularies as well as projects rendering entire knowledge bases. Snowman's templating system comes with RDF- and SPARQL-tailored functions, and features and takes its data from SPARQL queries.

## Installation

[Download the latest release for your OS/architecture](https://github.com/glaciers-in-archives/snowman/releases).

If your OS/architecture combination is not available, you will need to build Snowman from source:

```bash
git clone https://github.com/glaciers-in-archives/snowman
cd snowman
go build -o snowman
```

## Usage

One way to get started is by copying the Wikidata example and modifying it for your own needs. The Wikidata example will generate a website listing Douglas Adams' various works. Run `snowman build` to generate the site. The `snowman server` command can then be used to serve the site with Snowman's built-in development server.

### From scratch

**As of Snowman 0.3.0 you can scaffold a new project with `snowman new --directory="project-name"`.**

This is a tutorial. You can at any time run `snowman --help` for a full list of options.

#### Setting the target endpoint

`snowman.yaml` should be located in the root directory of your project. It defines the URL of your SPARQL endpoint as well as optional HTTP headers and custom metadata.

```yaml
sparql_client:
  endpoint: "https://query.wikidata.org/sparql"
  http_headers: 
    User-Agent: "project-tutorial Snowman (https://github.com/glaciers-in-archives/snowman)"
```

#### Defining queries

SPARQL queries provide data to views, but, because a single query can be used for multiple views and even partial rendering, all your SPARQL files should be located in the `queries` directory (or child directories) of your project. Let's put this in `queries/works.rq`:

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

Snowman uses [Go templates](https://golang.org/pkg/html/template/). A template can access a single SPARQL result, or an entire resultset.

Let's start with an example that demonstrates how to access data in a view template intended to access an entire resultset. Note that the `index` and `range` keywords must be used to access data. Let's put the following in `templates/index.html`:

```html
<h1>Works by {{ (index . 0).title }}</h1>
<ul>
    {{ range . }}
    <li><a href="works/{{ .qid }}.html">{{ .workLabel }}</a></li>
    {{ end }}
</ul>
```

Snowman can also create a file from each result in a resultset. If a view has been configured for this, only a given result will be accessible from within a template. Put the following template in `templates/work.html`.

```html
<h1>{{ .workLabel }}</h1>
```

#### Connecting templates and queries with views

By design, both templates and queries can be used across various views. For example, one could use the single query defined above in both of our templates. The following view will use the specified query and template to generate a file named `index.html` in the root directory of your site.

Views are defined in a file named `views.yaml`, which should be in the root directory of your project:

```yaml
views:
  - output: "index.html"
    query: "works.rq"
    template: "index.html"
```

While the above view takes all the results from the works query and forwards them to the template, we can also generate a file from each result. We do this by wrapping the SPARQL variable we want to use in the resulting filename with double curly brackets in the `output` option. Note that the variable, therefore, must be unique.

The following view should generate a file for each result and use the `qid` SPARQL variable as the filename. You should append the following YAML to `views.yaml`:

```yaml
  - output: "works/{{qid}}.html"
    query: "works.rq"
    template: "work.html"
```

Now you can generate the site by running `snowman build`. Your static site should appear in the `site` directory in the root directory of your project. To start the server and view your site, run the `snowman server` command.

## Documentation
### Static files

Static files are placed in the `static` directory and will be copied to the root of your built site. For example, the file `static/css/buttons.css` would be copied to `site/css/buttons.css`.

If you have made changes to static files only and want to rebuild your site, you can do so with the `snowman build --static` command. The `static` flag ensures that Snowman updates only static files, rather than doing a full build.

### Child templates

While child templates are regular Go templates, they are invoked with Snowman's `include` or `include_text` functions with the full path to a template rather than a Go template name.

`include` expects HTML templates, while `include_text` will treat the rendered content as text, and might escape it if the parent template is an HTML template.

### Layouts

Layouts in Snowman are regular Go templates that are defined with `define` and `block` statements and are used with the `template` statement. Layout files must, however, be placed under `templates/layouts` to be discovered by Snowman.

### Static files with templates

If you want to use layouts and templates within a static file, you'll need to create a view and a template for it, but in the view configuration you should exclude the `query` option.

### Built-in template functions

Note that most functions take strings as arguments, and not RDF terms. You can access a string representation of an RDF term through `rdfTerm.String`.

##### Now

Snowman exposes the [time.Now](https://golang.org/pkg/time/#Now) function in all templates. It can be used as follows:

```
{{ now.Format "2006-01-02" }}
```

```
{{ now.UTC.Year }}
```

For more on how to format dates, see [the official Go documentation](https://golang.org/pkg/time/#pkg-constants).

##### Split

Snowman exposes the [strings.Split](https://golang.org/pkg/strings/#Split) function in all templates. The following example illustrates how to split a comma-separated string in a `range` statement:

```
{{ range split .list_of_values "," }}
  {{ . }}
{{ end }}
```

##### Join

Snowman exposes a ´join´ function which takes a separator and any number of strings and merges them. The following examples illustrate how to merge three strings—first with, and then without, a separator:

```
{{ join "," "comma" "separated" }}
```

```
{{ join "" "Hello" " " "World" }}
```

##### Replace

Snowman exposes the [strings.Replace](https://golang.org/pkg/strings/#Replace) function in all templates. The following example illustrates how to replace part of a string:

```
{{ replace . "https://en.wikipedia.org/wiki/" "" 1 }}
```

##### Env

`env` allows you to access environment variables from within your templates. `env` returns the value of an environment as a string.

```
{{ env "PATH" }}
```

##### Ucase, lcase, and tcase

Snowman provides `ucase`, `lcase`, and `tcase` for changing strings into uppercase, lowercase, and title case respectively.

```
{{ lcase .YourStringVariable }}
```

##### Query

Snowman provides a `query` function that allows for the issuing of SPARQL queries or parameterized SPARQL queries during rendering. The function takes one or more parameters. The first is the name of the query, and the following parameters, optionally, is are strings to inject into the query. The given injection strings will replace instances of `{{.}}` in their given order.

```
{{ $sparql_result := query "name_of_parameterized_query.rq" $var }}
{{ $another_resultset := query "name_of_query.rq" }}
```

##### Config

Snowman exposes your site's configuration through the function `config`. The following example illustrates how to retrieve your SPARQL endpoint:

```
{{ config.Client.Endpoint }}
```

##### Safe HTML

The `safe_html` function allows you to render an HTML string as-is, without the default escaping performed in unsafe templates. **Note that you should never trust third-party HTML.**

```
{{ safe_html "<p>This renders as HTML</p>" }}
```

##### URI

The `uri` function takes a string and attempts to cast it to a URI, and produces an error upon failure.

```
{{ uri "https://schema.org/Person" }}
```

##### Add

The `add` function sums integer values and takes at least two arguments.

```
{{ add 5 6 7 }}
```

##### Sub

The `sub` function subtracts two given integer values.

```
{{ sub 10 5 }}
```

##### Div

The `div` function divides two given integer values.

```
{{ div 10 2 }}
```

##### Mul

The `mul` function multiplies two given integer values.

```
{{ mul 5 6 }}
```

##### Mod

The `mod` function returns the modulus of two given values.

```
{{ div 5 2 }}
```

##### Rand

Given two values, the `rand` function returns a random integer between them.

```
{{ rand 5 10 }}
```

##### Add1

The `add1` function increments the given integer by 1.

```
{{ add1 $your_integer }}
```
##### Type

The `type` function returns the given variable's type as a string.

```
{{ type $uri_html_string_or_anything_else }}
```

##### To JSON

The `to_json` function converts a given argument to a JSON-formatted string.

```
{{ to_json $your_variable }}
```

##### From JSON

The `from_json` function converts a given JSON-formatted string to a Go-interface which templates can use.

```
{{ from_json $your_json_string }}
```

##### Version

The `version` function returns the Snowman version used to build the page.

```
{{ version }}
```

##### Trim

The `trim` function trims leading and trailing white space from a given string.

```
{{ trim $your_variable }}
```

##### Get Remote

The `get_remote` function retrieves the contents of a remote URL and returns it as a string.

```
{{ get_remote "https://fornpunkt.se/lamning/lNJVbNa.geojson" }}
```

Combine it with `from_json` to parse remote JSON.

##### Get Remote with Config

The `get_remote_with_config` function also retrieves the contents of a remote URL and returns it as a string. However, it takes a second argument, which allows you to set custom HTTP request headers.

```
{{ get_remote "https://fornpunkt.se/lamning/lNJVbNa.geojson" $your_config }}
```

### Working with cache

#### Default behaviour

By default, Snowman will issue SPARQL queries only when the result of a query is not found in the cache. To ignore the cache or update it, use the `cache` flag when running the `build` command to set a caching strategy:

```bash
snowman build --cache never
```

#### Inspect cache

Snowman allows you to inspect the cached data for a particular query or parameterized query using the `cache` command. The cache command takes as arguments first the path of the query and then, optionally, the argument used in a parameterized query:

```bash
snowman cache list-of-icecream.rq

snowman cache icecream.rq "your parameter"
```

#### Invalidate cache

Especially when you build very large sites or use expensive SPARQL queries it can be useful to invalidate specific portions of the cache. You can do so using the `cache` command. Specify the query or parameterized query for which you want to invalidate the cache, and add the flag `invalidate`:

```bash
snowman cache list-of-icecream.rq --invalidate

snowman cache icecream.rq "your parameter" --invalidate
```

Sometimes, following changes to your queries and external data, you can end up with unused cache items. You can clear these using the `--unused` selector flag:

```bash
snowman cache --unused --invalidate
```

### Using the built-in server

Snowman comes with a built-in development server exposed through the `server` command. The `server` command has two optional arguments, `port` and `address`, which can be used to bind Snowman to an IP address and port:

```bash
snowman server

snowman server --port 4000 --address 0.0.0.0
```

### Timing your builds

Sometimes when you work on large sites, it can be useful to time your build processes to measure the impact of changes. All Snowman commands, therefore, have a flag named `timeit`. This prints a command's execution time to the console. While this is mostly useful for measuring build times, all Snowman commands support it.

## License

Copyright (c) 2020- Albin Larsson & contributors. Snowman is made available under the GNU Lesser General Public License.
