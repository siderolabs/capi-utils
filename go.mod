module github.com/talos-systems/capi-utils

go 1.16

replace (
	// See https://github.com/talos-systems/go-loadbalancer/pull/4
	// `go get github.com/smira/tcpproxy@combined-fixes`, then copy pseudo-version there
	inet.af/tcpproxy => github.com/smira/tcpproxy v0.0.0-20201015133617-de5f7797b95b

	// keep older versions of k8s.io packages to keep compatiblity with cluster-api
	k8s.io/api v0.21.3 => k8s.io/api v0.20.5
	k8s.io/apimachinery v0.21.3 => k8s.io/apimachinery v0.20.5
	k8s.io/client-go v0.21.3 => k8s.io/client-go v0.20.5

	sigs.k8s.io/cluster-api v0.3.20 => sigs.k8s.io/cluster-api v0.3.9
)

require (
	github.com/spf13/cobra v1.1.3
	github.com/talos-systems/go-debug v0.2.1
	github.com/talos-systems/go-retry v0.3.1
	github.com/talos-systems/talos/pkg/machinery v0.11.5
	google.golang.org/grpc v1.40.0
	k8s.io/api v0.21.3
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	k8s.io/klog/v2 v2.8.0 // indirect
	sigs.k8s.io/cluster-api v0.3.20
	sigs.k8s.io/controller-runtime v0.6.3
)
