## Function Proxy

Proxy is a command for managing the list of outbound proxies configured on connectors.

Outbound proxies are only configurable on Connectors, the commands listed under will return an error if used on a Smart Function of another type: `Error: outbound proxies can only be configured for connector functions`

### proxy add

```
genaiz sf proxy add HOST:PORT  \
  --tcp --udp --inactive
```

#### host

* if the host string does not match a valid domain name (validity based on pattern `^[a-zA-Z0-9.-]*$`), the command will return an error for the field with the invalid value: `Error: value [...] is invalid for option [function.publish.proxyadd.host] is invalid`

#### port

* if the port specified is not higher than 0 and below 65536, the command will return an error for the field with the invalid value: `Error: value [...] is invalid for option [function.publish.proxyadd.port] is invalid`

#### tcp

* if specified, the TCP flag will be applied to the configured outbound proxy
* if not specified and the [udp](#udp) flag is set, the TCP flag will not be applied to the configured outbound proxy
* if not specified and the [udp](#udp) flag is not set, the TCP flag WILL be applied to the configured outbound proxy

#### udp

* if specified, the UDP flag will be applied to the configured outbound proxy
* if not specified and the [tcp](#tcp) flag is set, the UDP flag will not be applied to the configured outbound proxy
* if not specified and the [tcp](#tcp) flag is not set, the UDP flag will not be applied to the configured outbound proxy

#### inactive

* if inactive is not specified, the outbound proxy will be configured with the active flag
* when inactive is specified, the active flag will not be applied to the proxy

### proxy rm

```
genaiz sf proxy rm HOST:PORT
```

#### host

* to allow removal of invalid outbound proxies, the host will not be validated
* if the host does not exist, the command will simply print a success as the host is not configured regardless of changes

#### port

* to allow removal of invalid outbound proxies, the port will not be validated
* if the port does not exist for its host, the command will simply print a success as the host is not configured regardless of changes
