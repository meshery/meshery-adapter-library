module github.com/layer5io/meshery-adapter-library

go 1.13

replace (
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/kudobuilder/kuttl => github.com/layer5io/kuttl v0.4.1-0.20200806180306-b7e46afd657f
	golang.org/x/sys => golang.org/x/sys v0.0.0-20200826173525-f9321e4c35a6
)

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/golang/protobuf v1.4.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/layer5io/learn-layer5/smi-conformance v0.0.0-20210317075357-06b4f88b3e34
	github.com/layer5io/meshkit v0.2.7
	github.com/layer5io/service-mesh-performance v0.3.2
	github.com/spf13/viper v1.7.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc v0.11.0
	go.opentelemetry.io/otel v0.11.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.11.0
	go.opentelemetry.io/otel/sdk v0.11.0
	google.golang.org/grpc v1.34.0
	google.golang.org/protobuf v1.25.0
	gopkg.in/yaml.v2 v2.3.0
	k8s.io/client-go v0.18.12
)
