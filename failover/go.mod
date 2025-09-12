module github.com/anasamu/microservices-library-go/failover

go 1.21

require github.com/sirupsen/logrus v1.9.3

require (
	golang.org/x/sys v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

require github.com/anasamu/microservices-library-go/failover/types v0.0.0

replace github.com/anasamu/microservices-library-go/failover/providers/consul => ./providers/consul

replace github.com/anasamu/microservices-library-go/failover/providers/kubernetes => ./providers/kubernetes

replace github.com/anasamu/microservices-library-go/failover/types => ./types
