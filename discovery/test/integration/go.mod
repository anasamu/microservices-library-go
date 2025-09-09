module github.com/anasamu/microservices-library-go/discovery/test/integration

go 1.21

require (
	github.com/anasamu/microservices-library-go/discovery/gateway v0.0.0
	github.com/anasamu/microservices-library-go/discovery/providers/static v0.0.0
	github.com/anasamu/microservices-library-go/discovery/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
)

replace github.com/anasamu/microservices-library-go/discovery/gateway => ../../gateway
replace github.com/anasamu/microservices-library-go/discovery/providers/static => ../../providers/static
replace github.com/anasamu/microservices-library-go/discovery/types => ../../types
