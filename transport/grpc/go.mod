module github.com/MouseHatGames/mice-plugins/transport/grpc

go 1.15

require (
	github.com/MouseHatGames/mice v1.2.9-0.20220821012106-dc95715ec8ac
	github.com/golang/protobuf v1.5.2
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/silenceper/pool v1.0.0
	go.opentelemetry.io/otel/exporters/jaeger v1.9.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.9.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	golang.org/x/sys v0.0.0-20220818161305-2296e01440c6 // indirect
	google.golang.org/grpc v1.46.2
	google.golang.org/protobuf v1.28.1
)
