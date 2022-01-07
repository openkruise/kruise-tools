module github.com/openkruise/kruise-tools

go 1.16

require (
	github.com/go-logr/logr v0.2.1 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/moby/term v0.0.0-20200312100748-672ec06f55cd // indirect
	github.com/openkruise/kruise-api v0.10.0
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.6.1
	k8s.io/api v0.21.6
	k8s.io/apimachinery v0.21.6
	k8s.io/cli-runtime v0.21.6
	k8s.io/client-go v0.21.6
	k8s.io/component-base v0.21.6
	k8s.io/klog/v2 v2.4.0
	k8s.io/kubectl v0.21.6
	sigs.k8s.io/controller-runtime v0.6.3
)

// Replace to match K8s 1.20.12
replace (
	k8s.io/api => k8s.io/api v0.20.12
	k8s.io/apimachinery => k8s.io/apimachinery v0.20.12
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.20.12
	k8s.io/client-go => k8s.io/client-go v0.20.12
	k8s.io/code-generator => k8s.io/code-generator v0.20.12
	k8s.io/component-base => k8s.io/component-base v0.20.12
	k8s.io/component-helpers => k8s.io/component-helpers v0.20.12
	k8s.io/kubectl => k8s.io/kubectl v0.20.12
	k8s.io/metrics => k8s.io/metrics v0.20.12
)
