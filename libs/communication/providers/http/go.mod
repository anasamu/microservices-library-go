module github.com/anasamu/microservices-library-go/libs/communication/providers/http

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/communication/gateway v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/libs/communication/gateway => ../../gateway
