module github.com/Workiva/frugal/test/integration

go 1.15

require (
	github.com/Workiva/frugal/lib/go v0.0.0
	github.com/apache/thrift v0.14.1
	github.com/go-stomp/stomp v2.1.3+incompatible
	github.com/nats-io/nats.go v1.10.1-0.20210228004050-ed743748acac
	github.com/sirupsen/logrus v1.8.1
)

replace github.com/Workiva/frugal/lib/go v0.0.0 => ../../lib/go
