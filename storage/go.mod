module github.com/anasamu/microservices-library-go/libs/storage

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/storage/gateway v0.0.0
	github.com/anasamu/microservices-library-go/libs/storage/providers/minio v0.0.0
	github.com/anasamu/microservices-library-go/libs/storage/providers/s3 v0.0.0
	github.com/anasamu/microservices-library-go/libs/storage/providers/gcs v0.0.0
	github.com/anasamu/microservices-library-go/libs/storage/providers/azure v0.0.0
)

replace github.com/anasamu/microservices-library-go/libs/storage/gateway => ./gateway
replace github.com/anasamu/microservices-library-go/libs/storage/providers/minio => ./providers/minio
replace github.com/anasamu/microservices-library-go/libs/storage/providers/s3 => ./providers/s3
replace github.com/anasamu/microservices-library-go/libs/storage/providers/gcs => ./providers/gcs
replace github.com/anasamu/microservices-library-go/libs/storage/providers/azure => ./providers/azure