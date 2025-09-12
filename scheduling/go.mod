module github.com/anasamu/microservices-library-go/scheduling

go 1.21

require github.com/sirupsen/logrus v1.9.3

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

require github.com/anasamu/microservices-library-go/scheduling/types v0.0.0-20250912212654-08af9e89ff53

replace github.com/anasamu/microservices-library-go/scheduling/providers/cron => ./providers/cron

replace github.com/anasamu/microservices-library-go/scheduling/providers/redis => ./providers/redis

replace github.com/anasamu/microservices-library-go/scheduling/types => ./types
