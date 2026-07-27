# Account Command Specs

## Test Cases

* [Account activation](../../features/account/activate_oidc_account.feature)
* [Account session inspect](../../features/account/inspect_user_account.feature)
* [Account listings](../../features/account/list_accounts_username_oidc.feature)
* [Account login with username](../../features/account/login_username.feature)
* [Account login with OIDC](../../features/account/login_oidc.feature)

## Commands

* [activate](#activate)
* [inspect](#inspect)
* [list](#list)
* [login](#login)
* [logout](#logout)

### Activate

`genaiz account activate HOST --username=USERNAME`

The command will activate a session if it can find one for the specified host and username. Note that activating a
session that is expired, does trigger a login request. If the intent is always to use a valid session [login](#login)
should be invoked.

If the command activates finds a session that is no longer valid, the command returns an error:
`Error session is expired`.

#### HOST

The host name to look for in the current session file.

* The command will fail with printing usage if there are no host argument specified.
* If no session can be found the command will return an error: `Error: could not elect a session`

#### USERNAME

Optionally a username for which session to look for. Since we should support multiple user accounts per brokerage,
username can be used to distinguish between duplicated hostnames.

* if the session targeted is an OIDC session, the username will be matched with only the first 10 characters of the
  token
* If no session can be found the command will return an error: `Error: could not elect a session`

### List

`genaiz account list [[USER_STRING@]HOST_STRING] --json`

The command will list currently known account sessions with their creation and expiry times. It should also indicate
with a color supported terminal which session are expired and which is the currently active session.

If the command does not find any results it should print `Error: No sessions found`

#### USER_STRING

* the `USER_STRING` and/or `@` symbols are purely optional. The command will test the string as a prefix or suffix to
  both the username and the full session name
* should not typically contain spaces. It could be interpreted as several arguments will be returned as an error with
  the command usage

> [!NOTE]
> When a session was established through OIDC, the username is unknown and the string is substituted with the first 10
> characters of the access token used.

#### HOST_STRING

* the `HOST_STRING` can be a prefix, a suffix or the fully qualified domain name of the host providing the account.
* since we match the string on all of 'username', 'host address' and 'session name' fields, if a username is also a host
  address, this can provide some false positives in the filtering results. Use `@` as a prefix or suffix to refine the
  search.

#### json

* if the command finds no accounts, an empty list is returned.
* boolean flag which will produce the output list as a JSON array of `account.UserAccount`

```json
[
  {
    "active": true,
    "username": "test_user",
    "hostAddr": "dev.genaiz.com",
    "created": "09:03:09",
    "expiry": "2026-06-19"
  }
]
```

### Inspect

`genaiz account inspect [HOST]`

The command allows inspection of the first session found for the specified host or for the active session under
`$HOME/.cache/genaiz`. It will also support overrides `GENAIZ_AUTH_URL` and `GENAIZ_AUTH_SESSION`, which can be used in
CI/CD environments to keep tokens from spilling out of Pipelines onto disk.

### Login

`genaiz account login HOST --username=USERNAME --refresh --no-browser`

The command manages an `.auth` file under the `$HOME/.cache/genaiz` folder. It records the token used to make requests
to one or several broker services.

It will always prompt for a password if a username is used. When no username is used the command will attempt to log in
a user using the Broker's OIDC provided urls. If the broker does not support OIDC, login can not work without a username
and password.

The command will activate a valid session to the provided HOST, if that session is not the active one under
`$HOST/.cache/genaiz/.auth` and is not expired.

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

`genaiz account logout [HOST] --username=USERNAME`

The command removes sessions from the ~/.cache/genaiz folder, deleting the session records from the Broker as well.

* if the command is invoked without any argument or option, it'll remove the active session only

#### HOST

* if the host is not specified, the command removes the first session for the specified username it can find
* the host string needs to match given the environment's default protocol rule. A prod build will assume https for all
  host strings, dev will only consider http for localhost.

#### username

* if the username is not specified the command removes the first session for the specified host it can find
* the username string needs to match the username used to call login.
* if the session was established with an OIDC token we usually don't have the username, but the command can
  auto-complete with a token prefix.
