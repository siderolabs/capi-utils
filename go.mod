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
	github.com/go-logr/zapr v0.2.0 // indirect
	github.com/golang/groupcache v0.0.0-20210331224755-41bb18bfe9da // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/google/uuid v1.2.0 // indirect
	github.com/googleapis/gnostic v0.5.5 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/onsi/ginkgo v1.16.3 // indirect
	github.com/prometheus/client_golang v1.11.0 // indirect
	github.com/spf13/cobra v1.1.3
	github.com/spf13/viper v1.8.1 // indirect
	github.com/talos-systems/go-debug v0.2.1
	github.com/talos-systems/go-retry v0.3.1
	github.com/talos-systems/talos/pkg/machinery v0.11.5
	golang.org/x/crypto v0.0.0-20210503195802-e9a32991a82e // indirect
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	golang.org/x/oauth2 v0.0.0-20210622215436-a8dc77f794b6 // indirect
	golang.org/x/term v0.0.0-20210615171337-6886f2dfbf5b // indirect
	golang.org/x/time v0.0.0-20210611083556-38a9dc6acbc6 // indirect
	google.golang.org/grpc v1.40.0
	k8s.io/api v0.21.3
	k8s.io/apiextensions-apiserver v0.19.1 // indirect
	k8s.io/apimachinery v0.21.3
	k8s.io/client-go v0.21.3
	k8s.io/cluster-bootstrap v0.17.9 // indirect
	k8s.io/klog/v2 v2.8.0 // indirect
	k8s.io/kube-openapi v0.0.0-20210305001622-591a79e4bda7 // indirect
	k8s.io/utils v0.0.0-20210527160623-6fdb442a123b // indirect
	sigs.k8s.io/cluster-api v0.3.20
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/structured-merge-diff/v4 v4.1.2 // indirect
)
