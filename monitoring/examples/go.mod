module github.com/anasamu/microservices-library-go/monitoring/examples

go 1.21

require (
	github.com/anasamu/microservices-library-go/monitoring/gateway v0.0.0
	github.com/anasamu/microservices-library-go/monitoring/providers/elasticsearch v0.0.0
	github.com/anasamu/microservices-library-go/monitoring/providers/jaeger v0.0.0
	github.com/anasamu/microservices-library-go/monitoring/providers/prometheus v0.0.0
	github.com/anasamu/microservices-library-go/monitoring/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/google/uuid v1.4.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace github.com/anasamu/microservices-library-go/monitoring/gateway => ../gateway

replace github.com/anasamu/microservices-library-go/monitoring/providers/elasticsearch => ../providers/elasticsearch

replace github.com/anasamu/microservices-library-go/monitoring/providers/jaeger => ../providers/jaeger

replace github.com/anasamu/microservices-library-go/monitoring/providers/prometheus => ../providers/prometheus

replace github.com/anasamu/microservices-library-go/monitoring/types => ../types
