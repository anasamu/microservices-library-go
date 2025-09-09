module github.com/anasamu/microservices-library-go/chaos/test/integration

go 1.21

require (
	github.com/anasamu/microservices-library-go/chaos/gateway v0.0.0
	github.com/anasamu/microservices-library-go/chaos/providers/http v0.0.0
	github.com/anasamu/microservices-library-go/chaos/providers/messaging v0.0.0
)

replace (
	github.com/anasamu/microservices-library-go/chaos/gateway => ../../gateway
	github.com/anasamu/microservices-library-go/chaos/providers/http => ../../providers/http
	github.com/anasamu/microservices-library-go/chaos/providers/messaging => ../../providers/messaging
)
