module github.com/openkruise/kruise-tools

go 1.16

require (
	github.com/go-errors/errors v1.0.1
	github.com/lithammer/dedent v1.1.0
	github.com/moby/term v0.0.0-20210610120745-9d4ed1856297
	github.com/openkruise/kruise-api v1.0.0
	github.com/openkruise/rollouts v0.1.0
	github.com/spf13/cobra v1.4.0
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.7.0
	k8s.io/api v0.22.6
	k8s.io/apimachinery v0.22.6
	k8s.io/cli-runtime v0.22.6
	k8s.io/client-go v0.22.6
	k8s.io/component-base v0.22.6
	k8s.io/klog/v2 v2.9.0
	k8s.io/kubectl v0.21.6
	k8s.io/utils v0.0.0-20210819203725-bdf08cb9a70a
	sigs.k8s.io/controller-runtime v0.10.3
	sigs.k8s.io/kustomize/api v0.8.11
	sigs.k8s.io/kustomize/kyaml v0.11.0
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
