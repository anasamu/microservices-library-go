module github.com/anasamu/microservices-library-go/auth/providers/authentication/jwt

go 1.21

require (
	github.com/anasamu/microservices-library-go/auth/types v0.0.0-00010101000000-000000000000
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.5.0
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/auth/types => ../../../types
