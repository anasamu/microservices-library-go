module github.com/anasamu/microservices-library-go/libs/messaging/providers/sqs

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/messaging/gateway v0.0.0
	github.com/aws/aws-sdk-go-v2 v1.24.0
	github.com/aws/aws-sdk-go-v2/config v1.26.1
	github.com/aws/aws-sdk-go-v2/credentials v1.16.12
	github.com/aws/aws-sdk-go-v2/service/sqs v1.29.6
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/libs/messaging/gateway => ../../gateway
