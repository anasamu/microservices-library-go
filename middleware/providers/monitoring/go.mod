module github.com/anasamu/microservices-library-go/middleware/providers/monitoring

go 1.21

require (
	github.com/anasamu/microservices-library-go/middleware/types v0.0.0
	github.com/prometheus/client_golang v1.17.0
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/middleware/types => ../../types
