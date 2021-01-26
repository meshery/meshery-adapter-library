module github.com/layer5io/meshery-adapter-library

go 1.13

replace github.com/kudobuilder/kuttl => github.com/layer5io/kuttl v0.4.1-0.20200806180306-b7e46afd657f

replace github.com/layer5io/meshkit v0.1.30 => ../meshkit

require (
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/layer5io/learn-layer5/smi-conformance v0.0.0-20210125055931-43e1dc9563c4
	github.com/layer5io/meshkit v0.1.30
	github.com/layer5io/service-mesh-performance v0.3.2
	github.com/spf13/viper v1.7.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc v0.11.0
	go.opentelemetry.io/otel v0.11.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.11.0
	go.opentelemetry.io/otel/sdk v0.11.0
	golang.org/x/net v0.0.0-20200822124328-c89045814202
	google.golang.org/grpc v1.31.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/client-go v0.18.12
)
