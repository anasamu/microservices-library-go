module github.com/anasamu/microservices-library-go/auth/providers/authentication/oauth

go 1.21

require (
	github.com/anasamu/microservices-library-go/auth/types v0.0.0
	github.com/google/uuid v1.5.0
	github.com/sirupsen/logrus v1.9.3
	golang.org/x/oauth2 v0.15.0
)

require (
	cloud.google.com/go/compute v1.20.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/anasamu/microservices-library-go/auth/types => ../../../types
