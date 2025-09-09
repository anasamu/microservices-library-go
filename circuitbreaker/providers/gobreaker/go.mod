module github.com/anasamu/microservices-library-go/circuitbreaker/providers/gobreaker

go 1.21

require (
	github.com/anasamu/microservices-library-go/circuitbreaker/types v0.0.0
	github.com/sony/gobreaker v0.5.0
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/sys v0.15.0
)

replace github.com/anasamu/microservices-library-go/circuitbreaker/types => ../../types
