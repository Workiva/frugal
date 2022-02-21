module github.com/Workiva/frugal/examples/go

go 1.14

require (
	github.com/Workiva/frugal/lib/go v0.0.0
	github.com/apache/thrift v0.16.0
	github.com/go-stomp/stomp v2.1.4+incompatible
	github.com/nats-io/nats.go v1.13.1-0.20220121202836-972a071d373d
	github.com/rs/cors v1.8.2
	github.com/sirupsen/logrus v1.8.1
)

replace github.com/Workiva/frugal/lib/go v0.0.0 => ../../lib/go
