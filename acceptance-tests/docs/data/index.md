# Data Command Specs

The data command provides functionality for listing and archiving data sources and stores available to a specific
account on an Orchestration broker. Data sources and stores will typically be created using
the [Locker](../locker/index.md) command group and then will be used by the [Workspace](../workspace/index.md) group,
targeting individual flow nodes.

## Features

### Source Listing

The listing data source activity implies that the user is logged onto an account. It also should be validated with a
data source published from the [locker publish](../locker/index.md#data-source-publish) activity. It involves a series 
of scenarios detailed under [list account data sources](../../features/data/list_account_data_sources.feature).

```mermaid
---
title: Source Listing Activity
---
flowchart LR
    user>user] --> login([account<br>login])
    login --> lkSourcePublish([locker<br>source publish])
    login --> dtSourceList([data<br>source list])
    lkSourcePublish --> dtSourceList
```

### store listing

The listing data store activity implies that the user is logged onto an account. It also should be validated with a
data store published from the [locker publish](../locker/index.md#data-source-publish) activity. It involves a series
of scenarios detailed under [list account data stores](../../features/data/list_account_data_stores.feature).

```mermaid
---
title: Store Listing Activity
---
flowchart LR
    user>user] --> login([account<br>login])
    login --> lkStorePublish([locker<br>store publish])
    login --> dtStoreList([data<br>store list])
    lkStorePublish --> dtStoreList
```

## Commands

* [source](source.md)
* [store](store.md)

## Test Cases

* [List Account Data Sources](../../features/data/list_account_data_sources.feature)
* [List Account Data Stores](../../features/data/list_account_data_stores.feature)
