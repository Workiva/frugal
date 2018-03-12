# Frugal

![Build Status](https://travis-ci.org/Workiva/frugal.svg?branch=develop)

Frugal is an extension of [Apache Thrift](https://thrift.apache.org/) which
provides additional functionality. Key features include:

- request headers
- request multiplexing
- request interceptors
- per-request timeouts
- thread-safe clients
- code-generated pub/sub APIs
- support for Go, Java, Dart, and Python (2.7 and 3.5)

Frugal is intended to act as a superset of Thrift, meaning it
implements the same functionality as Thrift with some additional
features. For a more detailed explanation, see the
[documentation](documentation).

## Installation

### Homebrew

```bash
brew install frugal
```

### Download

Pre-compiled binaries for OS X and Linux are available from the Github
releases tab. Currently, adding these binaries is a manual process. If
a downloadable release is missing, notify the messaging team to have it
added.

If go is already installed and setup you can also simply:

```bash
$ go get github.com/Workiva/frugal
```

### From Source

1.  Install [go](https://golang.org/doc/install) and setup [`GOPATH`](https://github.com/golang/go/wiki/GOPATH).
1.  Clone the frugal repo

    ```bash
    $ mkdir -p $GOPATH/src/github.com/Workiva && cd $_
    $ git clone git@github.com:Workiva/frugal.git
    ```

1.  Install the CLI binary
    ```bash
    $ cd $GOPATH/src/github.com/Workiva/frugal
    $ go install
    ```

When generating go, be aware the frugal go library and the frugal compiler
have separate dependencies.

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

### Annotations

Annotations are extra directive in the IDL that can alter the way code is generated.
Some common annotations are listed below

| Annotation    | Values        | Allowed Places | Description
| ------------- | ------------- | -------------- | -----------
| vendor        | Optional location | Namespaces, Includes | See [vendoring includes](#vendoring-includes)
| deprecated    | Optional description | Service methods | Marks a method as deprecated (if supported by the language) and logs a warning if the method is called.

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
include if `use_vendor` language option is set since this flag indicates
intention to use the vendored code as advertised by the `vendor` annotation.

If no location is specified by the `vendor` annotation, the behavior is defined
by the language generator.

The `vendor` annotation is currently only supported by Go and Dart.

The example below illustrates how this works.

bar.frugal ("providing" IDL):
```thrift
namespace go bar (vendor="github.com/Workiva/my-repo/gen-go/bar")
namespace dart bar (vendor="my-repo/gen-go")

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
frugal -r -gen go:package_prefix=github.com/Workiva/my-other-repo/gen-go,use_vendor foo.frugal
```

When we run the above command to generate `foo.frugal`, Frugal will not
generate code for `bar.frugal` since `use_vendor` is set and the "providing"
IDL has a vendor path set for the Go namespace. Instead, the generated code for
`foo.frugal` will reference the vendor path specified in `bar.frugal`
(github.com/Workiva/my-repo/gen-go/bar).


## Thrift Parity

Frugal is intended to be a superset of Thrift, meaning valid Thrift should be
valid Frugal. File an issue if you discover an inconsistency in compatibility
with the IDL.
