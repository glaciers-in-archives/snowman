# `version`

The `version` command prints the version of Snowman.

```sh
snowman version
```

## Format

The output is a single line consisting of three parts separated by spaces:

```
Snowman 0.5.0-development linux/amd64
```

1. The name of the application, allways `Snowman`.
2. The version of the application([semver](https://semver.org/)).
3. The operating system and architecture the binary was built for.

## Flags

- `--help` (`-h`) Shows help for the command.
