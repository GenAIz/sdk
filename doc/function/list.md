## Function List

```
genaiz sf list --context=PATH --file=FILE --tag=TAG_MATCH --version=VERSION_MATCH
```

List executes the equivalent of a `docker image ls` and a `docker container ls --all` with the necessary filter configuration to only display images and containers related to the current Smart Function working dir or the specified **context**.

### context

Context for the list command establishes the working directory of the Smart Function for which we want a list of images and containers.

* if there is no context specified the run command assumes all paths from the current working dir.
* if the resolved context does not correspond to an existing folder, the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [function.build.context] is invalid`
    * The option key used is for **build**, list does not have its own context key

### file

> [!NOTE]
> List does not use the file argument for the current implementation of the SDK. It is provided as a common argument and will be validated with the same rules as other smart function commands, but should not be used. 

* if there is no file specified the build command assumes it is looking for a Dockerfile under the specified context.
* if the command can not find the resolved file, it will return as error of the form: `Error: value [...] for option [function.build.file] is invalid`
    * The option key used is for **build**, list does not have its own context key

### tag

Tag can be used to set the container prefix to be matched. The list function looks for images and containers prefixed by their repository name

* if tag is not specified, the value specified under `Genaiz.yaml` under the [context](#context) will be used.
* if tag is empty, a default value will be composed of with the current working direction: `parent/current:version`
* if the resolved tag is not a valid string matching a valid repository string (see [repository validity](index.md#repository)), the command will return an error with the key of the field and the invalid value; `Error: value [...] for option [function.build.tag] is invalid`
    * The option key used is for **build**, list does not have its own context key

### version

Version can be used to set the image version to be matched. The list function will look for images and containers which only belong to this version.

* if version is not specified, `latest` will be used by default.
* note that the version string used for build is not validated against the SemVer format
    * The option key used is for **build**, list does not have its own context key
