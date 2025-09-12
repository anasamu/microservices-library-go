module github.com/anasamu/microservices-library-go/failover

go 1.21

require (
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

require (
	github.com/anasamu/microservices-library-go/failover/gateway v0.0.0
	github.com/anasamu/microservices-library-go/failover/providers/consul v0.0.0
	github.com/anasamu/microservices-library-go/failover/providers/kubernetes v0.0.0
	github.com/anasamu/microservices-library-go/failover/types v0.0.0
)

replace github.com/anasamu/microservices-library-go/failover/providers/consul => ./providers/consul
replace github.com/anasamu/microservices-library-go/failover/providers/kubernetes => ./providers/kubernetes
replace github.com/anasamu/microservices-library-go/failover/types => ./types
