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

For the time being the easiest way to get started is by copying the Wikidata example and modifying it for your own needs. The Wikidata example will generate a website listing Douglas Adams various works. Run `snowman build` to generate the site.

### From scratch

This is a tutorial. You can at anytime run `snowman --help` for a full list of options.

#### Defining target endpoint

`snowman.yaml` should be located in the root of your project and define the URL of your SPARQL endpoint.

```yaml
---
  sparql_endpoint: "https://query.wikidata.org/sparql"
```

#### Defining queries

SPARQL queries provides data to views but because a single query can be used for multiple views and even partial rendering all your SPARQL files should be located in the `queries` directory(or child directories) of your project. Let's put this in `queries/works.rq`.

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

Snowman templates are just [Go templates](https://golang.org/pkg/html/template/), a template can access a single SPARQL result or an entire resultset.

Let's start with an example demonstrating how to access data in a view template intended to access an entire resultset. Note that one needs to use the `index` and `range` keywords to access data. Let's put the following template in `templates/index.html`.

```html
<h1>Works by {{ (index (index . 0) "title").Value }}</h1>
<ul>
    {{ range . }}
    <li><a href="works/{{ (index . "qid").Value }}.html">{{ (index . "workLabel").Value }}</a></li>
    {{ end }}
</ul>
```

Snowman can also use each result in a SPARQL resultset to create a file for each result. If a view has been configured for this only a given result is accessible from within a template. Put the following template in `templates/work.html`.

```html
<h1>{{ (index . "workLabel").Value }}</h1>
```

#### Turning templates and queries into views

By design, both templates and queries can be used across various views. For example, one can use the single query defined above in both of our templates. The following view will use the mentioned query and template to generate a file named `index.html` in your site's root. Views are placed in the `views` directory, name the following `index.yaml`.

```yaml
---
    output: "index.html"
    query: "works.rq"
    template: "index.html"
```

While the above view takes all the results from the works query and forwards them to the template we can also generate a file from each result. We do this by wrapping the SPARQL variable we want to use in the resulting filename with double curly brackets in the `output` option. Note that the variable therefor needs to be unique.

The following view should generate a file for each result and use the `qid` SPARQL variable as the filename. You can name this view `work.yaml`.

```yaml
---
    output: "works/{{qid}}.html"
    query: "works.rq"
    template: "work.html"
```

Now you can generate the site by running `snowman build`. Your static site should appear in the `site` directory in your project's root. `snowman clean` deletes the `site` directory so that you can regenerate the site when you have made changes to your code.

#### Static files

Static files are placed in the `static` directory and will be copied to the root of your built site. For example, the file `static/css/buttons.css` would be copied to `site/css/buttons.css`.

#### Layouts and child templates

Child templates and layouts are just regular Go templates that use the `define`, `block`, and `template` statements. To make layouts and child templates discoverable to Snowman they should be placed under `views/includes`. You can see both layouts and child templates in the examples provided in the examples directory.
