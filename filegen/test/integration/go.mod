module github.com/anasamu/microservices-library-go/filegen/test/integration

go 1.21

require (
	github.com/anasamu/microservices-library-go/filegen/providers/docx v0.0.0
	github.com/anasamu/microservices-library-go/filegen/providers/excel v0.0.0
	github.com/anasamu/microservices-library-go/filegen/providers/csv v0.0.0
	github.com/anasamu/microservices-library-go/filegen/providers/pdf v0.0.0
)

replace (
	github.com/anasamu/microservices-library-go/filegen/providers/docx => ../../providers/docx
	github.com/anasamu/microservices-library-go/filegen/providers/excel => ../../providers/excel
	github.com/anasamu/microservices-library-go/filegen/providers/csv => ../../providers/csv
	github.com/anasamu/microservices-library-go/filegen/providers/pdf => ../../providers/pdf
)
