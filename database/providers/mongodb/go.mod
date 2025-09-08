module github.com/anasamu/microservices-library-go/libs/database/providers/mongodb

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/database/gateway v0.0.0
	github.com/sirupsen/logrus v1.9.3
	go.mongodb.org/mongo-driver v1.13.1
)

replace github.com/anasamu/microservices-library-go/libs/database/gateway => ../../gateway
