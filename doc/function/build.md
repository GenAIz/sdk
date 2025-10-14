## Function Build

```bash
genaiz sf build --context=PATH --file=FILE --tag=LOCAL --version=VERSION \
--label --legacy-builder --no-cache --prune
 ```
Build executes a single build task which may or may not use the docker-cli executable to complete the command. Build will use the legacy Moby compiled in build instruction, using dockerd if it can not locate a docker command on the user's PATH. Otherwise, it will invoke `docker build` as a child process.

A successful command should print a summary of the image, created with its tag, short id and size. The id printed should be referenced under the [list](list.md) command with matching size and tag.

### context

* if there is no context specified the build command assumes it is established as the current working dir.
* the context is passed to the docker build command or the dockerd build endpoint to match the Dockerfile's build context, it doesn't necessarily imply the same folder as the [file](#file) param.
* if the resolved context does not correspond to an existing folder, the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.build.context] is invalid`

### file

* if there is no file specified the build command assumes it is looking for a Dockerfile under the specified context.
* if the command can not find the resolved file, it will return as error of the form: `Error: value [...] for option [sf.build.file] is invalid`

### tag

Tag can be used to change the local `name/repository:tag` of the built image. It will not affect the [publish](publish.md) command, which will still use `oem/handle:version`. The [list](list.md) command uses the same parameter of filter images built.

* if tag is not specified, a default value will be composed of with the current working direction: `parent/current:version`
* if the resolved is not a valid string matching a valid repository string (see [repository validity](index.md#repository)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.build.tag] is invalid`

### version

* if version is not specified, `latest` will be used by default.
* the build version is not the same as the published version, in normal circumstances the version will always be `latest`, but since smart functions can be branched in source repositories, the option can be updated with the [init](init.md) command to say `<branch>-latest`
* note that the version string used for build is not validated against the SemVer format

### label

* label instructs the build command to add a docker label to the image built with the same value as the [tag](#tag) value.
* labels are used by the [prune](#prune) option to remove previously built dangling images with the same label value.

### legacy-builder

* legacy-builder instructs the build command to use the legacy build endpoint of dockerd to create the image. This can be useful in cases where the `docker` command is not available or in cases where Buildkit is not compatible with the user's platform (Windows)
* by default, the build command will look up the users' PATH to locate `docker`. If it's not located, the legacy-builder will be used.

### no-cache

* no-cache is an option passed to the docker builder. It instructs the builder to try avoiding caching Dockerfile statements which can have a serious effect if the file retrieves binaries, or libraries, with single tags, like `latest`, or without versions relying on the default.
* no-cache used with [prune](#prune) instructs the build command to prune all intermediary built artefacts. The default behavior is to prune only since 12h

### prune

* prune instructs the build command to remove all dangling images with the same [label](#label) as the [tag](#tag) value.
* if using prune without the legacy builder, it will also try to remove all intermediary build artefacts passed 12h of lifetime. If used with [no-cache](#no-cache), it removes all artefacts.