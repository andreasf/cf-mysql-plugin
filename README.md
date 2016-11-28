# Cloud Foundry CLI MySQL Plugin

[![Build Status](https://travis-ci.org/andreasf/cf-mysql-plugin.svg?branch=master)](https://travis-ci.org/andreasf/cf-mysql-plugin)

cf-mysql-plugin makes it easy to connect the `mysql` command line client to any MySQL-compatible database used by
Cloud Foundry apps. Use it to

* inspect databases for debugging purposes
* manually adjust schema or contents in development environments
* dump and restore databases

## Usage

Get a list of available databases:

```bash
$ cf mysql
```

Databases are *available* if they are bound to a running app - see below for an explanation why.

Connect to the database `my-db`:

```bash
$ cf mysql my-db
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
