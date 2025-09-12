module github.com/anasamu/microservices-library-go/database

go 1.21

require github.com/sirupsen/logrus v1.9.3

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/database/migrations => ./migrations

replace github.com/anasamu/microservices-library-go/database/providers/postgresql => ./providers/postgresql

replace github.com/anasamu/microservices-library-go/database/providers/mongodb => ./providers/mongodb

replace github.com/anasamu/microservices-library-go/database/providers/redis => ./providers/redis

replace github.com/anasamu/microservices-library-go/database/providers/mysql => ./providers/mysql

replace github.com/anasamu/microservices-library-go/database/providers/elasticsearch => ./providers/elasticsearch

replace github.com/anasamu/microservices-library-go/database/providers/cassandra => ./providers/cassandra

replace github.com/anasamu/microservices-library-go/database/providers/sqlite => ./providers/sqlite

replace github.com/anasamu/microservices-library-go/database/providers/influxdb => ./providers/influxdb

replace github.com/anasamu/microservices-library-go/database/providers/cockroachdb => ./providers/cockroachdb
