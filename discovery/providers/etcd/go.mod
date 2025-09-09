module github.com/anasamu/microservices-library-go/discovery/providers/etcd

go 1.21

require (
	github.com/anasamu/microservices-library-go/discovery/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
	go.etcd.io/etcd/clientv3 v3.5.10
)

replace github.com/anasamu/microservices-library-go/discovery/types => ../../types
