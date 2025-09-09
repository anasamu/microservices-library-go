module github.com/anasamu/microservices-library-go/circuitbreaker/test/mocks

go 1.21

require (
	github.com/anasamu/microservices-library-go/circuitbreaker/types v0.0.0
	github.com/stretchr/testify v1.8.4
	golang.org/x/sys v0.15.0
)

replace github.com/anasamu/microservices-library-go/circuitbreaker/types => ../../types
