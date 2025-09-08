module github.com/anasamu/microservices-library-go/auth/examples

go 1.21

require (
	github.com/anasamu/microservices-library-go/auth/gateway v0.0.0
	github.com/anasamu/microservices-library-go/auth/providers/authentication/jwt v0.0.0
	github.com/anasamu/microservices-library-go/auth/providers/authentication/oauth v0.0.0
	github.com/anasamu/microservices-library-go/auth/providers/authentication/twofa v0.0.0
	github.com/anasamu/microservices-library-go/auth/providers/middleware v0.0.0
	github.com/anasamu/microservices-library-go/auth/types v0.0.0
	github.com/sirupsen/logrus v1.9.3
)

require (
	cloud.google.com/go/compute v1.20.1 // indirect
	cloud.google.com/go/compute/metadata v0.2.3 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.0 // indirect
	github.com/golang/protobuf v1.5.3 // indirect
	github.com/google/uuid v1.5.0 // indirect
	golang.org/x/net v0.19.0 // indirect
	golang.org/x/oauth2 v0.15.0 // indirect
	golang.org/x/sys v0.15.0 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.31.0 // indirect
)

replace github.com/anasamu/microservices-library-go/auth/gateway => ../gateway

replace github.com/anasamu/microservices-library-go/auth/providers/authentication/jwt => ../providers/authentication/jwt

replace github.com/anasamu/microservices-library-go/auth/providers/authentication/oauth => ../providers/authentication/oauth

replace github.com/anasamu/microservices-library-go/auth/providers/authentication/twofa => ../providers/authentication/twofa

replace github.com/anasamu/microservices-library-go/auth/providers/middleware => ../providers/middleware

replace github.com/anasamu/microservices-library-go/auth/types => ../types
