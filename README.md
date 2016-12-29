# Frugal

Frugal is an extension of [Apache Thrift](https://thrift.apache.org/) which
provides additional functionality. Specifically, it includes support for
request headers, request multiplexing, thread safety, and code-generated
pub/sub APIs. Frugal is intended to act as a superset of Thrift, meaning it
implements the same functionality as Thrift with some additional
features. For a more detailed explanation, see the
[documentation](documentation).

Currently supported languages are Go, Java, Dart, and Python (2.7 and 3.5).

## Installation

### Download

Pre-compiled binaries for OS X and Linux are available from the Github
releases tab. Currently, adding these binaries is a manual process. If
a downloadable release is missing, notify the messaging team to have it
added.

### From Source

1.  Install [go](https://golang.org/doc/install) and setup [`GOPATH`](https://github.com/golang/go/wiki/GOPATH).
1.  Install [godep](https://github.com/tools/godep).
1.  Get the frugal source code

    ```bash
    $ go get github.com/Workiva/frugal
    ```

    Or you can manually clone the frugal repo

    ```bash
    $ mkdir -p $GOPATH/src/github.com/Workiva/
    $ cd $GOPATH/src/github.com/Workiva
    $ git clone git@github.com:Workiva/frugal.git
    ```

1.  Install frugal with godep
    ```bash
    $ cd $GOPATH/src/github.com/Workiva/frugal
    $ godep go install
    ```

When generating go, be aware the frugal go library and the frugal compiler
have separate dependencies.

## Using the Thrift compiler (DEPRECATED)
You may use the thrift compiler to generate parts of the internal code used by frugal using the --gen_with_frugal=false flag.

To do so, install Thrift. Dart support has not yet been released for Thrift, so we use a fork for the time-being.

### From Source

Clone the fork
```bash
git clone git@github.com:stevenosborne-wf/thrift.git
cd thrift
git checkout 0.9.3-wk-3
```

Configure the install (Note: you make need to install build dependencies)
```bash
./bootstrap.sh
./configure --without-perl --without-php --without-cpp --without-nodejs --enable-libs=no --enable-tests=no --enable-tutorial=no PY_PREFIX="$VIRTUAL_ENV"
```

Install Thrift
```
make
make install
```

### From Homebrew

Add the Workiva tap:

```
brew tap Workiva/workiva git@github.com:Workiva/homebrew-workiva.git
```

then install Thrift:

```
brew install Workiva/workiva/thrift
```

Expect the build to take about 3 minutes.

## Usage

Define your Frugal file which contains your pub/sub interface, or *scopes*, and
Thrift definitions.

```thrift
# event.frugal

// Anything allowed in a .thrift file is allowed in a .frugal file.
struct Event {
    1: i64 ID,
    2: string Message
}

// Scopes are a Frugal extension for pub/sub APIs.
scope Events {
    EventCreated: Event
}
```

Generate the code with `frugal`. Currently, only Go, Java, Dart, and Python are
supported.

```
$ frugal -gen=go event.frugal
```

By default, generated code is placed in a `gen-*` directory. This code can then
be used as such:

```go
// publisher.go
func main() {
    conn, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        panic(err)
    }

    var (
        protocolFactory  = frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
        transportFactory = frugal.NewFNatsScopeTransportFactory(conn)
        provider         = frugal.NewFScopeProvider(transportFactory, protocolFactory)
        publisher        = event.NewEventsPublisher(provider)
    )
    publisher.Open()
    defer publisher.Close()

    event := &event.Event{ID: 42, Message: "Hello, World!"}
    if err := publisher.PublishEventCreated(frugal.NewFContext(""), event); err != nil {
        panic(err)
    }
}
```

```go
// subscriber.go
func main() {
    conn, err := nats.Connect(nats.DefaultURL)
    if err != nil {
        panic(err)
    }

    var (
        protocolFactory  = frugal.NewFProtocolFactory(thrift.NewTBinaryProtocolFactoryDefault())
        transportFactory = frugal.NewFNatsScopeTransportFactory(conn)
        provider         = frugal.NewFScopeProvider(transportFactory, protocolFactory)
        subscriber       = event.NewEventsSubscriber(provider)
    )

    _, err = subscriber.SubscribeEventCreated(func(ctx *frugal.FContext, e *event.Event) {
        fmt.Println("Received event:", e.Message)
    })
    if err != nil {
        panic(err)
    }

    wait := make(chan bool)
    log.Println("Subscriber started...")
    <-wait
}
```

### Prefixes

By default, Frugal publishes messages on the topic `<scope>.<operation>`. For
example, the `EventCreated` operation in the following Frugal definition would
be published on `Events.EventCreated`:

```thrift
scope Events {
    EventCreated: Event
}
```

Custom topic prefixes can be defined on a per-scope basis:

```thrift
scope Events prefix foo.bar {
    EventCreated: Event
}
```

As a result, `EventCreated` would be published on `foo.bar.Events.EventCreated`.

Prefixes can also define variables which are provided at publish and subscribe
time:

```thrift
scope Events prefix foo.{user} {
    EventCreated: Event
}
```

This variable is then passed to publish and subscribe calls:

```go
var (
    event = &event.Event{ID: 42, Message: "hello, world!"}
    user  = "bill"
)
publisher.PublishEventCreated(frugal.NewFContext(""), event, user)

subscriber.SubscribeEventCreated(user, func(ctx *frugal.FContext, e *event.Event) {
    fmt.Printf("Received event for %s: %s\n", user, e.Message)
})
```

### Generated Comments

In Thrift, comments of the form `/** ... */` are included in generated code. In
Frugal, to include comments in generated code, they should be of the form `/**@
... */`.

```thrift
/**@
 * This comment is included in the generated code because
 * it has the @ sign.
 */
struct Foo {}

/**@ This comment is included too. */
service FooService {
    /** This comment isn't included because it doesn't have the @ sign. */
    Foo getFoo()
}
```

### Vendoring Includes

Frugal does not generate code for includes by default. The `-r` flag is
required to recursively generate includes. If `-r` is set, Frugal generates the
entire IDL tree, including code for includes, in the same output directory (as
specified by `-out`) by default. Since this can cause problems when using a
library that uses a Frugal-generated object generated with the same IDL in two
or more places, Frugal provides special support for vendoring dependencies
through a `vendor` annotation on includes and namespaces.

The `vendor` annotation is used on namespace definitions to indicate to any
consumers of the IDL where the generated code is vendored so that consumers can
generate code that points to it. This cannot be used with `*` namespaces since
it is language-dependent. Consumers then use the `vendor` annotation on
includes they wish to vendor. The value provided on the include-side `vendor`
annotation, if any, is ignored.

When an include is annotated with `vendor`, Frugal will skip generating the
include if `-use-vendor` is set since this flag indicates intention to use the
vendored code as advertised by the `vendor` annotation.

If no location is specified by the `vendor` annotation, the behavior is defined
by the language generator.

The `vendor` annotation is currently only supported by Go.

The example below illustrates how this works.

bar.frugal ("providing" IDL):
```thrift
namespace go bar (vendor="github.com/Workiva/my-repo/gen-go/bar")

struct Struct {}
```

foo.frugal ("consuming" IDL):
```thrift
include "bar.frugal" (vendor)

service MyService {
    bar.Struct getStruct()
}
```

```
frugal -r -gen go:package_prefix=github.com/Workiva/my-other-repo/gen-go -use-vendor foo.frugal
```

When we run the above command to generate `foo.frugal`, Frugal will not
generate code for `bar.frugal` since `-use-vendor` is set and the "providing"
IDL has a vendor path set for the Go namespace. Instead, the generated code for
`foo.frugal` will reference the vendor path specified in `bar.frugal`
(github.com/Workiva/my-repo/gen-go/bar).

### Plugins

Plugins are used to augment code-generation behavior in Frugal by exposing
various "behavior hooks" (described below) which can be implemented.
Currently, they are only supported by Java.

If a behavior hook is implemented in a loaded plugin, Frugal will call it to
supplement or replace the default behavior (depending on the hook). Note that
if plugins are used, it is the user's responsibility to ensure the plugin
generates valid/compilable code and that Frugal's compatibility with other
languages is not violated. As such, plugins should only be used to enhance
code generated by Frugal and not affect wire-level compatibility.

Since multiple plugins can be provided, some behavior hooks may be called
more than once. Hooks which can only be called once are documented. If
multiple plugins implement these hooks, precedence is determined by the
order in which the plugins are provided. For example, if plugins A, B, and C
all implement the `GenerateEnumValue` hook, only A's will be invoked.

Frugal plugins are implemented as dynamically linked libraries and supported
in Go 1.8 and above. Note that the same version of Go used to build Frugal
must be used to build the plugin. Plugins are built as follows:

```
$ go build -buildmode=plugin plugin.go
```

Plugins are then loaded using the "plugins" generator option:

```
$ frugal -gen java:plugins=plugin.so my_idl.frugal
```

#### Supported Behavior Hooks (Java):

```go
// GenerateEnumFields is called when generating enum fields, e.g.
//
// private final int value; // produced by a call to GenerateEnumFields
//
// Note that Frugal requires the int value for wire-level compatibility.
GenerateEnumFields(enum *parser.Enum) ([]java.Field, error)

// GenerateEnumConstructors is called when generating enum constructors, e.g.
//
// MyEnum(int value) {      // produced by
//     this.value = value;  // a call to
// }                        // GenerateEnumConstructors
//
// Note that Frugal requires the int value for wire-level compatibility.
GenerateEnumConstructors(enum *parser.Enum) ([]java.EnumConstructor, error)

// GenerateEnumValue is called when generating enum values, e.g.
//
// public enum MyEnum implements TEnum {
//     FOO(1), // produced by a call to GenerateEnumValue
//     BAR(2),
//     BAZ(3);
//
//     ...
//
// If multiple plugins implement this hook, only one will be invoked. Note that
// Frugal requires the int value for wire-level compatibility.
GenerateEnumValue(*parser.Enum, *parser.EnumValue) (java.EnumValue, error)

// GenerateEnumMethods is called when generating enum methods, e.g.
//
// public int getValue() {
//     return value;
// }
//
// Note that Frugal requires the int value for wire-level compatibility.
GenerateEnumMethods(enum *parser.Enum) ([]java.Method, error)
```

There are example plugins in the [examples](examples/plugins) directory.

## Thrift Parity

Frugal is intended to be a superset of Thrift, meaning valid Thrift should be
valid Frugal. File an issue if you discover an inconsistency in compatibility
with the IDL.

## Docker

### Via Shipyard

Grab the frugal Docker image id for the image you would like to use from
[Shipyard](https://shipyard.workiva.org/repo/Workiva/frugal).

Switch to the directory that has the files you would like to generate.

Then run the docker image. This command will mount your local directory into
the image. It supports all of the standard Frugal commands.

```
docker run -v "$(pwd):/data" drydock.workiva.org/workiva/frugal:{SHIPYARD_ID} frugal -gen={LANG} {FILE_TO_GEN}
```

An example to generate the Go code off the event.frugal definition in the
example directory.

```
$ cd example
$ docker run -v "$(pwd):/data" drydock.workiva.org/workiva/frugal:17352 frugal -gen=go event.frugal
```
