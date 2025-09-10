module github.com/anasamu/microservices-library-go/database/providers/mariadb

go 1.21

require (
	github.com/anasamu/microservices-library-go/database/gateway v0.0.0
	github.com/jmoiron/sqlx v1.3.5
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/go-sql-driver/mysql v1.7.1 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace github.com/anasamu/microservices-library-go/database/gateway => ../../gateway
