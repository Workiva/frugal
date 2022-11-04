module github.com/Workiva/frugal/test/integration

go 1.18

require (
	github.com/Workiva/frugal/lib/go v0.0.0
	github.com/apache/thrift v0.14.2
	github.com/go-stomp/stomp v2.1.4+incompatible
	github.com/nats-io/nats.go v1.19.0
	github.com/sirupsen/logrus v1.9.0
)

require (
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20220829220503-c86fa9a7ed90 // indirect
	golang.org/x/sys v0.0.0-20220906135438-9e1f76180b77 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)

replace github.com/Workiva/frugal/lib/go v0.0.0 => ../../lib/go
