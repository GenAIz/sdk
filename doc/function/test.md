## Function Test

```bash
genaiz sf test --context=PATH --file=FILE --tag=LOCAL --version=VERSION \
  --mount-in=PATH --mount-out=PATH --mount-log=PATH --mount-var=PATH \
  --image=IMAGE --prefix|p=PREFIX
```

Test executes a build task if the image option is not specified. A fresh image with the tag and version specified under the `Genaiz.yaml` file or from the options on the command line will be built.

The command the proceeds to run the image built under a **disposable** container using the mount points specified under the `Genaiz.yaml` file or from the options on the command line. The container standard out and error should be attached to the command's console until it completes.

Test will forward signals such as SIGINT (ctrl-C) to the container running on containerd.

### context

Context for the test command establishes the working directory for all the specified mount paths and for the [build](build.md) tasks as well.

* if there is no context specified the test command assumes all paths from the current working dir.
* the context is passed to the docker build command or the dockerd build endpoint to match the Dockerfile's build context, it doesn't necessarily imply the same folder as the [file](#file) param.
* if the resolved context does not correspond to an existing folder, the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.build.context] is invalid`
    * The option key used is for **build**, test does not have its own context key

### file

File should not be considered if the [image](#image) option is used, as it would disable any pre-build tasks. This parameter is used to set the [file](build.md#file) parameter on the pre-build task otherwise.

* if there is no file specified the build command assumes it is looking for a Dockerfile under the specified context.
* if the command can not find the resolved file, it will return as error of the form: `Error: value [...] for option [sf.build.file] is invalid`
    * The option key used is for **build**, test does not have its own context key

### tag

Tag can be used to change the local `name/repository:tag` of the built image. This should only affect the [list](list.md) command and the name displayed under a `docker image ls` command.

* if tag is not specified, a default value will be composed of with the current working direction: `parent/current:version`
* if the resolved tag is not a valid string matching a valid repository string (see [repository validity](index.md#repository)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [sf.build.tag] is invalid`
    * The option key used is for **build**, test does not have its own context key

### version

Version can be used to change the local `name/repository:tag` of the built image. This should only affect the [list](list.md) command and the name displayed under a `docker image ls` command.

* if version is not specified, `latest` will be used by default.
* the build version is not the same as the published version, in normal circumstances the version will always be `latest`, but since smart functions can be branched in source repositories, the option can be updated with the [init](init.md) command to say `<branch>-latest`
* note that the version string used for build is not validated against the SemVer format
    * The option key used is for **build**, test does not have its own context key

### mount-in

* if mount-in is not specified, the default key `sf.test.in` will be read from the Smart Function `Genaiz.yaml`
* if mount-in is not specified, no [SF_INPUT_PATH](index.md#sf_input_path) host mount point will be provided to the function's container.
* if mount-in specified does not resolve to an existing path, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [sf.test.input] is invalid`

### mount-out

* if mount-out is not specified, the default key `sf.test.out` will be read from the Smart Function `Genaiz.yaml`
* if mount-out is not specified, no [SF_OUTPUT_PATH](index.md#sf_output_path) host mount point will be provided to the function's container.
* if mount-out specified does not resolve to an existing path, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [sf.test.output] is invalid`

### mount-log

* if mount-log is not specified, the default key `sf.test.log` will be read from the Smart Function `Genaiz.yaml`
* if mount-log is not specified, no [SF_LOG_PATH](index.md#sf_log_path) host mount point will be provided to the function's container.
* if mount-log specified does not resolve to an existing path, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [sf.test.log] is invalid`

### mount-var

* if mount-var is not specified, the default key `sf.test.var` will be read from the Smart Function `Genaiz.yaml`
* if mount-var is not specified, no [SF_VAR_PATH](index.md#sf_var_path) host mount point will be provided to the function's container.
* if mount-var specified does not resolve to an existing path, the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [sf.test.var] is invalid`

### image

* if image is not specified, the command will build an equivalent repository string of the form `tag:version`, with the fields of the build command.
* if image is specified, the command will use it as-is to retrieve an image from `containerd`, the local Docker registry, and run it as-is.
    * This is useful for running published or imported functions

### prefix

The prefix field is not particularly useful to test, since it disposes of the created containers when they exit. When test outputs data on STDOUT from the attached Smart Function, it is possible to use another console to run `docker container ls` or `genaiz sf list` within the build context and see the container prefix.

* if prefix is used, the container created will be named `<prefix>-d<timestamp>`
    * note that run disposes of the containers created automatically, see [start](start.md) command to preserve them
* if prefix is not specified, the container will be named `<tag>-d<timestamp>`
* if the prefix does not match a valid component string (see [component validity](index.md#component)), the command will return an error with the key of the field and the invalid value: `Error: value [...] for option [sf.test.prefix] is invalid`