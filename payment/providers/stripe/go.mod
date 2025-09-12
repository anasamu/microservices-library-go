module github.com/anasamu/microservices-library-go/payment/providers/stripe

go 1.21

require (
	github.com/anasamu/microservices-library-go/payment v0.0.0
	github.com/sirupsen/logrus v1.9.3
	github.com/stripe/stripe-go/v78 v78.8.0
)

replace github.com/anasamu/microservices-library-go/payment => ../../

require (
	github.com/google/uuid v1.6.0 // indirect
	golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
)
