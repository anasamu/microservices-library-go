module github.com/anasamu/microservices-library-go/libs/database

go 1.21

replace github.com/anasamu/microservices-library-go/libs/database/gateway => ./gateway

replace github.com/anasamu/microservices-library-go/libs/database/providers/postgresql => ./providers/postgresql

replace github.com/anasamu/microservices-library-go/libs/database/providers/mongodb => ./providers/mongodb

replace github.com/anasamu/microservices-library-go/libs/database/providers/redis => ./providers/redis

replace github.com/anasamu/microservices-library-go/libs/database/providers/mysql => ./providers/mysql

replace github.com/anasamu/microservices-library-go/libs/database/providers/elasticsearch => ./providers/elasticsearch
