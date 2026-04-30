# Account Command Specs

## Test Cases

* [Account login with username](../../features/account/login_username.feature)
* [Account login with OIDC](../../features/account/login_oidc.feature)

## Commands

* [login](#login)
* [logout](#logout)

### Login

`genaiz account login HOST --username=USERNAME --refresh --no-browser`

The command manages an .auth file under the ~/.cache/genaiz folder. It records the token used to make requests to one or several broker services.

It will always prompt for a password if a username is used. When no username is used the command will attempt to log in a user using the Broker's OIDC provided urls. If the broker does not support OIDC, login can not work without a username and password.

#### HOST

* mandatory
* defaults to https://HOST if the protocol is not provided
* default to http://localhost if the HOST is localized and the build targets the dev environment `go build -tags dev`
* include any ports with the string if necessary. By default, it won't add any

#### username

* if omitted the command will attempt a login with an OIDC url provided by the broker
* if no OIDC url is provided by the broker, the command will prompt for a username
* any string will do even empty `--username=`, the validity of it belongs to the orchestrator.

#### refresh

* the flag is used to skip validation of the current session an obtain a new one
* refreshing a non-existent session should not prevent a login
* refreshing a session for Broker A should not affect the session of Broker B

### Logout

`genaiz account logout --host=HOST --username=USERNAME`

The command removes sessions from the ~/.cache/genaiz folder, deleting the session records from the Broker as well.

* if the command is invoked without any options, it'll remove the active session only

#### host

* if the host is not specified, the command removes the first session for the specified username it can find
* the host string needs to match given the environment's default protocol rule. A prod build will assume https for all host strings, dev will only consider http for localhost.

#### username

* if the username is not specified the command removes the first session for the specified host it can find
* the username string needs to match the username used to call login.