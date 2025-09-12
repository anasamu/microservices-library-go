module github.com/anasamu/microservices-library-go/config

go 1.21

replace github.com/anasamu/microservices-library-go/config/types => ./types

replace github.com/anasamu/microservices-library-go/config/providers/file => ./providers/file

replace github.com/anasamu/microservices-library-go/config/providers/env => ./providers/env

replace github.com/anasamu/microservices-library-go/config/providers/vault => ./providers/vault

replace github.com/anasamu/microservices-library-go/config/providers/consul => ./providers/consul

require github.com/anasamu/microservices-library-go/config/types v0.0.0-00010101000000-000000000000
