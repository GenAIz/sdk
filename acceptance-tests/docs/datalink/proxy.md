## Datalink Proxy

The proxy commands of Datalink are used to define outbound proxy configurations available globally to all data sources
and data stores defined for the Datalink type. An outbound proxy address and port allows the Smart Function using the
Datalink to communicate outside the network under which it runs.

This is specifically critical for testing Smart Function that are of the `Connector` type. The majority, if not all,
connector Smart Function have a role to retrieve or store data externally to the execution network.

* if the proxy commands can not find the specified datalink locally, they will return an error with the `FQDN` not
  found:
  `Error: data link [FQDN] not found`
    * either [create](create.md) or [sync](sync.md) will need to be called to bring the definition to edit locally

### proxy add

```
genaiz dk proxy add [OEM/]HANDLE[:VERSION] ADDRESS[:PORT] \
    --oem=OEM \
    --version=VERSION \
    --config-type=yaml|toml|json \
    --user-defined \
    --udp --tcp
```

#### HANDLE

* the command expects its first argument to at least be a handle, but it may be a FQDN string composed of the
  OEM/HANDLE:VERSION fields, which are all required
* if the handle value does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command
  will return an error with the key of the field and the invalid value:
  `Error value [...] for option [datalink.proxyadd.handle] is invalid`

#### OEM

* if the oem value is specified as part of the first argument, it will override the option regardless of validity.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will
  return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.proxyadd.oem] is invalid`

#### VERSION

* if the version value is specified as part of the first argument, it will override the option regardless of validity.
* if no version is specified the value will default to `1.0.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the
  command will return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.proxyadd.version] is invalid`

#### ADDRESS

* the command expects its second argument to be a valid network address. If no address is specified usage will be
  printed.
* the command will accept `'*'` as a way to specify that all outbound connections are allowed.
* parsing of the address is made with the net package in go and allows a broad range of invalid addresses to be used.
  When the command fails parsing an error is returned: `Error: address ...: too many colons in address`

> [!IMPORTANT]
> Address parsing is basic at best. The command relies on the Orchestration to validate the published proxies.

#### PORT

* the command expects its second argument to contain a port denomination, unless the [address](#address) specified is
  `'*'`. Standing for allow all connections
* if the port is not specified, the command will return as error: `Error: address ...: missing port in address`

#### config-type

* if config-type is not specified, the default is YAML
* if the config-type specified is not a recognized value, the command will return an error with the field and the value:
  `value [...] for option [datalink.proxyadd.configtype] is invalid`

> [!IMPORTANT]
> Only YAML is supported on the current early version of the SDK

#### user-defined

* user-defined is set to `TRUE` by default, which will direct the command to read and write configuration files under
  the user's .config folder
* if user-defined is explicitly set to `FALSE`, the command will use the current working directory to persist datalink
  modifications

#### tcp

* if specified, the TCP flag will be applied to the configured outbound proxy
* if not specified and the [udp](#udp) flag is set, the TCP flag will not be applied to the configured outbound proxy
* if not specified and the [udp](#udp) flag is not set, the TCP flag WILL be applied to the configured outbound proxy

#### udp

* if specified, the UDP flag will be applied to the configured outbound proxy
* if not specified and the [tcp](#tcp) flag is set, the UDP flag will not be applied to the configured outbound proxy
* if not specified and the [tcp](#tcp) flag is not set, the UDP flag will not be applied to the configured outbound
  proxy

### proxy rm

```
genaiz dk proxy rm [OEM/]HANDLE[:VERSION] ADDRESS[:PORT] \
    --oem=OEM \
    --version=VERSION \
    --config-type=yaml|toml|json \
    --user-defined
```

When a proxy address and port are not found as an outbound proxy, the command is a no-op and returns nothing.
Considering that what was asked to be removed does not make the state of the Data Link invalid if it does not exist.

#### HANDLE

* the command expects its first argument to at least be a handle, but it may be a FQDN string composed of the
  OEM/HANDLE:VERSION fields, which are all required
* if the handle value does not match a valid handle string (see [handle validity](index.md#handle-and-oem)), the command
  will return an error with the key of the field and the invalid value:
  `Error value [...] for option [datalink.proxyrm.handle] is invalid`

#### OEM

* if the oem value is specified as part of the first argument, it will override the option regardless of validity.
* if the resolved oem does not match a valid oem string (see [oem validity](index.md#handle-and-oem)), the command will
  return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.proxyrm.oem] is invalid`

#### VERSION

* if the version value is specified as part of the first argument, it will override the option regardless of validity.
* if no version is specified the value will default to `1.0.0`
* if the resolved version does not match a valid SemVer version string (see [version validity](index.md#version)), the
  command will return an error with the key of the field and the invalid value:
  `Error: value [...] for option [datalink.proxyrm.version] is invalid`

#### ADDRESS

* the command expects its second argument to be a valid network address. If no address is specified usage will be
  printed.
* the command will accept `'*'` as a way to specify that all outbound connections are allowed.
* parsing of the address is made with the net package in go and allows a broad range of invalid addresses to be used.
  When the command fails parsing an error is returned: `Error: address ...: too many colons in address`

#### PORT

* the command expects its second argument to contain a port denomination, unless the [address](#address) specified is
  `'*'`. Standing for allow all connections
* if the port is not specified, the command will return as error: `Error: address ...: missing port in address`

#### config-type

* if config-type is not specified, the default is YAML
* if the config-type specified is not a recognized value, the command will return an error with the field and the value:
  `value [...] for option [datalink.proxyrm.configtype] is invalid`

> [!IMPORTANT]
> Only YAML is supported on the current early version of the SDK

#### user-defined

* user-defined is set to `TRUE` by default, which will direct the command to read and write configuration files under
  the user's .config folder
* if user-defined is explicitly set to `FALSE`, the command will use the current working directory to persist datalink
  modifications
