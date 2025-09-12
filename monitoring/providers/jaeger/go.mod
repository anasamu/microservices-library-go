module github.com/anasamu/microservices-library-go/monitoring/providers/jaeger

go 1.21

require (
	github.com/anasamu/microservices-library-go/monitoring/types v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.4.0
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/monitoring/types => ../../types
