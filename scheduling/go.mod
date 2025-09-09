module github.com/anasamu/microservices-library-go/scheduling

go 1.21

require (
	github.com/anasamu/microservices-library-go/scheduling/gateway v0.0.0
	github.com/anasamu/microservices-library-go/scheduling/providers/cron v0.0.0
	github.com/anasamu/microservices-library-go/scheduling/providers/redis v0.0.0
	github.com/anasamu/microservices-library-go/scheduling/types v0.0.0
)

replace github.com/anasamu/microservices-library-go/scheduling/gateway => ./gateway
replace github.com/anasamu/microservices-library-go/scheduling/providers/cron => ./providers/cron
replace github.com/anasamu/microservices-library-go/scheduling/providers/redis => ./providers/redis
replace github.com/anasamu/microservices-library-go/scheduling/types => ./types
