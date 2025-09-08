module github.com/anasamu/microservices-library-go/database/examples

go 1.21

require (
	github.com/anasamu/microservices-library-go/database/gateway v0.0.0
	github.com/anasamu/microservices-library-go/database/providers/elasticsearch v0.0.0
	github.com/anasamu/microservices-library-go/database/providers/mongodb v0.0.0
	github.com/anasamu/microservices-library-go/database/providers/mysql v0.0.0
	github.com/anasamu/microservices-library-go/database/providers/postgresql v0.0.0
	github.com/anasamu/microservices-library-go/database/providers/redis v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/elastic/elastic-transport-go/v8 v8.3.0 // indirect
	github.com/elastic/go-elasticsearch/v8 v8.11.1 // indirect
	github.com/golang/snappy v0.0.1 // indirect
	github.com/jmoiron/sqlx v1.3.5 // indirect
	github.com/klauspost/compress v1.13.6 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/montanaflynn/stats v0.0.0-20171201202039-1bf9dbcd8cbe // indirect
	github.com/redis/go-redis/v9 v9.3.0 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20181117223130-1be2e3e5546d // indirect
	go.mongodb.org/mongo-driver v1.13.1 // indirect
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d // indirect
	golang.org/x/sync v0.0.0-20220722155255-886fb9371eb4 // indirect
	golang.org/x/sys v0.0.0-20220722155257-8c9f86f7a55f // indirect
	golang.org/x/text v0.7.0 // indirect
)

replace github.com/anasamu/microservices-library-go/database/gateway => ../gateway

replace github.com/anasamu/microservices-library-go/database/providers/elasticsearch => ../providers/elasticsearch

replace github.com/anasamu/microservices-library-go/database/providers/mongodb => ../providers/mongodb

replace github.com/anasamu/microservices-library-go/database/providers/mysql => ../providers/mysql

replace github.com/anasamu/microservices-library-go/database/providers/postgresql => ../providers/postgresql

replace github.com/anasamu/microservices-library-go/database/providers/redis => ../providers/redis
