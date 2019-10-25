module github.com/solo-io/autopilot/examples/canary

require (
	github.com/envoyproxy/go-control-plane v0.9.0 // indirect
	github.com/go-openapi/spec v0.19.2
	github.com/gogo/googleapis v1.3.0 // indirect
	github.com/hashicorp/vault/api v1.0.4 // indirect
	github.com/mitchellh/hashstructure v1.0.0
	github.com/operator-framework/operator-sdk v0.11.1-0.20191024224924-17d389050d46
	github.com/pkg/errors v0.8.1
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/sirupsen/logrus v1.4.2
	github.com/solo-io/gloo v0.0.0-00010101000000-000000000000
	github.com/solo-io/go-utils v0.10.21 // indirect
	github.com/solo-io/solo-kit v0.11.5 // indirect
	github.com/spf13/pflag v1.0.3
	go.uber.org/zap v1.10.0
	k8s.io/api v0.0.0-20191025025715-ac1bc6bf0668
	k8s.io/apimachinery v0.0.0-20191025025535-ced427e1ea5f
	k8s.io/cli-runtime v0.0.0-20191025031152-0b44683c44df // indirect
	k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/kube-openapi v0.0.0-20190816220812-743ec37842bf
	sigs.k8s.io/controller-runtime v0.2.0
)

// Pinned to kubernetes-1.14.1
replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.2.0
	github.com/solo-io/gloo => github.com/solo-io/gloo v0.20.10-0.20191024223947-69a04e343c3c
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190918200908-1e17798da8c1

	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20191025031152-0b44683c44df // indirect
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
)

replace (
	// Indirect operator-sdk dependencies use git.apache.org, which is frequently
	// down. The github mirror should be used instead.
	// Locking to a specific version (from 'go mod graph'):
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/coreos/prometheus-operator => github.com/coreos/prometheus-operator v0.31.1
	// Pinned to v2.10.0 (kubernetes-1.14.1) so https://proxy.golang.org can
	// resolve it correctly.
	github.com/prometheus/prometheus => github.com/prometheus/prometheus v1.8.2-0.20190525122359-d20e84d0fb64
)

replace github.com/operator-framework/operator-sdk => github.com/operator-framework/operator-sdk v0.11.0

go 1.13
