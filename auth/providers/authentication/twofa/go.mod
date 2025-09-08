module github.com/anasamu/microservices-library-go/auth/providers/authentication/twofa

go 1.21

require (
	github.com/anasamu/microservices-library-go/auth/types v0.0.0
	github.com/google/uuid v1.5.0
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.15.0 // indirect

replace github.com/anasamu/microservices-library-go/auth/types => ../../../types
