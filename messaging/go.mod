module github.com/anasamu/microservices-library-go/libs/messaging

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/messaging/gateway v0.0.0
	github.com/anasamu/microservices-library-go/libs/messaging/providers/kafka v0.0.0
	github.com/anasamu/microservices-library-go/libs/messaging/providers/rabbitmq v0.0.0
	github.com/anasamu/microservices-library-go/libs/messaging/providers/sqs v0.0.0
)

replace github.com/anasamu/microservices-library-go/libs/messaging/gateway => ./gateway
replace github.com/anasamu/microservices-library-go/libs/messaging/providers/kafka => ./providers/kafka
replace github.com/anasamu/microservices-library-go/libs/messaging/providers/rabbitmq => ./providers/rabbitmq
replace github.com/anasamu/microservices-library-go/libs/messaging/providers/sqs => ./providers/sqs