<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**

- [What is Frugal?](#what-is-frugal)
  - [Why does Frugal exist?](#why-does-frugal-exist)
    - [Problems with Thrift](#problems-with-thrift)
    - [Frugal Features](#frugal-features)
- [Frugal Architecture](#frugal-architecture)
  - [Transports](#transports)
  - [Protocols](#protocols)
  - [The IDL and Compiler](#the-idl-and-compiler)
  - [User-Defined Types](#user-defined-types)
  - [Service Stubs](#service-stubs)
    - [Client Stubs](#client-stubs)
    - [Server Stubs](#server-stubs)
- [Designing Frugal Structs](#designing-frugal-structs)
  - [Namespaces](#namespaces)
  - [Constants](#constants)
  - [Structs](#structs)
  - [Base Types](#base-types)
  - [Typedefs](#typedefs)
  - [Field Ids](#field-ids)
  - [Enums](#enums)
  - [Collections](#collections)
  - [Unions](#unions)
  - [Requiredness and Optional Fields](#requiredness-and-optional-fields)
  - [Type Evolution](#type-evolution)
    - [Renaming Fields](#renaming-fields)
    - [Adding Fields](#adding-fields)
    - [Deleting Fields](#deleting-fields)
    - [Changing a Field’s Type](#changing-a-field%E2%80%99s-type)
    - [Changing a Field’s Requiredness](#changing-a-field%E2%80%99s-requiredness)
    - [Changing a Field’s Default Value](#changing-a-field%E2%80%99s-default-value)
- [Designing Frugal Services](#designing-frugal-services)
  - [Declaring Services](#declaring-services)
    - [Parameters](#parameters)
    - [Return Types](#return-types)
    - [Oneway Functions](#oneway-functions)
  - [Evolving Services](#evolving-services)
- [Designing Frugal Scopes](#designing-frugal-scopes)
  - [Declaring Scopes](#declaring-scopes)
  - [Prefixes](#prefixes)
- [Glossary](#glossary)
  - [FAsyncCallback*](#fasynccallback)
  - [FContext](#fcontext)
  - [FProcessor](#fprocessor)
  - [FProcessorFactory](#fprocessorfactory)
  - [FProcessorFunction*](#fprocessorfunction)
  - [FProtocol](#fprotocol)
  - [FProtocolFactory](#fprotocolfactory)
  - [FScopeProvider](#fscopeprovider)
  - [FScopeTransport*](#fscopetransport)
  - [FScopeTransportFactory*](#fscopetransportfactory)
  - [FServer](#fserver)
  - [FServiceProvider](#fserviceprovider)
  - [FSubscription](#fsubscription)
  - [FRegistry*](#fregistry)
  - [FTransport](#ftransport)
  - [FTransportFactory](#ftransportfactory)
  - [FTransportMonitor](#ftransportmonitor)
  - [Scope](#scope)
  - [Service](#service)
  - [ServiceMiddleware](#servicemiddleware)
- [Protocol](#protocol)

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

# What is Frugal?

Frugal is an extension of [Apache Thrift](https://thrift.apache.org/) which
provides additional features and improvements. Conceptually, Frugal is a
superset of Thrift, meaning valid Thrift is also valid Frugal (there are some
caveats to this).

Frugal makes use of several core parts of Thrift, including protocols and
transports. This means most of the components that ship with Thrift "just work"
out of the box. Frugal wraps many of these components to extend their
functionality.

See the [glossary](glossary.md) for definitions of Frugal concepts.

## Why does Frugal exist?

Frugal was created to address many of Thrift's shortcomings without completely
reinventing the wheel. Thrift is a solid, mature RPC framework used widely in
production systems. However, it has several key problems.

### Problems with Thrift

- Head-of-line blocking: a single, slow request will block any following
  requests for a client.

- Out-of-order responses: an out-of-order response puts a Thrift transport in a
  bad state, requiring it to be torn down and reestablished. E.g. if a slow
  request times out at the client, the client issues a subsequent request, and
  a response comes back for the first request, the client blows up.

- Concurrency: a Thrift client cannot be shared between multiple threads of
  execution, requiring each thread to have its own client issuing requests
  sequentially. This, combined with head-of-line blocking, is a major
  performance killer.

- RPC timeouts: Thrift does not provide good facilities for per-request
  timeouts, instead opting for a global transport read timeout.

- Request headers: Thrift does not provide support for request metadata, making
  it difficult to implement things like authentication and authorization.
  Instead, you are required to bake these things into your IDL. The problem
  with this is it puts the onus on service providers rather than allowing an
  API gateway or middleware to perform these functions in a centralized way.

- Middleware: Thrift does not have any support for client or server middleware.
  This means clients must be wrapped to implement interceptor logic and
  middleware code must be duplicated within handler functions. This makes it
  impossible to implement AOP-style logic in a clean, DRY way.

- RPC-only: Thrift has limited support for asynchronous messaging patterns, and
  even asynchronous RPC is largely language-dependent and susceptible to the
  head-of-line blocking and out-of-order response problems.

### Frugal Features

Frugal was built to address these concerns. Below are some of the things it
provides.

- Request multiplexing: client requests are fully multiplexed, allowing them to
  be issued concurrently while simultaneously avoiding the head-of-line
  blocking and out-of-order response problems. This also lays some groundwork
  for asynchronous messaging patterns.

- Thread-safety: clients can be safely shared between multiple threads in which
  requests can be made in parallel.

- Pub/sub: IDL and code-generation extensions for defining pub/sub APIs in a
  type-safe way.

- Request context: a first-class request context object is added to every
  operation which allows defining request/response headers and per-request
  timeouts. By making the context part of the Frugal protocol, headers can be
  introspected or even injected by external middleware. This context could be
  used to send OAuth2 tokens, distributed tracing metadata, and user-context
  information, avoiding the need to include it everywhere in your IDL and
  handler logic. Correlation IDs are also built into the request context for
  tying logs, traces, and anytics together.

- Middleware: client- and server- side middleware is supported for RPC and
  pub/sub APIs. This allows you to implement interceptor logic around handler
  functions, e.g. for authentication, logging, or retry policies.

# Frugal Architecture

At a high-level, Frugal is organized into several layers as in Figure 1.
The layers highlighted in yellow represent application code that is
written by a user. The portions in red represent code generated by the
Frugal compiler from an interface definition defined in an IDL file. The
layers in orange are portions of Frugal available as library code imported
into your application as a dependency. Lastly, the device layer in blue
represents the physical device transmitting messages.

![Frugal Architecture](./images/02-Frugal/FrugalArchitecture.png)

To understand the Frugal architecture requires understanding how each of
the layers in Figure 1 interact with one another. Let’s figure that out by
discussing each of the layers in turn.

## Transports

The Frugal transport layer provides an abstraction over physical devices.
This allows Frugal to mask device specific details and provide a common
API for upper layers that need to use physical devices. Specifically, the
transport layer handles byte-level communication with the underlying
device. By providing this abstraction, Frugal allows for new devices or
middleware to be supported without any impact to the rest of the Frugal
architecture.

Transports are further organized into composable layers called
a *transport stack*. This functionality allows you to abstract a physical
device using a particular transport, then wrap that abstraction with
another logical transport layer such as buffering or encryption, without
changing the interface to transport consumers. Furthermore, the
application is free to select or change the transport stack at compile
time or run time, allowing you to define a flexible transport stack
suitable for different needs.

## Protocols

Protocols provide the means for serializing data types into byte streams
for use by transports. Frugal does not support every type in every
language. Rather, it supports a basic type system that can be converted
into representations in each language. Any valid Frugal protocol
implementation must be capable of reading and writing all types defined by
the Frugal type system (specified by the Frugal interface definition
language, or IDL).

The protocol layer sits on top of a transport stack and has the
responsibility of serializing language specific data types into
language-agnostic byte streams capable of being transmitted using
a transport stack. *The client and server are required to use the same
serialization protocol for proper communication*.

## The IDL and Compiler

The IDL is designed to make describing an applications data types and
interfaces language independent. This language independent representation
is then used by the Frugal compiler to generate language specific
implementations of the data types and interfaces for use by a user
application.

![Frugal Compiler](./images/02-Frugal/IDLCompiler.png)

## User-Defined Types

While protocols are responsible for serializing at the individual data
type level, the IDL provides support for user-defined data types, structs,
and interfaces, allowing for cross-language serialization of rich
structured data and interfaces. The Frugal compiler uses the IDL to
generate code that automates the serialization of these more complex data
structures using particular protocols.

## Service Stubs

The compiler is also responsible for generating service stubs used by
clients and servers to fulfill any interfaces defined by the IDL. This
provides Frugal users support for cross language RPC. The service stubs
can be divided into two parts: client and server.

### Client Stubs

The client stub acts as a proxy for the remote service. The client process
calls the proxy to interact with the service and the proxy handles
communication with the underlying protocol and transport. The protocol and
transport are responsible for serializing the messages and receiving back
the results.

### Server Stubs

The server stub uses the protocol and transport stack to deserialize
incoming method calls, providing hooks into user-defined code for
processing messages as they arrive. The result of these method calls are
then serialized again using the same protocol and transport stack and
delivered back to the client.

The user-defined code is responsible for implementing the service
interface. This code is called by the generated service processor for each
incoming request.

The full request-response life cycle is shown in Figure 3.

![Frugal Compiler](./images/02-Frugal/RPC.png)

# Designing Frugal Structs

Frugal services are described using the Interface Definition Language
(IDL). This section describes how to use the IDL to develop your messaging
service.

In Frugal, types define the structure of data exchanged through service
interfaces. Frugal comes with several base types like enums, strings,
collections, and common numeric types. Using these base types, you are
free to design your own higher-level structures, called *user defined
types*, using Frugal structs.

To develop some intuition on how to define your own user defined types, we
will work through an example that uses each of the available Frugal base
types. Imagine an application that records observations observed by
a radio telescope. In particular, for each observation it records the
strength and source of radio waves.

Let’s assume that our radio astronomy observations consist of the
following fields:

* The position of the object observed
* The time of the observation
* The number of telescopes used to make the observation
* The magnitude of radio waves detected over each of several frequencies
* Which telescope system recording the measurement
* A visible spectrum bitmap image of the sky at the time of the
  observation

There are several possible data types which can be used to define the
position of the object observed by the radio telescope. Here is the IDL
describing the system.

```thrift
// Radio Telescope Observation Interface

namespace * radio_observation

const string Version = "1.0.0"

//Mutiple position types are used by the telescopes supported
struct EarthRelPosition {
    1: double latitude
    2: double longitude
    3: double elevation
}
struct RelVector {
    1: EarthRelPosition pos
    2: double declination
    3: double azimuth
}
struct ICRFPosition {
    1: double right_ascension
    2: double declination
    3: optional i16 ecliptic_year
}

/**
 * The focal point of an observation. May be described by
 * one of several underlying types.
 */
union Position {
    1: EarthRelPosition erpos
    2: RelVector rv
    3: ICRFPosition icrfpos
}

/**
 * Time in seconds and fractions of seconds since Jan 1, 1970.
 */
typedef double Time

/**
 * Radio Telescopes and arrays making observations.
 */
enum RadioObservationSystem {
    Parkes  = 1
    Arecibo = 2
    GMRT    = 17
    LOFAR   = 18
    Socorro = 25
    VLBA    = 51
}

/**
 * An observation made by a radio telescope.
 */
struct RadioObservation {
    1: i32 telescope_count
    2: Time time
    //3: string Researcher; retired
    4: RadioObservationSystem system
    5: map<i64,double> freq_amp
    6: Position pos
    7: optional binary sky_bmp
}
```

The IDL above is one of many possible solutions to the interface
requirements for our radio telescope observations. The types provided in
our IDL capture a number of important design decisions. Many of these
decisions are driven by a desire to ensure that our interface supports the
requirements in a way that allows the UDT to be efficient and flexible.
Let’s look at each of the design choices in detail.

## Namespaces

```thrift
namespace * radio_observation
```

The first non-comment line of our IDL declares the wildcard (`*`)
namespace “radio_observation”. Namespace declarations must be listed
before any services, types or constants are declared. The asterisk
indicates that the namespace should be used for all output languages
generated by the IDL Compiler.

It is a good idea to place all of your interface definitions in
a descriptive namespace. Specifying a namespace keeps all of the names
created in your IDL out of the global scope when you generate code in most
languages, reducing the opportunity for name collisions.

## Constants

```thrift
const string Version = "1.0.0"
```

The IDL allows multiple versions of an interface to interoperate. It can
be useful to know which interface version each program is using. This
version string can be accessed programmatically through a “Version”
constant. This is purely a user defined construct and the Frugal framework
takes no notice of it.

The interface evolution features of Frugal will automatically provide
backwards and forwards compatibility if used correctly. That said, this
constant allows us to quickly identify which version of the interface we
are using and it can be logged easily.

## Structs

```thrift
struct RadioObservation {
    ...
}
```

Frugal structs are used to define new types represented by a packaged
group of fields. Conceptually, structs are the tool used to represent
messages, objects, records and any other grouping needed by an interface.

The RadioObservation struct is the focus of this IDL and the next several
topics describe the features of our RadioObservation struct.

## Base Types

The first field in our struct is `telescope_count`. Simple value fields
are typically represented with base types.

```thrift
struct RadioObservation {
    1: i32 telescope_count
    ...
}
```

This field will store the number of radio telescopes used to make the
observation. An integer type is a good fit for our telescope count and i32
strikes a good balance between size (4 bytes) and assurance that we will
be able to capture the count of even the largest telescope array. The IDL
does not support unsigned integers. Semantics such as (0==unknown) and (<
0 is illegal) should be documented in the IDL when not explicit in the
type declaration.

## Typedefs

Typedefs allow a new type to be created from a preexisting type. The
“Time” type in our radio_observation IDL is an example.

```thrift
typedef double Time

struct RadioObservation {
    ...
    2: Time time
    ...
}
```

If you are designing an interface with a semantic type implemented in
terms of a base type, but the semantic type is particularly significant or
widely used it may be a good candidate for a typedef. Typedef types are
self-documenting and, in statically typed languages, ensure that the
underlying type is not used accidentally in places where the typedef type
is required.

## Field Ids

All of the fields defined in a struct or union should be given a positive
16 bit integer Id. Once assigned, the Id should never be reassigned for
the life of that type.

```thrift
struct RadioObservation {
    ...
    //3: string researcher; retired
    ...
}
```

In our example, field #3 has been retired. Ids can be retired safely but
should not be forgotten. The reason for this is that older code relying on
prior versions of the interface will have semantic expectations associated
with old Ids. In our example, older versions of this interface used field
3 to store a string with the Researcher name. Deleting field three and
then reusing it to describe something else would be very confusing to an
older program still expecting field 3 to represent Researcher. By leaving
the field comment in the IDL source we can ensure that people extending
the interface at a later time will not reuse the Id value.

You may also wish to leave the field in and prefix the name with
`OBSOLETE_`. If you then mistakenly reuse the same field Id, compilation
of your IDL will fail, preventing the reused field Id from being
deployed.

## Enums

Enums create a new type with a discrete set of possible values, usually
more naturally described with human language rather than integers.

```thrift
enum RadioObservationSystem {
    Parkes = 1
    Arecibo = 2
    GMRT    = 17
    LOFAR   = 18
    Socorro = 25
    VLBA    = 51
}

struct RadioObservation {
    ...
    4: RadioObservationSystem system
    ...
}
```

When you are dealing with stable sets and the names of the elements are
more important than the numbers used to represent them, an enum is usually
a good option.

## Collections

```thrift
struct RadioObservation {
    ...
    5: map<i64,double> freq_amp
    ...
}
```

The IDL provides set, list, and map types for representing collections of
data. In our example type, we use a map to capture frequency/amplitude
pairs from the observation. The IDL allows any type to be used as a key in
maps and sets. However, integers, strings and enums are usually the best
choice as they provide the most consistent inter-language operability.

## Unions

Unions single purpose is to allow a type to change.

```thrift
union Position {
    1: EarthRelPosition erpos
    2: RelVector rv
    3: ICRFPosition icrfpos
}

struct RadioObservation {
    ...
    6: Position pos
    ...
}
```

Unions are declared just like structs but all fields are implicitly
optional and only one field may be set at a time.

In our example we are faced with supplying a position for our observation
but have several ways in which the position might be expressed. By
representing the position as a union we can use any of the position types
to capture the position of our observation and, should the need arise, we
can add new types in the future. The presence of a union tells developers
using this interface several things. First it tells them that the type of
the position field is not constant. This implies that they must supply
support for various position types. Second, and more subtle, it tells them
that new position types may be added, so it is possible that they may
recover a position type they do not understand. This later insight will
encourage union users to write code which degrades favorably in the face
of unknown types and/or to support plugins which allow new type support to
be added dynamically.

## Requiredness and Optional Fields

All fields in a struct have requiredness trait. Normal field declarations
are said to have default requiredness. Fields can also be declared as
“required” or “optional”. In our example the optional modifier is applied
to a bitmap field.

```thrift
struct RadioObservation {
    ...
    7: optional binary sky_bmp
}
```

In scenarios where you would like the ability to serialize a field but
only some of the time, making the field optional is a good choice. By
making the field optional all of the users of the interface will know
about it but only those requiring it need to serialize it.

In most situations default requiredness is a good compromise, because it
is always serialized but need not be present during deserialization. If
you would like the flexibility to decide whether a field is serialized or
not, choose optional requiredness. Required requiredness fields create
runtime errors when not found during deserialization and cannot evolve,
they should be avoided unless you want to permanently enforce a field’s
presence.

## Type Evolution

Interface evolution is one of the most important features of the IDL. Type
evolution enables us to change types over time without breaking
compatibility with preexisting programs. With proper planning, you can
safely change almost any aspect of a user-defined type exceptfor Ids. So
how might a type evolve over time? There are many possibilities:

* The name of a field may need to be changed
* A new field may be required
* An existing field may no longer be required
* The type of a field may need to be changed
* The requiredness of the field may need to be changed
* The default value of a field may need to be changed

Let’s consider each of these. In our discussion we will look at the impact
an IDL change has on programs on either side of the change. Programs using
the old IDL before the change will be referred to as OLD and programs
using the new IDL after the change will be referred to as NEW.

### Renaming Fields

Fields can be renamed at anytime without impacting interoperability.This
is because protocols do not serialize field names. Instead, fields are
identified in serialized form by the combination of the field Id and field
type. The Id is the field’s unique identifier. The type is used to ensure
that the writer and the reader are using the same type.

### Adding Fields

Perhaps the most common change facing evolving data types is the need to
add more fields. Fortunately, adding a field to a structure can also be
a backward compatible operation. During deserialization, fields which are
not recognized are ignored, this allows OLD programs to tolerate unknown
newly added fields. NEW programs must also tolerate not receiving the new
field when deserializing data from OLD programs.

New fields should *not* be made required. Unless all systems using the UDT
with the new required field will be updated at once, required fields will
cause exceptions in NEW programs when receiving copies of the UDT from OLD
programs which do not provide the new required field. The required flag
should not be used when adding fields to an interface.

### Deleting Fields

As we saw in our Radio Observation example, the proper way to delete
a field is to comment it out and retire its Id permanently. When NEW
programs send UDTs to OLD programs the OLD programs must tolerate the
absence of fields which have been deleted in the NEW IDL. This often
requires the field to have a default value assigned to it in OLD IDL
versions.

If OLD programs cannot tolerate the absence of a deleted field the field
should be marked as required. Required fields must never be deleted
because they are expected to be present by OLD programs.

Deleted field Ids should never be reused.

### Changing a Field’s Type

Changing the type of a field will cause the field to be ignored by
programs expecting a different type. When a program reads a field it
checks the field Id and the field type. If these match a field that the
program knows the field is deserialized. If these do not match, the field
is skipped. Thus, changing the type of a field will cause OLD programs to
no longer recognize the field when transmitted by NEW programs and it will
cause NEW programs to no longer recognize the field when transmitted by
OLD programs. However, NEW programs can share the field amongst themselves
and OLD programs can share the field amongst themselves. While precarious,
you may find this an acceptable way to migrate programs to a new field
type.

There is a more effective option but it requires preplanning. If you know
that a field may have more than one type representation, it makes sense to
make it of union type.

### Changing a Field’s Requiredness

Fields with Required requiredness should not be changed. Making a Required
field Optional will cause exceptions when OLD programs find the field
missing.

In situations where Required fields need to be removed or made optional,
changing the field from Required to Default can be a stepping stone to
removal since languages always serialize default values. If you ensure
that NEW programs using Default requiredness all serialize the field you
will not break the OLD programs requiring the field. This can give you
time to update the OLD programs to the NEW schema gradually. Once all of
the programs are running the NEW schema based on Default requiredness you
can transition the field to Optional or delete it altogether. To do this
you must have control over all of the programs using the interface (rarely
possible with public APIs).

### Changing a Field’s Default Value

Default values are best used with default requiredness fields as a way to
enable the field to be added or deleted. When adding default requiredness
fields, a default value provides a rational value when the field is not
provided by OLD programs. When deleting default requiredness fields with
a default value, OLD programs will use the default value when the field is
not provided by NEW programs. Other effects of default value changes are
subtle and application defined.

# Designing Frugal Services

Frugal services interfaces are declared in a Frugal IDL and then compiled
by the Frugal compiler, which generates client and server stubs. The
compiler generates two main components: (1) a client that exposes the
service interface for RPC calls and a matching (2) processor that services
RPC calls from a client.

## Declaring Services

Services are declared using the `service` keyword. Services have a name
and a set of functions. Each function has a name, return type, and a set
of parameters

```thrift
service SocialLookup {
    string GetSiteByRank( 1: i32 rank )
    i32 GetSiteRankByName( 1: string name )
}
```

### Parameters

Parameters are declared the same way as fields of a struct, having
a numeric identifier, a requiredness, a type, a name, and a default value.

Parameter Ids are used by the Frugal framework to uniquely identify
function parameters during RPC processing. Ids are 16 bit integers and
must be unique within a parameter list. Ids are optional, however,
removing them greatly complicates the process of making incremental
changes to the interface.

Parameter requiredness is similar to struct requiredness. Parameters can
be assigned one of two requiredness levels:

1. required - must always be present and may never be changed or removed
2. default - callers must supply the parameter but servers should not
   require it. This allows parameters to be added or deleted at some point
   in the future.

Parameters can be any valid IDL type other than `void`.

### Return Types

Any legal IDL type can be returned by a function. `void` indicates no
return value.

### Oneway Functions

Oneway functions send data to the server but do not receive anything back.

```thrift
service SocialLookup {
    ...
    oneway void UpdateSiteUsers(1: string name, 2: i32 users);
}
```

Because oneway functions do not receive a response from the server, it is
impossible to know when or if a call completed. There are two benefits to
oneway functions. First, they cut the number of messages exchanged in
half. Second, they return to the client as soon as possible, allowing the
client to continue with other work. Under the right circumstances, oneway
functions can be a useful asset.

## Evolving Services

Frugal allows you to make incremental changes without breaking existing
applications. Here are some of the common modifications supported by
Frugal.

* Adding a parameter to a function
    * OLD clients can call NEW servers if default values are provided for
      the new parameter.
    * NEW clients can call OLD servers, old servers will ignore the
      parameter.
* Removing a parameter from a function
    * OLD clients can call NEW servers which will ignore the deleted
      parameter.
    * NEW clients *cannot* call OLD servers, unless the removed parameter
      provided a default value.
* Adding functions
    * OLD clients can call NEW servers.
    * NEW clients *cannot* call OLD servers, unless prepared to handle not
      implemented exceptions.
* Removing functions.
    * OLD clients *cannot* call NEW servers, unless prepared to handle not
      implemented exceptions.
    * NEW clients can call OLD servers.

The only client hostile modification in this list is removing a function.
A common approach to dealing with this scenario is to deprecate functions
first, then remove them in a later release. This provides a transition
period for clients to adopt the newer interface.

# Designing Frugal Scopes

Frugal scopes are declared in a Frugal IDL and then compiled by the Frugal
compiler, which generates client and server stubs. The compiler generates
two main components: (1) a publisher that exposes the interface for
publishing messages to a topic and a matching (2) subscriber that
subscribes to a topic and processes incoming messages.

Scopes are designed to provided asynchronous pub/sub messaging through the
Messaging Platform.

## Declaring Scopes

Scopes are declared using the `scope` keyword. Scopes are named and
declare key-value pairs specifying the name of the message to publish and
data type that is published. For example, the following IDL declares an
Events scope with a single publisher for signalling event creation.
Semantically, the publisher publishes the entire Event struct for each
event that has been created.

```thrift
struct Event {
    1: i64 ID,
    2: string Message
}

scope Events {
    EventCreated: Event
}
```

## Prefixes

By default, Frugal publishes messages on the topic `<scope>.<operation>`.
For example, the `EventCreated` operation in the following Frugal definition
would be published on the `Events.EventCreated` topic:

```thrift
scope Events {
    EventCreated: Event
}
```

Custom topic prefixes can be defined on a per-scope basis:

```thrift
scope Events prefix lux.doc {
    EventCreated: Event
}
```

As a result, `EventCreated` would be published on `lux.doc.Events.EventCreated`.

Prefixes can also define variables which are provided at publish and
subscribe time:

```thrift
scope Events prefix lux.doc.{membershipID} {
    EventCreated: Event
}
```

This variable is then passed to publish and subscribe calls:

```go
var (
    event         = &event.Event{ID: 42, Message: "hello, world!"}
    membershipID  = "ALKJ123LK90"
)
publisher.PublishEventCreated(frugal.NewFContext(""), event, membershipID)

subscriber.SubscribeEventCreated(user, func(ctx *frugal.FContext, e *event.Event) {
    fmt.Printf("Received event for %s: %s\n", membershipID, e.Message)
})
```

All messages being routed through the Messaging Frontend must be prefixed
with `{accountrid}.{membershiprid}`.

Clients may specify a membershipID of “global” to receive messages from
all members. Clients may specify an accountID of
“unsecure_global_broadcast” to receive messages from all accounts. If
specifying an unsecure_global_broadcast, membershipID must also be global.
Absolutely no customer data should be pushed over global topics.

# Glossary

This describes at a high level some of the concepts found in Frugal. Most
components in Frugal are prefixed with "F", i.e. "FTransport", in order to
differentiate from Thrift, which prefixes things with "T". Components marked
with an asterisk are internal details of Frugal and not something a user
interacts with directly but are documented for posterity. As a result, some
internal components may vary between language implementations.

## FAsyncCallback*

FAsyncCallback is an internal callback which is constructed by generated code
and invoked by an FRegistry when a RPC response is received. In other words,
it's used to complete RPCs. The operation ID on FContext is used to look up the
appropriate callback. FAsyncCallback is passed an in-memory TTransport which
wraps the complete message. The callback returns an error or throws an
exception if an unrecoverable error occurs and the transport needs to be
shutdown.

## FContext

FContext is the context for a Frugal message. Every RPC has an FContext, which
can be used to set request headers, response headers, and the request timeout.
The default timeout is five seconds. An FContext is also sent with every publish
message which is then received by subscribers.

In addition to headers, the FContext also contains a correlation ID which can
be consumed by the user and used for tying logs, traces, and anytics together.
A random correlation ID is generated for each FContext if one is not provided.
As an intermediate node in a chain of requests, it is best practice to pass
along a clone of the incoming request context using the provided `Clone`
function.

FContext also plays a key role in Frugal's multiplexing support. A unique,
per-request operation ID is set on every FContext before a request is made.
This operation ID is sent in the request and included in the response, which is
then used to correlate a response to a request. The operation ID is an internal
implementation detail and is not exposed to the user.

An FContext should belong to a single request for the lifetime of that request.
The 'Clone' function is provided to allow for downstream use of the same
context, forwarding on headers, such as correlation ID, trace ID, etc.

## FProcessor

FProcessor is Frugal's equivalent of Thrift's TProcessor. It's a generic object
which operates upon an input stream and writes to an output stream.
Specifically, an FProcessor is provided to an FServer in order to wire up a
service handler to process requests.

## FProcessorFactory

FProcessorFactory produces FProcessors and is used by an FServer. It takes a
TTransport and returns an FProcessor wrapping it.

## FProcessorFunction*

FProcessorFunction is used internally by generated code. An FProcessor
registers an FProcessorFunction for each service method. Like FProcessor, an
FProcessorFunction exposes a single process call, which is used to handle a
method invocation.

## FProtocol

FProtocol is Frugal's equivalent of Thrift's TProtocol. It defines the
serialization protocol used for messages, such as JSON, binary, etc. FProtocol
actually extends TProtocol and adds support for serializing FContext. In
practice, FProtocol simply wraps a TProtocol and uses Thrift's built-in
serialization. FContext is encoded before the TProtocol serialization of the
message using a simple binary protocol. See the
[protocol documentation](protocol.md) for more details.

## FProtocolFactory

FProtocolFactory creates new FProtocol instances. It takes a TProtocolFactory
and a TTransport and returns an FProtocol which wraps a TProtocol produced by
the TProtocolFactory. The TProtocol itself wraps the provided TTransport. This
makes it easy to produce an FProtocol which uses any existing Thrift transports
and protocols in a composable manner.

## FScopeProvider

FScopeProvider is used exclusively for pub/sub and produces FScopeTransports
and FProtocols for use by pub/sub scopes. It does this by wrapping an
FScopeTransportFactory and FProtocolFactory.

## FScopeTransport*

FScopeTransport extends Thrift's TTransport and is used exclusively for pub/sub
scopes. Subscribers use an FScopeTransport to subscribe to a pub/sub topic.
Publishers use it to publish to a topic.

## FScopeTransportFactory*

FScopeTransportFactory produces FScopeTransports and is typically used by an
FScopeProvider.

## FServer

FServer is Frugal's equivalent of Thrift's TServer. It's used to run a Frugal
RPC service by executing an FProcessor on client connections. FServer can
optionally support a high-water mark which is the maximum amount of time a
request is allowed to be enqueued before triggering server overload logic (e.g.
load shedding).

Currently, Frugal includes two implementations of FServer: FSimpleServer, which
is a basic, accept-loop based server that supports traditional Thrift
TServerTransports, and FNatsServer, which is an implementation that uses
[NATS](https://nats.io/) as the underlying transport.

## FServiceProvider

FServiceProvider is the service equivalent of FScopeProvider. It produces
FTransports and FProtocols for use by RPC service clients.

## FSubscription

FSubscription is a subscription to a pub/sub topic created by a scope. The
topic subscription is actually handled by an FScopeTransport, which the
FSubscription wraps. Each FSubscription should have its own FScopeTransport.
The FSubscription is used to unsubscribe from the topic.

## FRegistry*

FRegistry is responsible for multiplexing and handling received messages.
Typically there is a client implementation and a server implementation. An
FRegistry is used by an FTransport.

The client implementation is used on the client side, which is making RPCs.
When a request is made, an FAsyncCallback is registered to an FContext. When a
response for the FContext is received, the FAsyncCallback is looked up,
executed, and unregistered.

The server implementation is used on the server side, which is handling RPCs.
It does not actually register FAsyncCallbacks but rather has an FProcessor
registered with it. When a message is received, it's buffered and passed to
the FProcessor to be handled.

## FTransport

FTransport is Frugal's equivalent of Thrift's TTransport. FTransport extends
TTransport and exposes some additional methods. An FTransport typically has an
FRegistry, so it provides methods for setting the FRegistry and registering and
unregistering an FAsyncCallback to an FContext. It also allows a way for
setting an FTransportMonitor and a high-water mark provided by an FServer.

FTransport wraps a TTransport, meaning all existing TTransport implementations
will work in Frugal. However, all FTransports must use a framed protocol,
typically implemented by wrapping a TFramedTransport.

Most Frugal language libraries include an FMuxTransport implementation, which
uses a worker pool to handle messages in parallel.

## FTransportFactory

FTransportFactory produces FTransports by wrapping a provided TTransport.

## FTransportMonitor

FTransportMonitor watches and heals an FTransport. It exposes a number of hooks
which can be used to add logic around FTransport events, such as unexpected
disconnects, expected disconnects, failed reconnects, and successful
reconnects.

Most Frugal implementations include a base FTransportMonitor which implements
basic reconnect logic with backoffs and max attempts. This can be extended or
reimplemented to provide custom logic.

## Scope

Scopes do not map directly to an actual object but are an important concept
within Frugal. A scope is defined in a Frugal IDL file, and it specifies a
pub/sub API. Each scope has one or more operations, each of which define a
pub/sub event. Frugal takes this definition and generates the corresponding
publisher and subscriber code.

The pub/sub topic, which is an implementation detail of the scope, is
constructed by Frugal and consists of the scope and operation names. However, a
scope prefix can be specified, which is prepended to the topic. This prefix can
have user-defined variables, allowing runtime subscription matching.

## Service

Services do not map directly to an actual object but, like scopes, are an
important concept. A service is defined in a Frugal IDL file, and it specifies
a RPC API. Each service has one or more methods which can be invoked remotely.
Frugal takes this definition and generates the corresponding client and server
interface.

## ServiceMiddleware

ServiceMiddleware is used to implement interceptor logic around API calls. This
can be used, for example, to implement retry policies on service calls,
logging, telemetry, or authentication and authorization. ServiceMiddleware can
be applied to both RPC services and pub/sub scopes.

# Protocol

This describes the binary protocol used to encode FContext by an FProtocol.

FProtocol serializes FContext headers using a custom protocol before the normal
serialization of the Thrift message, as produced by TProtocol. FProtocol is a
framed protocol, meaning the length of the serialized message, or frame, is
prepended to the frame itself. As such, a serialized Frugal message looks like
the following on the wire at a high level:

```
+------------+------------------+-------------------+
| frame size | FContext headers | TProtocol message |
+------------+------------------+-------------------+
```

The serialization of the TProtocol message is handled entirely by the Thrift
TProtocol. For example, this could itself be framed if a TFramedTransport is
used. However, the frame size and FContext headers are serialized by FProtocol.
The header protocol reserves a single byte for versioning purposes. Currently,
only v0 is supported.

The complete binary wire layout is documented below. Network byte order is
assumed.

```
   0     1     2     3     4     5     6     7     8     9     10    11    12    13    14  ...
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+...+-----+-----+-----+-----+-----+-----+-----+...+-----+...+-----+-----+...+-----+
|     frame size n      | ver |    headers size m     |  header name size k   |  0  |  1  |...| k-1 |  header value size v  |  0  |  1  |...| v-1 |...|  0  |  1  |...| t-1 |
+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+-----+...+-----+-----+-----+-----+-----+-----|-----+...+-----+...+-----+-----+...+-----+
|<-------4 bytes------->|<----------5 bytes---------->|<-------4 bytes------->|<------k bytes------>|<-------4 bytes------->|<------v bytes------>|   |<------t bytes------>|
                                                      |<-------------------------------------------m bytes------------------------------------------->|
|<------------------------------------------------------------------------------n bytes------------------------------------------------------------------------------------>|
```

| Name                | Size    | Definition                                                   |
|---------------------|---------|--------------------------------------------------------------|
| frame size n        | 4 bytes | unsigned integer representing length of entire frame         |
| ver                 | 1 byte  | unsigned integer representing header protocol version        |
| headers size m      | 4 bytes | unsigned integer representing length of header data          |
| header name size k  | 4 bytes | unsigned integer representing the length of the header name  |
| header name         | k bytes | the header name                                              |
| header value size v | 4 bytes | unsigned integer representing the length of the header value |
| header value        | v bytes | the header value                                             |
| Thrift message      | t bytes | the TProtocol-serialized message                             |
Header key-value pairs are repeated
