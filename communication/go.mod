module github.com/anasamu/microservices-library-go/communication

go 1.21

replace github.com/anasamu/microservices-library-go/communication/gateway => ./gateway

replace github.com/anasamu/microservices-library-go/communication/providers/http => ./providers/http

replace github.com/anasamu/microservices-library-go/communication/providers/websocket => ./providers/websocket

replace github.com/anasamu/microservices-library-go/communication/providers/grpc => ./providers/grpc

replace github.com/anasamu/microservices-library-go/communication/providers/graphql => ./providers/graphql

replace github.com/anasamu/microservices-library-go/communication/providers/quic => ./providers/quic

replace github.com/anasamu/microservices-library-go/communication/providers/sse => ./providers/sse
