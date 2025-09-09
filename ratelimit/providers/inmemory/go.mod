module github.com/anasamu/microservices-library-go/ratelimit/providers/inmemory

go 1.21

require (
	github.com/anasamu/microservices-library-go/ratelimit/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/ratelimit/types => ../../types
