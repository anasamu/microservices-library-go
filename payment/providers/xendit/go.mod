module github.com/anasamu/microservices-library-go/payment/providers/xendit

go 1.21

require (
	github.com/anasamu/microservices-library-go/payment/gateway v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace github.com/anasamu/microservices-library-go/payment/gateway => ../../gateway
