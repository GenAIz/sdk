## Locker Init

```
genaiz lk init [PATH] --overwrite|-o --update|-u
```

Locker init is the command used to initialize a locker on the provided path. The command can overwrite the locker if the
path already exists. It can also read the locker and re-write it using an updated password.

By default, the command will always confirm updating the path if it already exists, overwriting it on a separate
confirmation if no update is wanted.

If the command is used to update an existing locker, the new passphrase given must not be the same as the old one. The
command will fail with an error if it is: `Error: the passphrase must be different`

If the passphrase provided does not match a valid [passphrase](index.md#passphrase), the command will return an error:
`Error: the passphrase must be 8 characters long and contain capital, small letters, digits at least one special character`

### PATH

The path of the locker file to create or update. By default, the path of the locker file is assumed to be
`$HOME/.config/genaiz/locker.bin`

* If the path exists and can not be read, the command will return an error: `Error: the locker [...] can not be read`
* If the path does not exist and can not be written to, the command will return an error:
  `Error: the locker [...] can not be written`
* If the path exists, but both [update](#update) and [overwrite](#overwrite) are not wanted, the command will return an
  error: `Error the locker [...] is already initialized`

### overwrite

Overwrite indicates the user wants to write a new locker file even if it already exists.

* If the [path](#path) resolved exists, but overwrite and [update](#update) are not specified, the command will
  confirm [update](#update) only.
* If overwrite is wanted at the same time as [update](#update), overwrite is implied with the update, therefor only the
  update is performed.
* If the [path](#path) resolved does not exist and overwrite is true, the command just completes normally without any
  confirmation required.

### update

Update indicates the user wants to update the password of the locker and re-encrypt the locker's content.

* If the [path](#path) resolved exists, but update and [overwrite](#overwrite) are not specified, the command will
  confirm the update only.
* If update is wanted at the same time as [overwrite](#overwrite), update has priority and overwrite is simply ignored.
* If the [path](#path) resolved does not exist and update is true, the command just completes normally without any
  confirmation required.
* If update is wanted but the supplied old password can not decrypt the locker file, the command returns an error:
  `Error: the passphrase failed decryption`
