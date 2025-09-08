module github.com/anasamu/microservices-library-go/messaging

go 1.21

replace github.com/anasamu/microservices-library-go/messaging/gateway => ./gateway

replace github.com/anasamu/microservices-library-go/messaging/providers/kafka => ./providers/kafka

replace github.com/anasamu/microservices-library-go/messaging/providers/rabbitmq => ./providers/rabbitmq

replace github.com/anasamu/microservices-library-go/messaging/providers/sqs => ./providers/sqs

replace github.com/anasamu/microservices-library-go/messaging/providers/nats => ./providers/nats
