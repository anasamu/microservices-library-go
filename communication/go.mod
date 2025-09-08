module github.com/anasamu/microservices-library-go/libs/communication

go 1.21

require (
	github.com/anasamu/microservices-library-go/libs/communication/gateway v0.0.0
	github.com/anasamu/microservices-library-go/libs/communication/providers/http v0.0.0
	github.com/anasamu/microservices-library-go/libs/communication/providers/websocket v0.0.0
	github.com/anasamu/microservices-library-go/libs/communication/providers/grpc v0.0.0
)

replace github.com/anasamu/microservices-library-go/libs/communication/gateway => ./gateway
replace github.com/anasamu/microservices-library-go/libs/communication/providers/http => ./providers/http
replace github.com/anasamu/microservices-library-go/libs/communication/providers/websocket => ./providers/websocket
replace github.com/anasamu/microservices-library-go/libs/communication/providers/grpc => ./providers/grpc
