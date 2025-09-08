module github.com/anasamu/microservices-library-go/auth/providers/authorization/acl

go 1.24.0

toolchain go1.24.2

require (
	github.com/anasamu/microservices-library-go/auth/types v0.0.0
	github.com/google/uuid v1.6.0
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.36.0 // indirect

replace github.com/anasamu/microservices-library-go/auth/types => ../../../types
