<!-- START doctoc generated TOC please keep comment here to allow auto update -->
<!-- DON'T EDIT THIS SECTION, INSTEAD RE-RUN doctoc TO UPDATE -->
**Table of Contents**  *generated with [DocToc](https://github.com/thlorenz/doctoc)*

- [Glossary](#glossary)
  - [FAsyncCallback*](#fasynccallback)
  - [FContext](#fcontext)
  - [FProcessor](#fprocessor)
  - [FProcessorFunction*](#fprocessorfunction)
  - [FProtocol](#fprotocol)
  - [FProtocolFactory](#fprotocolfactory)
  - [FScopeProvider](#fscopeprovider)
  - [FPublisherTransport*](#fpublishertransport)
  - [FSubscriberTransport*](#fsubscribertransport)
  - [FPublisherTransportFactory*](#fpublishertransportfactory)
  - [FSubscriberTransportFactory*](#fsubscribertransportfactory)
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

<!-- END doctoc generated TOC please keep comment here to allow auto update -->

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
consumed by the user and used for distributed tracing purposes. A random
correlation ID is generated for each FContext if one is not provided.

FContext also plays a key role in Frugal's multiplexing support. A unique,
per-request operation ID is set on every FContext before a request is made.
This operation ID is sent in the request and included in the response, which is
then used to correlate a response to a request. The operation ID is an internal
implementation detail and is not exposed to the user.

An FContext should belong to a single request for the lifetime of that request.
It can be reused once the request has completed, though they should generally
not be reused.

## FProcessor

FProcessor is Frugal's equivalent of Thrift's TProcessor. It's a generic object
which operates upon an input stream and writes to an output stream.
Specifically, an FProcessor is provided to an FServer in order to wire up a
service handler to process requests.

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
FPublisherTransportFactory, an FSubscriberTransportFactory, and FProtocolFactory.

## FPublisherTransport*
FPublisherTransport is a transport that publishes messages to a pub/sub topic.

## FSubscriberTransport*
FSubscriberTransport is a transport that subscribes to a pub/sub topic. It receives
and processes messages published to the subscribed topic.

## FPublisherTransportFactory*
FPublisherTransportFactory produces FPublisherTransports, and is generally used by
an FScopeProvider.

## FSubscriberTransportFactory*
FSubscriberTransportFactory produces FSubscriberTransports, and is generally used
by an FScopeProvider.

## FServer

FServer is Frugal's equivalent of Thrift's TServer. It's used to run a Frugal
RPC service by executing an FProcessor on client connections.

Currently, Frugal includes two implementations of FServer: FSimpleServer, which
is a basic, accept-loop based server that supports traditional Thrift
TServerTransports, and FNatsServer, which is an implementation that uses
[NATS](https://nats.io/) as the underlying transport.

## FServiceProvider

FServiceProvider is the service equivalent of FScopeProvider. It produces
FTransports and FProtocols for use by RPC service clients.

## FSubscription

FSubscription is a subscription to a pub/sub topic created by a scope. The
topic subscription is actually handled by an FSubscriberTransport, which the
FSubscription wraps. Each FSubscription should have its own FSubscriberTransport.
The FSubscription is used to unsubscribe from the topic.

## FRegistry*

FRegistry is responsible for multiplexing and handling received messages.
Typically registries are only used by a client FTransport, which is making RPCs.
When a request is made, an FAsyncCallback is registered to an FContext. When a
response for the FContext is received, the FAsyncCallback is looked up,
executed, and unregistered.

## FTransport

FTransport is Frugal's equivalent of Thrift's TTransport. FTransport is
comparable to Thrift's TTransport in that it represents the transport layer
for frugal clients. However, frugal is callback based and sends only framed
data. Due to this, instead of read, write, and flush methods, FTransport has
a send method that sends framed frugal messages. To handle callback data, an
FTransport also has an FRegistry, so it provides methods for registering
and unregistering an FAsyncCallback to an FContext.

Most Frugal language libraries include an FAdapterTransport implementation, which
allows a Thrift TTransport to be used as an FTransport.

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
