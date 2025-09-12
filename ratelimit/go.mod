module github.com/anasamu/microservices-library-go/ratelimit

go 1.21

require (
	github.com/anasamu/microservices-library-go/ratelimit/types v0.0.0-20250912212654-08af9e89ff53-20250912214337-08af9e89ff53
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/ratelimit/providers/inmemory => ./providers/inmemory

replace github.com/anasamu/microservices-library-go/ratelimit/providers/redis => ./providers/redis

replace github.com/anasamu/microservices-library-go/ratelimit/types => ./types
