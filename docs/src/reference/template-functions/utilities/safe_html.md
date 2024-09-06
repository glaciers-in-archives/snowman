# `safe_html`

The `safe_html` function allows you to render an HTML string as-is, without the default escaping performed in HTML/unsafe templates.

```
{{ safe_html "<p>This renders as HTML</p>" }}
```

Note that you should only use with content you trust and control, as it can expose your site to cross-site scripting attacks.
