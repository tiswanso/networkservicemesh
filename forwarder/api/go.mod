module cisco-app-networking.github.io/networkservicemesh/forwarder/api

go 1.13

require (
	github.com/golang/protobuf v1.3.2
	google.golang.org/grpc v1.27.0
)

replace (
	github.com/census-instrumentation/opencensus-proto v0.1.0-0.20181214143942-ba49f56771b8 => github.com/census-instrumentation/opencensus-proto v0.0.3-0.20181214143942-ba49f56771b8
	github.com/networkservicemesh/networkservicemesh/controlplane/api => cisco-app-networking.github.io/networkservicemesh/controlplane/api latest
	github.com/networkservicemesh/networkservicemesh/forwarder/api => ./
)
