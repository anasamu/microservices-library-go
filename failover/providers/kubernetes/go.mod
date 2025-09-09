module github.com/anasamu/microservices-library-go/failover/providers/kubernetes

go 1.21

require (
	github.com/anasamu/microservices-library-go/failover/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
	k8s.io/api v0.28.4
	k8s.io/apimachinery v0.28.4
	k8s.io/client-go v0.28.4
)

replace github.com/anasamu/microservices-library-go/failover/types => ../../types
