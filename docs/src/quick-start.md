# Quick start

```
snowman new  --directory="best-project-name-ever"
cd best-project-name-ever
snowman build && snowman server
```

This will create a new Snowman project in the directory `best-project-name-ever`, build it, and start a local server. You can now visit `http://localhost:8080` to see the built site. To get you started it fetches a few triples from Wikidata, gives you a basic layout, and a static file. 

## The project structure

The project structure upon using the `snowman new` command is as follows:

```
best-project-name-ever
├── queries # your SPARQL queries go here, create subdirectories to organize them
│   └── index.rq
├── static # static files go here, they are copied to the root of the build directory, make subdirectories to modify the output path
│   └── style.css
├── templates # your templates go here, create subdirectories to organize them, go beyond HTML!
    ├── includes # common name for components and partial templates
    │   └── footer.html
    ├── layouts # layouts are special templates that wrap other templates
    │   └── default.html
    ├── static.html # a page not feed by SPARQL but with full access to the template engine
    └── index.html # a page feed by SPARQL
├── snowman.yaml # core configuration go here, like the SPARQL endpoint and site metadata
└── views.yaml # all your views go here, a view connects a template to a SPARQL query, look for the index.html and static.html in this file
```
