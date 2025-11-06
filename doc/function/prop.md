# Function Prop

Prop is a command for managing the property specifications of a Smart Function. These specifications are translated into **environment variables**, when the function runs locally or remotely.

Typically, specifying parameters on a Smart Function involves either defining the values in a .env file which is sourced before its container is created, or specifying the variable on the command line with the [env](run.md#env) option.

Environment variables which are not recognized as specified properties are filtered out from the execution both locally and remotely. It is then crucial to provide a specification for each and every parameter required by the Smart Function.

> [!NOTE]
> It is not required to provide any specification for the [built-in](index.md#environment) environment variables

## prop add

```
genaiz sf prop add KEY --type=bool|double|int|string|enum \
  --name=NAME --description=DESC --default-value=VALUE \
  --enum-value=ENUM_VALUE_1 --enum-value=ENUM_VALUE_2
```

Property add is used to define a new property specification for the Smart Function in the current directory.

* The default value should always be validated against its type. A double spec could not validate a default string value.

### KEY

* if the key does not match a valid Key string (see [KEY validity](#valid-keys)), the command will return an error with the offending key: `Error: [...] is not a valid environment key`
* if the key is already used by another specification, an error will be returned with the offending key: `Error: the key [...] already exists`

### type

* if the type is not specified the default type is "string"
* if the type does not match a valid type string (see [type validity](#valid-types))

### name

* if the name does not match a valid name string (see [name validity](#valid-names)), the command will return an Error: `Error: value [...] for option [sf.publish.propspecadd.name] is invalid`
* if the resolved name is empty, name will default to the value [KEY](#key)

### description

* if the description does not match a valid description string (see [description validity](#valid-description)), the command will return an Error: `Error: value [...] for option [sf.publish.propspecadd.description] is invalid`
* a description is optional and will be left empty if not specified

### default-value

* a default value must always be valid according to the specified type
* if default value is left empty, it will be interpreted as no default value and the property must be specified. 

### enum-value

* can be specified multiple times or be a comma separated string of values
* if an enum value does not match a valid enum string (see [enum validity](#valid-enum-values)), the command will return an error with the key of the field and the invalid value; `Error: value [n:...] for option [sf.publish.propspecadd.enumvalue] is invalid`

## prop edit

```
genaiz sf prop edit KEY --name=NAME --description=DESC \
  --default-value=VALUE --add-enum-value=VAL --rm-enum-value=VAL \
  --enum-value=ENUM_VALUE_1 --enum-value=ENUM_VALUE_2
```

Property edit is used to modify characteristics for a specification other than the type, which is not editable. To change the type of specification, the property spec must be removed first, and then redefined.

There is a relationship between [enum-value](#enum-value-1), [add-enum-value](#add-enum-value) and [rm-enum-value](#rm-enum-value). Normal usage of those will gravitate around using each option individually, but since those are options they can be used together. In this case know that the priority of interpretation is the following:

1. Replace all enum values with [enum-value](#enum-value-1) if specified.
2. Add all enum values from [add-enum-value](#add-enum-value) if specified.
3. Remove all enum values specified in [rm-enum-value](#rm-enum-value) if any.

### KEY

* if the key does not exist in the current property specifications, the command returns an error: `Error: the key [...] could not be found`

### name

* if the name is not specified, the field will not be modified
* if the name does not match a valid name string (see [name validity](#valid-names)), the command will return an Error: `Error: value [...] for option [sf.publish.propspecedit.name] is invalid`

### description

* if the description is not specified, the field will not be modified
* if the description does not match a valid description string (see [description validity](#valid-description)), the command will return an Error: `Error: value [...] for option [sf.publish.propspecedit.description] is invalid`

### default-value

* if the default-value is not specified, the field will not be modified
* if specified, a default value must always be valid according to the specified type
  * replaced, added or removed enum values should be considered for this validity.

### add-enum-value

* can be specified multiple times or be a comma separated string of values
* if omitted, no new value will be added to the current set of enum values
* if an enum value does not match a valid enum string (see [enum validity](#valid-enum-values)), the command will return an error with the key of the field and the invalid value; `Error: value [n:...] for option [sf.publish.propspecedit.enumvalue] is invalid`
* adding an enum value to a non-enum property specification will return an error: `Error: the property spec type does not allow enum values`

### rm-enum-value

* can be specified multiple times or be a comma separated string of values
* if omitted, no value will be removed from the current set of enum values
* removing an enum value which is not a current enum value, **will not yield an error**. This is because the state of the specification is still valid and will not contain what the user wanted to remove
* removing an enum value from a non-enum property specification will return an error: `Error: the property spec type does not allow enum values`

### enum-value

* can be specified multiple times or be a comma separated string of values
* if omitted, the values of the current property specification will be unchanged
* if specified, the enum values of the current property specification will be replaced by the new set of enum values.
* replacing enum values from a non-enum property specification will return an error: `Error: the property spec type does not allow enum values`

## prop rm

```
genaiz sf prop rm KEY 
```

Removes a property by KEY from the property specifications of a function. The key string is not validated as it does not need to be for the removal to be successful, valid key or not. This is also done to avoid messy migrations if the KEY syntax changes to be more restrictive.

### KEY

* if the key does not exist in the current property specifications, the command **will not return an error**. Specifically because the key is not present in the prop specs to begin and the state of the function is still valid after execution.

## prop env

```
genaiz sf prop env KEY VALUE --context=PATH --env-file=FILE_PATH
```

The env command provides a utility for setting a property specification to a certain value using an environment file or the default file sourced by the [run](run.md), [start](start.md) and [test](test.md) commands. The value is validated against the specification and the command can add or replace variables.

### KEY

* if the KEY specified is not part of the property specifications for the Smart Function, the command will return an error: `Error proprety specification for key [...] is not defined`

### VALUE

* if the VALUE specified is not valid according to the property type, the command will return an error with the type specified: `Error: illegal TYPE value`

### context

context is a globally accepted options, which changes the working directory before applying the command. In this case, it would cause the [env-file](#env-file) specified to be written relatively to the context path.

* if the context is not specified, the default is the current working directory.
* if the context path is invalid, the command will return an error with the field and the invalid value; `Error: value [...] for option [sf.env.context] is invalid`

> [!NOTE]
> The env command will look for the Genaiz.yaml under the working directory, it is then necessary to use the global --config option, if --context is used, in order for the property specifications to be validated.

### env-file

* if env-file is not specified, the default is .env under the current [context](#context)
* if env-file does not exist, the command will create it if the context exists
* if env-file can not be read or created due to permissions, the command will return an error: `Error: open ...: permission denied`

## Specification Validation

### Valid Keys

* The key of a property specification mirrors the string used to define conventional Environment Variables.
* It must be composed only of capitalized alphanumeric characters and underscores; `[A-Z_][A-Z0-9_]*`
* Keys can not be expanded into other keys. That is you can not define a key using the value of another one. For example, MY_KEY_$KEY_INDEX is not a valid key.

### Valid Names

* A name can contain any kind of characters for as long as it does not extend to more than 255 characters.

### Valid Description

* A description can extend to up to 4096 characters.

### Valid Types

* Only "STRING", "INT", "BOOL", "DOUBLE" and "ENUM" are allowed.
* Lower case strings will be capitalized on publishing and writing.

### Valid Enum Values

* Enum values must have between 1 and 512 characters