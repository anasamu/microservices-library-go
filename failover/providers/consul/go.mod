module github.com/anasamu/microservices-library-go/failover/providers/consul

go 1.21

require (
	github.com/anasamu/microservices-library-go/failover/types v0.0.0
	github.com/hashicorp/consul/api v1.25.1
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/failover/types => ../../types
