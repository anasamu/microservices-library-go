module github.com/anasamu/microservices-library-go/storage

go 1.21

require (
	github.com/google/uuid v1.6.0
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.15.0 // indirect

replace github.com/anasamu/microservices-library-go/storage/providers/minio => ./providers/minio

replace github.com/anasamu/microservices-library-go/storage/providers/s3 => ./providers/s3

replace github.com/anasamu/microservices-library-go/storage/providers/gcs => ./providers/gcs

replace github.com/anasamu/microservices-library-go/storage/providers/azure => ./providers/azure

replace github.com/anasamu/microservices-library-go/storage/types => ./types
