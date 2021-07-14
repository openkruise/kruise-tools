module github.com/openkruise/kruise-tools

go 1.16

require (
	github.com/coreos/bbolt v1.3.3 // indirect
	github.com/coreos/etcd v3.3.17+incompatible // indirect
	github.com/golangplus/bytes v0.0.0-20160111154220-45c989fe5450 // indirect
	github.com/golangplus/fmt v0.0.0-20150411045040-2a5d6d7d2995 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/openkruise/kruise-api v0.8.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/xlab/handysort v0.0.0-20150421192137-fb3537ed64a1 // indirect
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/component-base v0.21.0
	k8s.io/klog v1.0.0
	k8s.io/kubectl v0.21.0
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/structured-merge-diff v1.0.1 // indirect
	sigs.k8s.io/structured-merge-diff/v2 v2.0.1 // indirect
	vbom.ml/util v0.0.0-20160121211510-db5cfe13f5cc // indirect
)

// Replace to match K8s 1.18.6
replace (
	k8s.io/api => k8s.io/api v0.18.6
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.6
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.6
	k8s.io/client-go => k8s.io/client-go v0.18.6
	k8s.io/component-base => k8s.io/component-base v0.18.6
	k8s.io/kubectl v0.21.0 => k8s.io/kubectl v0.18.6
	k8s.io/metrics => k8s.io/metrics v0.18.6
)
