module github.com/anasamu/microservices-library-go/ratelimit

go 1.21

require (
	github.com/anasamu/microservices-library-go/ratelimit/gateway v0.0.0
	github.com/anasamu/microservices-library-go/ratelimit/providers/inmemory v0.0.0
	github.com/anasamu/microservices-library-go/ratelimit/providers/redis v0.0.0
	github.com/anasamu/microservices-library-go/ratelimit/types v0.0.0
)

replace github.com/anasamu/microservices-library-go/ratelimit/gateway => ./gateway
replace github.com/anasamu/microservices-library-go/ratelimit/providers/inmemory => ./providers/inmemory
replace github.com/anasamu/microservices-library-go/ratelimit/providers/redis => ./providers/redis
replace github.com/anasamu/microservices-library-go/ratelimit/types => ./types
