# insightsd

insightsd is a collection of utilities and system services that provide other
system services a single interface to interact with the Insights platform
from a host.

insightsd provides functionality that fall into 3 categories.

* Scheduling and running collections
  * This includes an optional redaction process over which the data of a
    collection is subjected prior to being archived and prepared for an upload.
* Uploading archives (or arbitrary payloads with a valid Content-Type)
* Updating collection cores

## Collection Modules

Collection modules are file archives that contain code and data necessary to
run a collection. They can be enabled or disabled interactively by a system
operator via the DBus API, or statically by editing the insightsd configuration
file (/etc/insightsd/insightsd.conf).

A collection module contains:

* config.ini: A config file that contains ancillary data about the collection
  module (see sample below).
* collect: An executable file that is the entrypoint to collection. This file
  is executed by `insightsd` when a collection is initiated.

A collection may contain any additional files or directories that will be made
available to the `collect` program at runtime.

### Sample config.ini
```
[Collection]
Name=foo
AutoUpdateURL=http://cloud.foo/bar/var/lib/foo.egg
Frequency=24h
```

### Running a Collection

`insightsd` will invoke a collection module's `/collect` entrypoint, passing in
a JSON object to its `stdin`. This JSON object defines parameters under which
the colleciton module is expected to operate. Examples include the destination
path to write collected data or files. A collection object must adhere to the
parameters specified in the JSON object; failure to do so will result in a
failed collection attempt.

#### Sample JSON input

```json
{
  "output_dir": "/tmp/insightsd.I1nyqpcgeX"
}
```

### D-Bus Interface

Most methods in the package library (see below) will map directly to a D-Bus
method. This design is intentional; it makes the interactions between a client,
a D-Bus server object, and the base library straightforward. For example, an
`Upload` function defined in the package library will have a corresponding
`Upload` D-Bus method, exported on the `com.redhat.insights1` D-Bus interface.

## Code Architecture

### `pkg/` - Go package implementing functional level behavior

This package implements the bulk of the insightsd functionality. While it is a
public package that downstream Go projects can consume, its primary purpose is
to provide a testable interface to `insightsd` and `insights-exec`. For example,
one could implement a custom uploader using the `insights.Upload()` functions.
But it is recommended to interact with `insightsd` through the D-Bus interface
or directly via `insights-exec`.

### `cmd/insightsd` - System daemon

This package is a program (`insightsd`) that is intended to be run on a host as
a system daemon. It implements a few D-Bus interfaces and exports objects onto
the system D-Bus for clients to interact with. `insightsd` is the main consumer
of `pkg/` and is the primary service with which clients should be interacting.

### `internal/` - Go package implementing functionality unique to insightsd

This package contains structures and functions that enable `cmd/insightsd`
and/or `cmd/insights-exec`, but aren't necessarily useful to a consumer at the
package level. For example, this package contains the XML interface files as
well as the source code files that `insightsd` uses to implement interfaces and
export objects onto the system D-Bus.

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
