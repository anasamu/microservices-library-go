module github.com/anasamu/microservices-library-go/middleware/test/unit

go 1.21

require (
	github.com/anasamu/microservices-library-go/middleware/gateway v0.0.0
	github.com/anasamu/microservices-library-go/middleware/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/anasamu/microservices-library-go/middleware/gateway => ../../gateway
replace github.com/anasamu/microservices-library-go/middleware/types => ../../types
