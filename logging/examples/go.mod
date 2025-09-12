module github.com/anasamu/microservices-library-go/logging/examples

go 1.21

require (
	github.com/anasamu/microservices-library-go/logging v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/logging/providers/console v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/logging/providers/elasticsearch v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/logging/providers/file v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/logging/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/elastic/elastic-transport-go/v8 v8.3.0 // indirect
	github.com/elastic/go-elasticsearch/v8 v8.11.1 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)

replace github.com/anasamu/microservices-library-go/logging/providers/console => ../providers/console

replace github.com/anasamu/microservices-library-go/logging/providers/elasticsearch => ../providers/elasticsearch

replace github.com/anasamu/microservices-library-go/logging/providers/file => ../providers/file

replace github.com/anasamu/microservices-library-go/logging/types => ../types

replace github.com/anasamu/microservices-library-go/logging => ..
