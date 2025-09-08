module github.com/anasamu/microservices-library-go/logging

go 1.21

replace github.com/anasamu/microservices-library-go/logging/types => ./types

replace github.com/anasamu/microservices-library-go/logging/gateway => ./gateway

replace github.com/anasamu/microservices-library-go/logging/providers/console => ./providers/console

replace github.com/anasamu/microservices-library-go/logging/providers/file => ./providers/file

replace github.com/anasamu/microservices-library-go/logging/providers/elasticsearch => ./providers/elasticsearch

replace github.com/anasamu/microservices-library-go/logging/examples => ./examples
