echo `pwd`
ver="v0.26.10"
kver="v1.26.10"
go mod edit -require=k8s.io/kubernetes@${kver}
go mod edit -require=sigs.k8s.io/controller-runtime@v0.14.7
go mod edit -require=k8s.io/apimachinery@${ver}
go mod edit -require=k8s.io/api@${ver}
go mod edit -require=k8s.io/client-go@${ver}
go mod edit -require=k8s.io/apiextensions-apiserver@${ver}
go mod edit -require=k8s.io/component-base@${ver}
go mod edit -require=k8s.io/apiserver@${ver}
go mod edit -require=k8s.io/kms@${ver}
go mod edit -require=k8s.io/kubectl@${ver}
go mod edit -require=k8s.io/kube-aggregator@${ver}
go mod edit -require=k8s.io/component-helpers@${ver}
go mod edit -require=k8s.io/cli-runtime@${ver}
go mod edit -require=k8s.io/metrics@${ver}