module github.com/anasamu/microservices-library-go/libs/communication/providers/websocket

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/communication/gateway v0.0.0
	github.com/gorilla/websocket v1.5.1
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/libs/communication/gateway => ../../gateway

require (
	golang.org/x/net v0.17.0 // indirect
)
