module github.com/anasamu/microservices-library-go/libs/messaging/providers/kafka

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/messaging/gateway v0.0.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/libs/messaging/gateway => ../../gateway
