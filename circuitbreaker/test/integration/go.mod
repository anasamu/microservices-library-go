module github.com/anasamu/microservices-library-go/circuitbreaker/test/integration

go 1.21

require (
	github.com/anasamu/microservices-library-go/circuitbreaker/gateway v0.0.0
	github.com/anasamu/microservices-library-go/circuitbreaker/providers/custom v0.0.0
	github.com/anasamu/microservices-library-go/circuitbreaker/providers/gobreaker v0.0.0
	github.com/anasamu/microservices-library-go/circuitbreaker/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
	golang.org/x/sys v0.15.0
)

replace github.com/anasamu/microservices-library-go/circuitbreaker/gateway => ../../gateway
replace github.com/anasamu/microservices-library-go/circuitbreaker/providers/custom => ../../providers/custom
replace github.com/anasamu/microservices-library-go/circuitbreaker/providers/gobreaker => ../../providers/gobreaker
replace github.com/anasamu/microservices-library-go/circuitbreaker/types => ../../types
