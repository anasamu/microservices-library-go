module github.com/anasamu/microservices-library-go/messaging

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/messaging/providers/kafka => ./providers/kafka

replace github.com/anasamu/microservices-library-go/messaging/providers/rabbitmq => ./providers/rabbitmq

replace github.com/anasamu/microservices-library-go/messaging/providers/sqs => ./providers/sqs

replace github.com/anasamu/microservices-library-go/messaging/providers/nats => ./providers/nats
