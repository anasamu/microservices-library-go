module github.com/anasamu/microservices-library-go/logging/providers/elasticsearch

go 1.21

require (
	github.com/anasamu/microservices-library-go/logging/types v0.0.0
	github.com/elastic/go-elasticsearch/v8 v8.11.1
)

require github.com/elastic/elastic-transport-go/v8 v8.3.0 // indirect

replace github.com/anasamu/microservices-library-go/logging/types => ../../types
