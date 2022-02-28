module github.com/layer5io/meshery-adapter-library

go 1.13

replace (
	github.com/docker/docker => github.com/moby/moby v17.12.0-ce-rc1.0.20200618181300-9dc6525e6118+incompatible
	github.com/kudobuilder/kuttl => github.com/layer5io/kuttl v0.4.1-0.20200806180306-b7e46afd657f
	github.com/spf13/afero => github.com/spf13/afero v1.5.1 // Until viper bug is resolved #1161
)

require (
	github.com/cenkalti/backoff/v4 v4.1.0
	github.com/golang/protobuf v1.5.2
	github.com/grpc-ecosystem/go-grpc-middleware v1.2.0
	github.com/layer5io/learn-layer5/smi-conformance v0.0.0-20210317075357-06b4f88b3e34
	github.com/layer5io/meshkit v0.5.8
	github.com/layer5io/service-mesh-performance v0.3.2
	github.com/spf13/viper v1.8.1
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc v0.11.0
	go.opentelemetry.io/otel v0.11.0
	go.opentelemetry.io/otel/exporters/trace/jaeger v0.11.0
	go.opentelemetry.io/otel/sdk v0.11.0
	golang.org/x/sys v0.0.0-20220110181412-a018aaa089fe // indirect
	google.golang.org/grpc v1.38.0
	google.golang.org/protobuf v1.26.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/client-go v0.21.0
)
