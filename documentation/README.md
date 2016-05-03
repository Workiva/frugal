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

# Why does Frugal exist?

Frugal was created to address many of Thrift's shortcomings without completely
reinventing the wheel. Thrift is a solid, mature RPC framework used widely in
production systems. However, it has several key problems.

## Problems with Thrift

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

## Frugal Features

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
  used to send OAuth2 tokens and user-context information, avoiding the need to
  include it everywhere in your IDL and handler logic. Correlation IDs for
  distributed tracing purposes are also built into the request context.

- Middleware: client- and server- side middleware is supported for RPC and
  pub/sub APIs. This allows you to implement interceptor logic around handler
  functions, e.g. for authentication, logging, or retry policies.
