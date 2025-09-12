module github.com/anasamu/microservices-library-go/backup/test/unit

go 1.21

require (
	github.com/anasamu/microservices-library-go/backup v0.0.0
	github.com/anasamu/microservices-library-go/backup/gateway v0.0.0-20250910142242-8bec92b8b0f4
	github.com/anasamu/microservices-library-go/backup/test/mocks v0.0.0-20250910142242-8bec92b8b0f4
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/anasamu/microservices-library-go/backup/providers/local v0.0.0-20250910142242-8bec92b8b0f4 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/anasamu/microservices-library-go/backup => ../../
