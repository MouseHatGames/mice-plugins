module github.com/MouseHatGames/mice-plugins/transport/grpc

go 1.15

replace github.com/MouseHatGames/mice => ../../../mice

require (
	github.com/MouseHatGames/mice v0.0.0-20201125151056-e701c86b93b1
	github.com/golang/protobuf v1.4.3
	google.golang.org/grpc v1.33.2
	google.golang.org/protobuf v1.25.0
)
