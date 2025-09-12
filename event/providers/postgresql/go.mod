module github.com/anasamu/microservices-library-go/event/providers/postgresql

go 1.21

require (
	github.com/anasamu/microservices-library-go/event/types v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.4.0
	github.com/lib/pq v1.10.9
	github.com/sirupsen/logrus v1.9.3
)

replace github.com/anasamu/microservices-library-go/event/types => ../../types

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect
