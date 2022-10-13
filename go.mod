module github.com/openkruise/kruise-tools

go 1.16

require (
	github.com/armon/consul-api v0.0.0-20180202201655-eb2c6b5be1b6 // indirect
	github.com/coreos/go-etcd v2.0.0+incompatible // indirect
	github.com/cpuguy83/go-md2man v1.0.10 // indirect
	github.com/docker/docker v0.7.3-0.20190327010347-be7ac8be2ae0 // indirect
	github.com/go-errors/errors v1.0.1 // indirect
	github.com/go-openapi/validate v0.19.5 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/moby/term v0.0.0-20210610120745-9d4ed1856297
	github.com/openkruise/kruise-api v1.0.0
	github.com/openkruise/rollouts v0.1.0
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/spf13/viper v1.7.0 // indirect
	github.com/stretchr/testify v1.7.0
	github.com/ugorji/go/codec v0.0.0-20181204163529-d75b2dcb6bc8 // indirect
	github.com/xordataexchange/crypt v0.0.3-0.20170626215501-b2862e3d0a77 // indirect
	go.etcd.io/etcd v0.0.0-20191023171146-3cf2f69b5738 // indirect
	gotest.tools v2.2.0+incompatible // indirect
	k8s.io/api v0.22.6
	k8s.io/apimachinery v0.22.6
	k8s.io/cli-runtime v0.22.6
	k8s.io/client-go v0.22.6
	k8s.io/component-base v0.22.6
	k8s.io/klog v1.0.0 // indirect
	k8s.io/klog/v2 v2.9.0
	k8s.io/kubectl v0.21.6
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a // indirect
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/kustomize/api v0.8.11 // indirect
	sigs.k8s.io/kustomize/kyaml v0.11.0 // indirect
	sigs.k8s.io/structured-merge-diff/v3 v3.0.0 // indirect
)

// Replace to match K8s 1.20.12
replace (
	k8s.io/api => k8s.io/api v0.22.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.22.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.22.6
	k8s.io/client-go => k8s.io/client-go v0.22.6
	k8s.io/code-generator => k8s.io/code-generator v0.22.6
	k8s.io/component-base => k8s.io/component-base v0.22.6
	k8s.io/component-helpers => k8s.io/component-helpers v0.22.6
	k8s.io/kubectl => k8s.io/kubectl v0.22.6
	k8s.io/metrics => k8s.io/metrics v0.22.6
)
