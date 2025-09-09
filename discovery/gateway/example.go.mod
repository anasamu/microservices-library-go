module github.com/anasamu/microservices-library-go/discovery/gateway/example

go 1.21

require (
	github.com/anasamu/microservices-library-go/discovery/gateway v0.0.0
	github.com/anasamu/microservices-library-go/discovery/providers/consul v0.0.0
	github.com/anasamu/microservices-library-go/discovery/providers/etcd v0.0.0
	github.com/anasamu/microservices-library-go/discovery/providers/kubernetes v0.0.0
	github.com/anasamu/microservices-library-go/discovery/providers/static v0.0.0
	github.com/anasamu/microservices-library-go/discovery/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/discovery/gateway => ./
replace github.com/anasamu/microservices-library-go/discovery/providers/consul => ../providers/consul
replace github.com/anasamu/microservices-library-go/discovery/providers/etcd => ../providers/etcd
replace github.com/anasamu/microservices-library-go/discovery/providers/kubernetes => ../providers/kubernetes
replace github.com/anasamu/microservices-library-go/discovery/providers/static => ../providers/static
replace github.com/anasamu/microservices-library-go/discovery/types => ../types
