# Locker Command Specs

Locker is used as a command initiator for all commands use to manipulate data lockers. A data locker is a binary
encrypted file containing [data source and store](../data/index.md) definitions. Locker files should always be treated
as black boxes. No properties from a locker file, secret or not, should ever be printed using the CLI.

Not all locker file interactions require usage of an account. All definitions, however are tied to
Orchestration [accounts](../account/index.md). To create a data source or a data store, the `locker` command will
require a valid [datalink](../datalink/index.md) from a session established beforehand.

* [Features](#features)
    * [Locker Initialization](#locker-initialization)
    * [Data Source Addition](#data-source-addition)
    * [Data Source Update](#data-source-update)
    * [Data Source Update for a Secret](#data-source-update-for-a-secret)
* [Test Cases](#test-cases)
* [Commands](#commands)
* [Validation](#validation)
    * [Passphrase](#passphrase)

## Features

### Locker Initialization

Locker initialization can be used to create a new locker file or to change the passphrase of an existing locker file.
The command will walk the user through STDIN if it finds an existing locker. When the user wants to refresh the
passphrase, the command will ask first the existing passphrase and second for a new passphrase.

```mermaid
---
title: Locker Initialization
config:
    flowchart:
        titleTopMargin: 50 
---
flowchart LR
    user>user] --> lkInit([init locker])
    lkInit --> lkUpdate([locker update?])
    lkInit --> lkCreate
    lkUpdate --> lkOverwrite([locker overwrite?])
    lkUpdate --> lkRefresh([locker refresh<br>passphrase])
    lkOverwrite --> lkCreate([locker create])
    lkRefresh --> lkCreate
```

### Data Source Addition

The data source addition activity is the localized update of an existing locker file. The activity requires a request
done to an Orchestration [Account](../account/index.md). A data source is always added for a
specific [Datalink](../datalink/index.md), which needs to be validated from the list of links available to the account
session.

```mermaid
---
title: Data Source Addition
config:
    flowchart:
        titleTopMargin: 50 
---
flowchart LR
    user>user] --> lkInit([init locker])
    user --> login([account<br>login])
    lkInit --> addDataSource([add<br>data source])
    login --> lsDataLink([list<br>datalinks])
    lsDataLink --> addDataSource
```

### Data Source Update

The data source update activity is the localized update of an existing data source under an existing locker file. The
activity does not require a request to any [Account](../account/index.md), but still requires to know under which
account to look for a data source handle. The data sources already enshrined inside a locker are considered valid until
the [Data Source Creation](../data/index.md#data-source-creation) activity is invoked.

```mermaid
---
title: Data Source Update
config:
    flowchart:
        titleTopMargin: 50 
---
flowchart LR
    user>user] --> updateDataSource([update<br>data source])
```

### Data Source Update for a Secret

The data source update activity above can not be used to update secret properties from text-based sources like the
command line. Secrets must live unencrypted for the shortest amount of time possible when an update or when
a [Create Data Source](../data/index.md#data-source-creation) activity is performed. Secrets can only be read through
STDIN when the update command is invoked.

The recommended activity for updating a secret is as follows:

```mermaid
---
title: Data Source Update for a Secret
config:
    flowchart:
        titleTopMargin: 50 
---
flowchart LR
    user>user] --> gpgDecrypt([gpg decrypt])
    gpgDecrypt --> updateDataSourceSecret([update<br>data source secret])
```

## Test Cases

* [Add data source to locker](../../features/locker/add_data_source_to_locker.feature)
* [Initialize locker](../../features/locker/init_locker.feature)
* [Update data source secret](../../features/locker/update_data_source_secret.feature)

## Commands

* [init](init.md)
* [source add](source.md#source-add)
* [source update](source.md#source-update)

## Validation

### Passphrase

* Passphrases need to be at least 8 characters long
* Required characters are capital and small letters, digits and special characters in the range of
  `!"#$%&'()*+,-./:;<=>?[\]^_{|}~`
