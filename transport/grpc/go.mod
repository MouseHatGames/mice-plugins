module github.com/MouseHatGames/mice-plugins/transport/grpc

go 1.15

replace github.com/MouseHatGames/mice => ../../../mice

require (
	github.com/MouseHatGames/mice v0.0.0-00010101000000-000000000000
	github.com/golang/protobuf v1.4.3
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.25.0
)
