module github.com/anasamu/microservices-library-go/config

go 1.21

replace github.com/anasamu/microservices-library-go/config/types => ./types

replace github.com/anasamu/microservices-library-go/config/providers/file => ./providers/file

replace github.com/anasamu/microservices-library-go/config/providers/env => ./providers/env

replace github.com/anasamu/microservices-library-go/config/providers/vault => ./providers/vault

replace github.com/anasamu/microservices-library-go/config/providers/consul => ./providers/consul

require github.com/anasamu/microservices-library-go/config/types v0.0.0-20250912212654-08af9e89ff53-20250912214337-08af9e89ff53
