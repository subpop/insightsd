# insightsd

insightsd is a collection of utilities and system services that provide other
system services a single interface to interact with the Insights platform
from a host.

insightsd provides functionality that fall into 3 categories.

* Scheduling and running collections
* Uploading archives (or arbitrary payloads with a valid Content-Type)
* Updating collection cores

## Collection Modules

Collection modules are file archives that contain code and data necessary to
run a collection. They can be enabled or disabled interactively by a system
operator via the DBus API, or statically by editing the insightsd configuration
file (/etc/insightsd/insightsd.conf).

A collection module contains:

* config.ini: A config file that contains ancillary data about the collection
  module (see sample below)
* collect: An executable file that is the entrypoint to collection. This file
  is executed by `insightsd` when a collection is initiated.

A collection may contain any additional files or directories that will be made
available to the `collect` program at runtime.

### Sample config.ini
```
[Collection]
Name=foo
AutoUpdateURL=http://cloud.foo/bar/var/lib/foo.egg
ContentType=application/tar-gz
```

### Collection Module Index

An index of available collection modules is available for fetch and parsing by
insightsd. This index is visible to system administrators via the DBus API (and
this via a client), with features that allow modules to be enabled or disabled.
Enabled modules sync the module package (see above) to the host and run data
collections on the specified interval.

## Code Architecture

### `lib/` - Go package implementing functional level behavior

This package implements the bulk of the insightsd functionality. While it is a
public package that downstream Go projects can consume, its primary purpose is
to provide a testable interface to `insightsd` and `insights-exec`. For example,
one could implement a custom uploader using the `insights.Upload()` functions.
But it is recommended to interact with `insightsd` through one of the IPC
interfaces or directly via `insights-exec`.

### `cmd/insightsd` - System daemon

This package is a program, `insightsd` that is intended to be run on a host as
a system daemon. It implements a few DBus interfaces and exports objects onto
the system DBus for clients to interact with. `insightsd` is the main consumer
of `lib/` and is the primary service with which clients should be interacting.

### `svc/dbus` - DBus interfaces describing the various operations of insightsd

This package contains the XML interface files as well as the stub service source
code files that `insightsd` uses to implement interfaces and export objects onto
the system DBus.

#### Collector

Methods used to request collection execution and manage collection scheduling.

* `Enable` - registers a collection operation to be executed on the provided
  frequency.
* `Show` - details about a current collection operation.
* `Edit` - edit a collection operation.
* `Disable` - deregisters a collection operation from the schedule.
* `List` - get a list of scheduled collections.
* `Collect` - trigger an ad-hoc collection execution.

#### Uploader

Methods used to upload a payload to the platform.

* `Upload` - initiate an upload.

#### Updater

Methods used to request core update and scheduling.

* `Show` - details about a core update.
* `Edit` - edit a core update.
* `List` - get a list of all registered core updates.
* `Update` - trigger an ad-hoc check for updates for a given core update.

### `svc/grpc` - gRPC interfaces implementing the various operations of insightsd

Much like the DBus interface, a gRPC interface can be vended, allowing clients
to interact with `insightsd` over gRPC.

### `cmd/inctl` - CLI client to `insightsd`

`inctl` is a client that provides a CLI for interacting with the `insightsd`
system service. This utility can be used by users or other programs on the
system to:

* register collection events upon installation or activation
* examine the current state of a collection or update
* trigger a collection
* trigger a collection followed by an upload
* trigger a core update, followed by a collection, followed by an upload

### `cmd/insights-exec` - oneshot entrypoint to perform a single ad-hoc operation

This utility provides a CLI to perform a single operation immediately, in process.
It is important to note that this utility *does not* interact with `insightsd`.
This makes it particularly well suited for containerized environments where 
running a full system daemon is not feasible. All `insightsd` operations are
executed in process (of course, due to the nature of a collection event, it is
still executed as a subprocess of `insights-exec`).

For example, to upload a pre-existing archive:

`insights-exec upload --content-type foo /var/tmp/foo.tar.gz`
