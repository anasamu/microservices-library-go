module github.com/anasamu/microservices-library-go/chaos/test/unit

go 1.21

require (
	github.com/anasamu/microservices-library-go/chaos v0.0.0
	github.com/anasamu/microservices-library-go/chaos/test/mocks v0.0.0
)

replace (
	github.com/anasamu/microservices-library-go/chaos => ../../gateway
	github.com/anasamu/microservices-library-go/chaos/test/mocks => ../mocks
)
