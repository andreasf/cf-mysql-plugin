# Cloud Foundry CLI MySQL Plugin

[![Build Status](https://travis-ci.org/andreasf/cf-mysql-plugin.svg?branch=master)](https://travis-ci.org/andreasf/cf-mysql-plugin)

cf-mysql-plugin makes it easy to connect the `mysql` command line client to any MySQL-compatible database used by
Cloud Foundry apps. Use it to

* inspect databases for debugging purposes
* manually adjust schema or contents in development environments
* dump and restore databases

## Usage

### Geting a list of available databases

Running the plugin without arguments should give a list of available MySQL databases:

```bash
$ cf mysql
MySQL databases bound to an app:

my-db
```

Databases are *available* if they are bound to a running app - see below for an explanation why.


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

## Installing and uninstalling

Download a binary release or build yourself by running `go build`. Then, install the plugin with

```bash
$ cd /path/to/plugin
$ cf install-plugin ./cf-mysql-plugin
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
