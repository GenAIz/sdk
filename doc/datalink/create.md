## Datalink Create

```
genaiz dk create HANDLE [FOLDER] --oem=OEM --version=VERSION \
  --name=NAME --description=DESCRIPTION \
```

### handle

* if handle does not match a valid handle string, (see [handle validity](index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `[...] is not a valid handle`

### folder

* if folder is specified the working directory will be set to folder for the create command.

> [!CAUTION]
> The folder path will only be used when we can sync data link properties

### oem

* if oem does not match a valid oem string, (see [oem validity](index.md#handle-and-oem)), the command will return an error with the field and the invalid value: `value [...] for option [datalink.create.oem] is invalid`

> [!CAUTION]
> The oem value will be inferrable from the [folder](#folder) path if the path contains a function

### version

* if version is not specified, the data link is created at version 1.0.0
* if version does not match a valid version string, (see [version validity](index.md#version)), the command will return an error with the field and the invalid value: `value [...] for option [datalink.create.version] is invalid`

### name

* if name is not specified, the value of handle will be used.
* if name does not match a valid name string, (see [name validity](index.md#name)), the command will return an error with the field and the shortened value: `value [...] for option [datalink.create.name] is invalid`

### description

* if description does not match a valid description, (see [description validity](index.md#description)), the command will return an error with the field and the shortened value: `value [...] for option [datalink.create.description] is invalid`