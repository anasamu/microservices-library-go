module github.com/siakad/microservices/libs/cache

go 1.21

require (
	github.com/sirupsen/logrus v1.9.3
	github.com/google/uuid v1.5.0
	github.com/redis/go-redis/v9 v9.3.0
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874
	github.com/patrickmn/go-cache v2.1.0+incompatible
	github.com/go-redis/cache/v9 v9.1.0
)

require (
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/vmihailenco/go-tinylfu v0.2.2 // indirect
	github.com/vmihailenco/msgpack/v5 v5.3.5 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
)
