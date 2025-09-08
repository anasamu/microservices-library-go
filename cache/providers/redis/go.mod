module github.com/anasamu/microservices-library-go/cache/providers/redis

go 1.21

require (
	github.com/anasamu/microservices-library-go/cache/types v0.0.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace github.com/anasamu/microservices-library-go/cache/types => ../../types
