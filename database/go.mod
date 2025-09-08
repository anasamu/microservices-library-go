module github.com/anasamu/microservices-library-go/libs/database

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/database/gateway v0.0.0
	github.com/anasamu/microservices-library-go/libs/database/providers/postgresql v0.0.0
	github.com/anasamu/microservices-library-go/libs/database/providers/mongodb v0.0.0
	github.com/anasamu/microservices-library-go/libs/database/providers/redis v0.0.0
	github.com/anasamu/microservices-library-go/libs/database/providers/mysql v0.0.0
)

replace github.com/anasamu/microservices-library-go/libs/database/gateway => ./gateway
replace github.com/anasamu/microservices-library-go/libs/database/providers/postgresql => ./providers/postgresql
replace github.com/anasamu/microservices-library-go/libs/database/providers/mongodb => ./providers/mongodb
replace github.com/anasamu/microservices-library-go/libs/database/providers/redis => ./providers/redis
replace github.com/anasamu/microservices-library-go/libs/database/providers/mysql => ./providers/mysql