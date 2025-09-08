module github.com/anasamu/microservices-library-go/libs/storage/providers/minio

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/storage/gateway v0.0.0
	github.com/minio/minio-go/v7 v7.0.66
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/libs/storage/gateway => ../../gateway
