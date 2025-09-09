module github.com/anasamu/microservices-library-go/filegen/gateway

go 1.21

require (
	github.com/anasamu/microservices-library-go/filegen/types v0.0.0
	github.com/anasamu/microservices-library-go/filegen/providers/csv v0.0.0
	github.com/anasamu/microservices-library-go/filegen/providers/custom v0.0.0
	github.com/anasamu/microservices-library-go/filegen/providers/docx v0.0.0
	github.com/anasamu/microservices-library-go/filegen/providers/excel v0.0.0
	github.com/anasamu/microservices-library-go/filegen/providers/pdf v0.0.0
)

require (
	github.com/mohae/deepcopy v0.0.0-20170929034955-c48cc78d4826 // indirect
	github.com/richardlehane/mscfb v1.0.4 // indirect
	github.com/richardlehane/msoleps v1.0.3 // indirect
	github.com/unidoc/unioffice v1.26.0 // indirect
	github.com/xuri/efp v0.0.0-20231025114914-d1ff6096ae53 // indirect
	github.com/xuri/excelize/v2 v2.8.0 // indirect
	github.com/xuri/nfp v0.0.0-20230919160717-d98342af3f05 // indirect
	golang.org/x/crypto v0.18.0 // indirect
	golang.org/x/net v0.20.0 // indirect
	golang.org/x/text v0.14.0 // indirect
)

replace (
	github.com/anasamu/microservices-library-go/filegen/types => ../types
	github.com/anasamu/microservices-library-go/filegen/providers/csv => ../providers/csv
	github.com/anasamu/microservices-library-go/filegen/providers/custom => ../providers/custom
	github.com/anasamu/microservices-library-go/filegen/providers/docx => ../providers/docx
	github.com/anasamu/microservices-library-go/filegen/providers/excel => ../providers/excel
	github.com/anasamu/microservices-library-go/filegen/providers/pdf => ../providers/pdf
)
