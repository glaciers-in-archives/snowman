# Installation

Snowman ships as a single binary, and is available for the most commons architectures and operating systems. Check out the [releases page](https://github.com/glaciers-in-archives/snowman/releases) for the latest version. Once downloaded, you should rename it to `snowman` and install it by moving it to a directory in your `PATH`.

## Using multiple versions

If you need to use multiple versions of Snowman, you can either rename the binary to something like `snowman-0.5.0` and then symlink it to `snowman` in your `PATH` or use it directly by specifying the path to the binary(`./path/to/snowman`).

To ensure team members are using compatible versions, you can specify a version requirement in your `snowman.yaml`:

```yaml
snowman_version: ">=0.7.0"
```

See the [`snowman.yaml` reference](reference/snowman-yaml.md#version-requirements) for more details on version constraints.

## Installing from source

If you would want to compile from source, you can do so:

```sh
git clone https://github.com/glaciers-in-archives/snowman
cd snowman
go build -o snowman
```

For all possible target operating systems and architectures, see the the [following table](https://www.digitalocean.com/community/tutorials/how-to-build-go-executables-for-multiple-platforms-on-ubuntu-16-04).
