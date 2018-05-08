# Cloud Foundry CLI MySQL Plugin

[![Build Status](https://travis-ci.org/andreasf/cf-mysql-plugin.svg?branch=master)](https://travis-ci.org/andreasf/cf-mysql-plugin)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://github.com/andreasf/cf-mysql-plugin/blob/master/LICENSE)

cf-mysql-plugin makes it easy to connect the `mysql` command line client to any MySQL-compatible database used by
Cloud Foundry apps. Use it to

* inspect databases for debugging purposes
* manually adjust schema or contents in development environments
* dump and restore databases

## Contents

* [Usage](#usage)
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
   Get a list of available databases:
   cf mysqldump

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
# download dependencies
go get -v ./...
go get github.com/onsi/ginkgo
go get github.com/onsi/gomega
go install github.com/onsi/ginkgo/ginkgo

# run tests and build
ginkgo -r
go build
```

## Details

### Obtaining credentials

cf-mysql-plugin gets credentials from service bindings, which are only available when your database services are bound
to a started app. If you don't currently have an app running, try the following to start an nginx app:

```bash
TEMP_DIR=`mktemp -d`
pushd $TEMP_DIR
touch Staticfile
cf push static-app -m 64M --no-route
popd
rm -r $TEMP_DIR
```

Then, bind the database to your app with:

```
cf bind-service static-app database-name
```

Using service keys would be an alternative to service bindings. I decided against service keys, because they need to
be deleted before a service can be deleted, making service administration more difficult.
