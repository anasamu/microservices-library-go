module github.com/anasamu/microservices-library-go/communication/providers/graphql

go 1.21

require (
	github.com/anasamu/microservices-library-go/communication/gateway v0.0.0
	github.com/graphql-go/graphql v0.8.1
	github.com/graphql-go/handler v0.2.3
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/communication/gateway => ../../gateway
