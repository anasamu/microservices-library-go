module github.com/anasamu/microservices-library-go/libs/storage/providers/gcs

go 1.21

require (
	cloud.google.com/go/storage v1.36.0
	github.com/anasamu/microservices-library-go/libs/storage/gateway v0.0.0
	github.com/sirupsen/logrus v1.9.3
	google.golang.org/api v0.155.0
)

replace github.com/anasamu/microservices-library-go/libs/storage/gateway => ../../gateway
