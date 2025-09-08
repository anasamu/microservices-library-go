module github.com/anasamu/microservices-library-go/libs/messaging/providers/rabbitmq

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/messaging/gateway v0.0.0
	github.com/rabbitmq/amqp091-go v1.9.0
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/libs/messaging/gateway => ../../gateway
