module github.com/anasamu/microservices-library-go/event/test/unit

go 1.21

require (
	github.com/anasamu/microservices-library-go/event v0.0.0-20250912091737-88f10ddf9713
	github.com/anasamu/microservices-library-go/event/types v0.0.0-20250910142242-8bec92b8b0f4
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
)

replace github.com/anasamu/microservices-library-go/event/types => ../../types

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/stretchr/objx v0.5.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
