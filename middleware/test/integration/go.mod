module github.com/anasamu/microservices-library-go/middleware/test/integration

go 1.21

require (
	github.com/anasamu/microservices-library-go/middleware/gateway v0.0.0
	github.com/anasamu/microservices-library-go/middleware/providers/auth v0.0.0
	github.com/anasamu/microservices-library-go/middleware/providers/logging v0.0.0
	github.com/anasamu/microservices-library-go/middleware/providers/ratelimit v0.0.0
	github.com/anasamu/microservices-library-go/middleware/providers/storage v0.0.0
	github.com/anasamu/microservices-library-go/middleware/providers/communication v0.0.0
	github.com/anasamu/microservices-library-go/middleware/providers/messaging v0.0.0
	github.com/anasamu/microservices-library-go/middleware/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stretchr/testify v1.8.4
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang-jwt/jwt/v5 v5.2.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/anasamu/microservices-library-go/middleware/gateway => ../../gateway
replace github.com/anasamu/microservices-library-go/middleware/providers/auth => ../../providers/auth
replace github.com/anasamu/microservices-library-go/middleware/providers/logging => ../../providers/logging
replace github.com/anasamu/microservices-library-go/middleware/providers/ratelimit => ../../providers/ratelimit
replace github.com/anasamu/microservices-library-go/middleware/providers/storage => ../../providers/storage
replace github.com/anasamu/microservices-library-go/middleware/providers/communication => ../../providers/communication
replace github.com/anasamu/microservices-library-go/middleware/providers/messaging => ../../providers/messaging
replace github.com/anasamu/microservices-library-go/middleware/types => ../../types
