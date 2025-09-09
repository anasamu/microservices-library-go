module github.com/anasamu/microservices-library-go/discovery

go 1.21

require (
	github.com/anasamu/microservices-library-go/discovery/gateway v0.0.0
	github.com/anasamu/microservices-library-go/discovery/types v0.0.0
)

replace github.com/anasamu/microservices-library-go/discovery/gateway => ./gateway
replace github.com/anasamu/microservices-library-go/discovery/types => ./types
