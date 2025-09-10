module github.com/anasamu/microservices-library-go/libs/database

go 1.21

replace github.com/anasamu/microservices-library-go/libs/database/gateway => ./gateway

replace github.com/anasamu/microservices-library-go/libs/database/migrations => ./migrations

replace github.com/anasamu/microservices-library-go/libs/database/providers/postgresql => ./providers/postgresql

replace github.com/anasamu/microservices-library-go/libs/database/providers/mongodb => ./providers/mongodb

replace github.com/anasamu/microservices-library-go/libs/database/providers/redis => ./providers/redis

replace github.com/anasamu/microservices-library-go/libs/database/providers/mysql => ./providers/mysql

replace github.com/anasamu/microservices-library-go/libs/database/providers/elasticsearch => ./providers/elasticsearch

replace github.com/anasamu/microservices-library-go/libs/database/providers/cassandra => ./providers/cassandra

replace github.com/anasamu/microservices-library-go/libs/database/providers/sqlite => ./providers/sqlite

replace github.com/anasamu/microservices-library-go/libs/database/providers/influxdb => ./providers/influxdb

replace github.com/anasamu/microservices-library-go/libs/database/providers/cockroachdb => ./providers/cockroachdb
