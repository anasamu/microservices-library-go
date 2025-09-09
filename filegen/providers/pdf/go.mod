module github.com/anasamu/microservices-library-go/filegen/providers/pdf

go 1.21

require (
	github.com/anasamu/microservices-library-go/filegen/types v0.0.0
	github.com/unidoc/unioffice v1.26.0
)

replace github.com/anasamu/microservices-library-go/filegen/types => ../../types

require github.com/richardlehane/msoleps v1.0.1 // indirect
