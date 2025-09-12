module github.com/anasamu/microservices-library-go/communication

go 1.21

require github.com/sirupsen/logrus v1.9.3

require golang.org/x/sys v0.0.0-20220715151400-c0bba94af5f8 // indirect

replace github.com/anasamu/microservices-library-go/communication/providers/http => ./providers/http

replace github.com/anasamu/microservices-library-go/communication/providers/websocket => ./providers/websocket

replace github.com/anasamu/microservices-library-go/communication/providers/grpc => ./providers/grpc

replace github.com/anasamu/microservices-library-go/communication/providers/graphql => ./providers/graphql

replace github.com/anasamu/microservices-library-go/communication/providers/quic => ./providers/quic

replace github.com/anasamu/microservices-library-go/communication/providers/sse => ./providers/sse
