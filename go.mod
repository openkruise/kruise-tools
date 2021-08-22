module github.com/openkruise/kruise-tools

go 1.16

require (
	github.com/davecgh/go-spew v1.1.1
	github.com/googleapis/gnostic v0.4.0
	github.com/kr/pretty v0.2.0 // indirect
	github.com/lithammer/dedent v1.1.0
	github.com/openkruise/kruise-api v0.8.0
	github.com/pkg/errors v0.9.1
	github.com/spf13/cobra v1.1.3
	github.com/spf13/pflag v1.0.5
	github.com/stretchr/testify v1.5.1
	golang.org/x/text v0.3.3
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.21.0
	k8s.io/apimachinery v0.21.0
	k8s.io/cli-runtime v0.21.0
	k8s.io/client-go v0.21.0
	k8s.io/component-base v0.21.0
	k8s.io/klog v1.0.0
	k8s.io/klog/v2 v2.8.0 // indirect
	k8s.io/kube-openapi v0.0.0-20200410145947-61e04a5be9a6
	k8s.io/kubectl v0.21.0
	sigs.k8s.io/controller-runtime v0.6.3
	sigs.k8s.io/kustomize v2.0.3+incompatible
	sigs.k8s.io/yaml v1.2.0
)

// Replace to match K8s 1.18.6
replace (
	k8s.io/api => k8s.io/api v0.18.4
	k8s.io/apimachinery => k8s.io/apimachinery v0.18.4
	k8s.io/cli-runtime => k8s.io/cli-runtime v0.18.4
	k8s.io/client-go => k8s.io/client-go v0.18.4
	k8s.io/component-base => k8s.io/component-base v0.18.4
	k8s.io/kubectl v0.21.0 => k8s.io/kubectl v0.18.4
	k8s.io/metrics => k8s.io/metrics v0.18.4
)
