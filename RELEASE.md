# Snowman release documentation

## Before you make a release

 - Ensure that all official examples can be built using the new executables.
 - Test for potential performance declines by using the new version to build the Glaciers in archives website using `--timeit`.
 - Ensure that there are no high priority requirements from the Glaciers in archives project.

## How to make a release

1. Start to document the changes and create a draft release on Github using one of the [past releases](https://github.com/glaciers-in-archives/snowman/releases).
2. Bump the version number in [`current_version.go`](https://github.com/glaciers-in-archives/snowman/blob/main/internal/version/current_version.go)
3. Build executables for all supported platforms using `release.bash`. Note that you must trim the project's path.
4. Upload the executables to your draft release.
5. Give the release notes a read-through and publish the release!
