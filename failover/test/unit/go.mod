module github.com/anasamu/microservices-library-go/failover/test/unit

go 1.21

require (
	github.com/anasamu/microservices-library-go/failover v0.0.0-20250912091737-88f10ddf9713
	github.com/anasamu/microservices-library-go/failover/test/mocks v0.0.0-00010101000000-000000000000
	github.com/anasamu/microservices-library-go/failover/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/anasamu/microservices-library-go/failover/test/mocks => ../mocks

replace github.com/anasamu/microservices-library-go/failover/types => ../../types
