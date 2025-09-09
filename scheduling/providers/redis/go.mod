module github.com/anasamu/microservices-library-go/scheduling/providers/redis

go 1.21

require (
	github.com/anasamu/microservices-library-go/scheduling/types v0.0.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/scheduling/types => ../../types
