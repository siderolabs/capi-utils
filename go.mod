module github.com/talos-systems/capi-utils

go 1.16

require (
	github.com/spf13/cobra v1.2.1
	github.com/spf13/viper v1.9.0
	github.com/talos-systems/go-debug v0.2.1
	github.com/talos-systems/go-retry v0.3.1
	github.com/talos-systems/talos/pkg/machinery v0.12.3
	google.golang.org/grpc v1.40.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/cluster-api v1.0.0
	sigs.k8s.io/controller-runtime v0.10.2
)
