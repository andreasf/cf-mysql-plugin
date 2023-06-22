# Cloud Foundry CLI MySQL Plugin

[![build-and-test](https://github.com/andreasf/cf-mysql-plugin/actions/workflows/build-and-test.yml/badge.svg)](https://github.com/andreasf/cf-mysql-plugin/actions/workflows/build-and-test.yml)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/andreasf/cf-mysql-plugin/blob/master/LICENSE)

cf-mysql-plugin makes it easy to connect the `mysql` command line client to any MySQL-compatible database used by
Cloud Foundry apps. Use it to

* inspect databases for debugging purposes
* manually adjust schema or contents in development environments
* dump and restore databases

## Contents

* [Usage](#usage)
* [Removing service keys](#removing-service-keys)
* [Installing and uninstalling](#installing-and-uninstalling)
* [Building](#building)
* [Details](#details)

## Usage

```bash
$ cf mysql -h
NAME:
   mysql - Connect to a MySQL database service

USAGE:
   Open a mysql client to a database:
   cf mysql <service-name> [mysql args...]


$ cf mysqldump -h
NAME:
   mysqldump - Dump a MySQL database

USAGE:
   Dumping all tables in a database:
   cf mysqldump <service-name> [mysqldump args...]

   Dumping specific tables in a database:
   cf mysqldump <service-name> [tables...] [mysqldump args...]
```

### Connecting to a database

Passing the name of a database service will open a MySQL client:

```bash
$ cf mysql my-db
Reading table information for completion of table and column names
You can turn off this feature to get a quicker startup with -A

Welcome to the MariaDB monitor.  Commands end with ; or \g.
Your MySQL connection id is 1377314
Server version: 5.5.46-log MySQL Community Server (GPL)

Copyright (c) 2000, 2016, Oracle, MariaDB Corporation Ab and others.

Type 'help;' or '\h' for help. Type '\c' to clear the current input statement.

MySQL [ad_67fd2577d50deb5]> 
```

### Piping queries or dumps into `mysql`

The `mysql` child process inherits standard input, output and error. Piping content in and out of `cf mysql` works
just like it does with plain `mysql`:

```bash
$ cat database-dump.sql | cf mysql my-db
```

### Passing arguments to `mysql`

Any parameters after the database name are added to the `mysql` invocation:

```bash
$ echo "select 1 as foo, 2 as bar;" | cf mysql my-db --xml
<?xml version="1.0"?>

<resultset statement="select 1 as foo, 2 as bar" xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance">
  <row>
        <field name="foo">1</field>
        <field name="bar">2</field>
  </row>
</resultset>
```

### Dumping a database

Running `cf mysqldump` with a database name will dump the whole database:

```bash
$ cf mysqldump my-db --single-transaction > dump.sql
```

### Dumping individual tables

Passing table names in addition to the database name will just dump those tables:

```bash
$ cf mysqldump my-db table1 table2 --single-transaction > two-tables.sql
```

## Removing service keys

The plugin creates a service key called 'cf-mysql' for each service instance a user connects to. The keys are reused
when available and never deleted. Keys need to be removed manually before their service instances can be removed:

```bash
$ cf delete-service -f somedb
Deleting service somedb in org afleig-org / space acceptance as afleig@pivotal.io...
FAILED
Cannot delete service instance. Service keys, bindings, and shares must first be deleted.
```
Deleting the service failed. The CLI hints at service keys and app bindings that might still exist.
```bash
$ cf service-keys somedb
Getting keys for service instance somedb as afleig@pivotal.io...

name
cf-mysql
```
A key called 'cf-mysql' is found for the service instance 'somedb', because we have used the plugin with 'somedb'
earlier. After removing the key, the service instance can be deleted:

```bash
$ cf delete-service-key -f somedb cf-mysql
Deleting key cf-mysql for service instance somedb as afleig@pivotal.io...
OK

$ cf delete-service -f somedb
Deleting service somedb in org afleig-org / space acceptance as afleig@pivotal.io...
OK
```

This behavior might change in the future as it's not optimal to leave a key around.

## Installing and uninstalling

The easiest way is to install from the repository:

```bash
$ cf install-plugin -r "CF-Community" mysql-plugin
```

You can also download a binary release or build yourself by running `go build`. Then, install the plugin with

```bash
$ cf install-plugin /path/to/cf-mysql-plugin
```

The plugin can be uninstalled with:

```bash
$ cf uninstall-plugin mysql
```

## Building

```bash
# install ginkgo test runner
go install github.com/onsi/ginkgo/v2/ginkgo

# run tests and build
ginkgo -r
go build
```

## Details

### Obtaining credentials

cf-mysql-plugin creates a service key called 'cf-mysql' to obtain credentials. It no longer retrieves credentials from
application environment variables, because with the introduction of [CredHub](https://github.com/cloudfoundry-incubator/credhub/blob/master/docs/secure-service-credentials.md),
service brokers can decide to return a CredHub reference instead.

The service key is currently not deleted after closing the connection. It can be deleted by running:

```
cf delete-service-key service-instance-name cf-mysql
```

A started application instance is still required in the current space for setting up an SSH tunnel. If you don't
have an app running, try the following to start an nginx app:

```bash
TEMP_DIR=`mktemp -d`
pushd $TEMP_DIR
touch Staticfile
cf push static-app -m 128M --no-route
popd
rm -r $TEMP_DIR
```
