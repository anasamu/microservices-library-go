module github.com/anasamu/microservices-library-go/event/examples

go 1.21

require (
	github.com/anasamu/microservices-library-go/event v0.0.0
	github.com/anasamu/microservices-library-go/event/providers/kafka v0.0.0
	github.com/anasamu/microservices-library-go/event/providers/nats v0.0.0
	github.com/anasamu/microservices-library-go/event/providers/postgresql v0.0.0
	github.com/anasamu/microservices-library-go/event/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/google/uuid v1.4.0 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/nats-io/nats.go v1.31.0 // indirect
	github.com/nats-io/nkeys v0.4.5 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/segmentio/kafka-go v0.4.47 // indirect
	golang.org/x/crypto v0.14.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
)

replace github.com/anasamu/microservices-library-go/event/providers/kafka => ../providers/kafka

replace github.com/anasamu/microservices-library-go/event/providers/nats => ../providers/nats

replace github.com/anasamu/microservices-library-go/event/providers/postgresql => ../providers/postgresql

replace github.com/anasamu/microservices-library-go/event/types => ../types

replace github.com/anasamu/microservices-library-go/event => ..
