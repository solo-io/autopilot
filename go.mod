module github.com/solo-io/autopilot

go 1.13

require (
	github.com/bxcodec/faker v2.0.1+incompatible // indirect
	github.com/chai2010/gettext-go v0.0.0-20170215093142-bf70f2a70fb1 // indirect
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.12+incompatible // indirect
	github.com/coreos/go-semver v0.3.0 // indirect
	github.com/coreos/pkg v0.0.0-20180928190104-399ea9e2e55f // indirect
	github.com/docker/docker v1.13.1 // indirect
	github.com/elazarl/goproxy v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/elazarl/goproxy/ext v0.0.0-20190421051319-9d40249d3c2f // indirect
	github.com/fsnotify/fsnotify v1.4.7
	github.com/gertd/go-pluralize v0.1.1
	github.com/go-logr/logr v0.1.0
	github.com/gobuffalo/envy v1.7.1 // indirect
	github.com/gobuffalo/packr v1.30.1
	github.com/gogo/protobuf v1.3.1
	github.com/golang/protobuf v1.3.2
	github.com/gophercloud/gophercloud v0.2.0 // indirect
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0 // indirect
	github.com/grpc-ecosystem/grpc-gateway v1.5.0 // indirect
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334
	github.com/ilackarms/protokit v0.0.0-20181231193355-ee2393f3bbf0
	github.com/mitchellh/go-homedir v1.1.0
	github.com/onsi/ginkgo v1.11.0
	github.com/onsi/gomega v1.8.1
	github.com/otiai10/copy v1.0.2 // indirect
	github.com/pborman/uuid v1.2.0 // indirect
	github.com/peterbourgon/diskv v2.0.1+incompatible // indirect
	github.com/pkg/errors v0.8.1
	github.com/prometheus/client_golang v1.2.1
	github.com/prometheus/common v0.7.0
	github.com/rogpeppe/go-internal v1.5.0
	github.com/sirupsen/logrus v1.4.2
	github.com/soheilhy/cmux v0.1.4 // indirect
	github.com/solo-io/go-utils v0.11.7
	github.com/solo-io/solo-kit v0.6.3
	github.com/spf13/cobra v0.0.5
	github.com/spf13/pflag v1.0.5
	github.com/technosophos/moniker v0.0.0-20180509230615-a5dbd03a2245 // indirect
	github.com/tmc/grpc-websocket-proxy v0.0.0-20190109142713-0ad062ec5ee5 // indirect
	github.com/ugorji/go v1.1.5-pre // indirect
	github.com/xiang90/probing v0.0.0-20190116061207-43a291ad63a2 // indirect
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
	go.etcd.io/bbolt v1.3.3 // indirect
	go.uber.org/zap v1.13.0
	golang.org/x/tools v0.0.0-20191029190741-b9c20aec41a5
	gopkg.in/airbrake/gobrake.v2 v2.0.9 // indirect
	gopkg.in/gemnasium/logrus-airbrake-hook.v2 v2.1.2 // indirect
	istio.io/api v0.0.0-20191109011807-2629c6ac1513
	istio.io/client-go v0.0.0-20191104174404-7b65e62d85b0
	k8s.io/api v0.0.0
	k8s.io/apiextensions-apiserver v0.0.0
	k8s.io/apimachinery v0.0.0
	k8s.io/apiserver v0.0.0-20190111033246-d50e9ac5404f+incompatible // indirect
	k8s.io/client-go v11.0.0+incompatible
	k8s.io/code-generator v0.0.0
	k8s.io/gengo v0.0.0-20191010091904-7fa3014cb28f
	k8s.io/helm v2.13.0+incompatible // indirect
	k8s.io/kubernetes v1.13.2 // indirect
	k8s.io/utils v0.0.0-20191010214722-8d271d903fe4
	knative.dev/pkg v0.0.0-20191111150521-6d806b998379
	sigs.k8s.io/controller-runtime v0.4.0
	sigs.k8s.io/yaml v1.1.0
	vbom.ml/util v0.0.0-20180919145318-efcd4e0f9787 // indirect
)

// Pinned to kubernetes-1.14.1
replace (
	github.com/Sirupsen/logrus => github.com/sirupsen/logrus v1.4.2
	k8s.io/api => k8s.io/api v0.0.0-20190409021203-6e4e0e4f393b
	k8s.io/apiextensions-apiserver => k8s.io/apiextensions-apiserver v0.0.0-20190409022649-727a075fdec8
	k8s.io/apimachinery => k8s.io/apimachinery v0.0.0-20190404173353-6a84e37a896d
	k8s.io/apiserver => k8s.io/apiserver v0.0.0-20190409021813-1ec86e4da56c
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.0.0-20190409023024-d644b00f3b79
	k8s.io/client-go => k8s.io/client-go v11.0.1-0.20190409021438-1a26190bd76a+incompatible
	k8s.io/cloud-provider => k8s.io/cloud-provider v0.0.0-20190409023720-1bc0c81fa51d
	k8s.io/code-generator => k8s.io/code-generator v0.0.0-20190311093542-50b561225d70
	k8s.io/kube-aggregator => k8s.io/kube-aggregator v0.0.0-20190409022021-00b8e31abe9d
	k8s.io/kube-openapi => k8s.io/kube-openapi v0.0.0-20190510232812-a01b7d5d6c22
	k8s.io/kubernetes => k8s.io/kubernetes v1.14.1
)

replace (
	// Indirect operator-sdk dependencies use git.apache.org, which is frequently
	// down. The github mirror should be used instead.
	// Locking to a specific version (from 'go mod graph'):
	git.apache.org/thrift.git => github.com/apache/thrift v0.0.0-20180902110319-2566ecd5d999
	github.com/operator-framework/operator-lifecycle-manager => github.com/operator-framework/operator-lifecycle-manager v0.0.0-20190605231540-b8a4faf68e36
)

// Remove when controller-tools v0.2.2 is released
// Required for the bugfix https://github.com/kubernetes-sigs/controller-tools/pull/322
replace sigs.k8s.io/controller-tools => sigs.k8s.io/controller-tools v0.2.2-0.20190919011008-6ed4ff330711

replace github.com/solo-io/solo-kit => ../solo-kit

replace github.com/docker/docker => github.com/moby/moby v0.7.3-0.20190826074503-38ab9da00309
