# `read_file`

The `read_file` function reads the contents of a file and returns it as a string.

```
{{ read_file "relative/path/to/file.txt" }}
```

Note that the path must be relative to the root of the project and that Snowman do not have access to files outside of the project directory.
