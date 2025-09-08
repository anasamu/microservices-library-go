module github.com/anasamu/microservices-library-go/libs/storage/providers/azure

go 1.21

require (
	github.com/Azure/azure-sdk-for-go/sdk/azidentity v1.5.1
	github.com/Azure/azure-sdk-for-go/sdk/storage/azblob v1.2.1
	github.com/anasamu/microservices-library-go/libs/storage/gateway v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/libs/storage/gateway => ../../gateway
