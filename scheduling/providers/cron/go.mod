module github.com/anasamu/microservices-library-go/scheduling/providers/cron

go 1.21

require (
	github.com/anasamu/microservices-library-go/scheduling/types v0.0.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/sirupsen/logrus v1.9.3
)

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/scheduling/types => ../../types
