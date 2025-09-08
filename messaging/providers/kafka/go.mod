module github.com/anasamu/microservices-library-go/messaging/providers/kafka

go 1.21

require (
	github.com/anasamu/microservices-library-go/messaging/gateway v0.0.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/google/uuid v1.6.0 // indirect
	github.com/klauspost/compress v1.15.9 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	golang.org/x/sys v0.13.0 // indirect
)

replace github.com/anasamu/microservices-library-go/messaging/gateway => ../../gateway
