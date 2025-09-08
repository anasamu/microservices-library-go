module github.com/anasamu/microservices-library-go/messaging/providers/nats

go 1.21

require (
	github.com/anasamu/microservices-library-go/messaging/gateway v0.0.0
	github.com/nats-io/nats.go v1.31.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/nats-io/nkeys v0.4.5 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	golang.org/x/crypto v0.6.0 // indirect
	golang.org/x/sys v0.5.0 // indirect
)

replace github.com/anasamu/microservices-library-go/messaging/gateway => ../../gateway
