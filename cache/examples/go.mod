module github.com/anasamu/microservices-library-go/cache/examples

go 1.21

require (
	github.com/anasamu/microservices-library-go/cache v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/cache/providers/memory v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/cache/providers/redis v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/anasamu/microservices-library-go/cache/types v0.0.0 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/go-redis/redis/v8 v8.11.5 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace github.com/anasamu/microservices-library-go/cache/providers/memory => ../providers/memory

replace github.com/anasamu/microservices-library-go/cache/providers/redis => ../providers/redis

replace github.com/anasamu/microservices-library-go/cache/types => ../types

replace github.com/anasamu/microservices-library-go/cache => ../
