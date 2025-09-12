module github.com/anasamu/microservices-library-go/middleware/providers/messaging

go 1.21

require (
	github.com/anasamu/microservices-library-go/middleware/types v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/middleware/types => ../../types
