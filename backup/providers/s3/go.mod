module github.com/anasamu/microservices-library-go/backup/providers/s3

go 1.21

require (
	github.com/anasamu/microservices-library-go/backup v0.0.0-00010101000000-000000000000
	github.com/aws/aws-sdk-go v1.50.0
)

require github.com/jmespath/go-jmespath v0.4.0 // indirect

replace github.com/anasamu/microservices-library-go/backup => ../../
