module github.com/anasamu/microservices-library-go/ratelimit/test/unit

go 1.21

require (
	github.com/anasamu/microservices-library-go/ratelimit/gateway v0.0.0
	github.com/anasamu/microservices-library-go/ratelimit/test/mocks v0.0.0
	github.com/anasamu/microservices-library-go/ratelimit/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
)

replace github.com/anasamu/microservices-library-go/ratelimit/gateway => ../../gateway
replace github.com/anasamu/microservices-library-go/ratelimit/test/mocks => ../mocks
replace github.com/anasamu/microservices-library-go/ratelimit/types => ../../types
