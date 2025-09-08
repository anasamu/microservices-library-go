module github.com/anasamu/microservices-library-go/libs/database/providers/redis

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/database/gateway v0.0.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/libs/database/gateway => ../../gateway
