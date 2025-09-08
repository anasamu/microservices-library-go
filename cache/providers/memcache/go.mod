module github.com/anasamu/microservices-library-go/cache/providers/memcache

go 1.21

require (
	github.com/anasamu/microservices-library-go/cache/types v0.0.0
	github.com/bradfitz/gomemcache v0.0.0-20250403215159-8d39553ac7cf
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/cache/types => ../../types
