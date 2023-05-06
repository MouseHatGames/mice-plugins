module github.com/MouseHatGames/mice-plugins/discovery/kubernetes

go 1.15

require (
	github.com/MouseHatGames/mice v1.2.9-0.20230506183718-4b778dc00f3c
	github.com/opentracing/opentracing-go v1.2.0 // indirect
	github.com/patrickmn/go-cache v2.1.0+incompatible
	go.opentelemetry.io/otel/exporters/jaeger v1.9.0 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.9.0 // indirect
	go.opentelemetry.io/proto/otlp v0.19.0 // indirect
	golang.org/x/sys v0.0.0-20220818161305-2296e01440c6 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
	k8s.io/apimachinery v0.24.4
	k8s.io/client-go v0.24.4
)
