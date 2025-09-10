module github.com/anasamu/microservices-library-go/libs/database/providers/cassandra

go 1.21

replace github.com/anasamu/microservices-library-go/libs/database/gateway => ../../gateway

require (
	github.com/anasamu/microservices-library-go/database/gateway v0.0.0-20250909181146-076a5bd412bb
	github.com/gocql/gocql v1.7.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/golang/snappy v0.0.3 // indirect
	github.com/hailocab/go-hostpool v0.0.0-20160125115350-e80d13ce29ed // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	gopkg.in/inf.v0 v0.9.1 // indirect
)
