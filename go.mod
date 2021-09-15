module github.com/talos-systems/capi-utils

go 1.16

require (
	github.com/spf13/cobra v1.1.3
	github.com/talos-systems/go-debug v0.2.1
	github.com/talos-systems/go-retry v0.3.1
	github.com/talos-systems/talos/pkg/machinery v0.11.5
	google.golang.org/grpc v1.40.0
	k8s.io/api v0.17.9
	k8s.io/apimachinery v0.17.9
	k8s.io/client-go v0.17.9
	sigs.k8s.io/cluster-api v0.3.23
	sigs.k8s.io/controller-runtime v0.5.14
)
