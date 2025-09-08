module github.com/anasamu/microservices-library-go/auth/providers/authorization/rbac

go 1.21

require (
	github.com/google/uuid v1.5.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	github.com/anasamu/microservices-library-go/auth/types v0.0.0-20250908142349-c02445e2700e
	golang.org/x/sys v0.15.0 // indirect
)

replace github.com/anasamu/microservices-library-go/auth/types => ../../../types
