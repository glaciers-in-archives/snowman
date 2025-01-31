# `now`

`now` returns the current date and time in the requested format. The function takes a format string as an argument.

```
{{ now.Format "2006-01-02" }}
```

```
{{ now.UTC.Year }}
```

This function is derived from the Go [time.Now](https://golang.org/pkg/time/#Now) function. For more on how to format dates, see [the official Go documentation](https://golang.org/pkg/time/#pkg-constants).
