module github.com/anasamu/microservices-library-go/libs/database/providers/mysql

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/database/gateway v0.0.0
	github.com/go-sql-driver/mysql v1.7.1
	github.com/jmoiron/sqlx v1.3.5
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/libs/database/gateway => ../../gateway
