# kubectl-kruise
kubectl plugin for OpenKruise

`kubectl` supports a plug-in mechanism, but the rollout and other related operations provided by this tool itself only support the native workload resources of Kubernetes.
Therefore, we need to create a kubectl plugin for OpenKruise, through which community users can use kubectl to operate Kruise’s workload resources.

So, `kubectl-kruise` was created.

### How to use
The development of  `kubectl-kruise`  is in progress, if you wanna to experience it, you can clone it and make it:

```
make build && cp bin/kubectl-kruise /usr/local/bin

```

### Use with command-line 

Then you can operate Openkruise resource by `kubectl-kruise`.
By now the `rollout` cmd such as `rollout undo`, `rollout status`, `rollout history` have been developed.

![](https://tva1.sinaimg.cn/large/008i3skNgy1gqmmcx5nlqj31eo0je420.jpg)


```bash
$kubectl-kruise --help
kubectl-kruise controls the OpenKruise manager.

 Find more information at: https://openkruise.io/

Aliases:
kubectl-kruise, kk

CloneSet Commands:
  rollout       Manage the rollout of a resource
  set           Set specific features on objects
  migrate       Migrate from K8s original workloads to Kruise workloads

AdvancedStatefulSet Commands:
  rollout       Manage the rollout of a resource
  set           Set specific features on objects

Basic Commands:
  scale         Set a new size for a CloneSet, Deployment, ReplicaSet or Replication Controller
  autoscale     Auto-scale a CloneSet, Deployment, ReplicaSet, or ReplicationController

Cluster Management Commands:
  certificate   Modify certificate resources.
  cluster-info  Display cluster info
  top           Display Resource (CPU/Memory/Storage) usage.
  cordon        Mark node as unschedulable
  uncordon      Mark node as schedulable
  drain         Drain node in preparation for maintenance
  taint         Update the taints on one or more nodes

Troubleshooting and Debugging Commands:
  describe      Show details of a specific resource or group of resources
  logs          Print the logs for a container in a pod
  attach        Attach to a running container
  exec          Execute a command in a container
  port-forward  Forward one or more local ports to a pod
  debug         Attach a debug container to a running pod

Advanced Commands:
  diff          Diff live version against would-be applied version
  apply         Apply a configuration to a resource by filename or stdin
  patch         Update field(s) of a resource using strategic merge patch
  replace       Replace a resource by filename or stdin
  wait          Experimental: Wait for a specific condition on one or many resources.
  kustomize     Build a kustomization target from a directory or a remote url.

Other Commands:
  api-resources Print the supported API resources on the server
  api-versions  Print the supported API versions on the server, in the form of "group/version"
  config        Modify kubeconfig files
  plugin        Provides utilities for interacting with plugins.
  version       Print the client and server version information

Usage:
  kubectl-kruise [flags] [options]

Use "kubectl-kruise <command> --help" for more information about a given command.
Use "kubectl-kruise options" for a list of global command-line options (applies to all commands).
```

Currently it also supports to migrate Pods from Deployment to CloneSet by `kruise migrate [options]`.
You can also import `github.com/openkruise/kruise-tools/pkg/migration` and trigger migration with its api.

```bash
$ kubectl-kruise migrate --help
  kruise is a command-line tool to use Kruise.
  
  Usage:
    kruise [flags]
    kruise [command]
  
  Available Commands:
    help        Help about any command
    migrate     Migrate from K8s original workloads to Kruise workloads
  
  Flags:
        --as string                      Username to impersonate for the operation
        --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
        --cache-dir string               Default HTTP cache directory (default "/Users/wsy/.kube/http-cache")
        --certificate-authority string   Path to a cert file for the certificate authority
        --client-certificate string      Path to a client certificate file for TLS
        --client-key string              Path to a client key file for TLS
        --cluster string                 The name of the kubeconfig cluster to use
        --context string                 The name of the kubeconfig context to use
    -h, --help                           help for kruise
        --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
        --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
    -n, --namespace string               If present, the namespace scope for this CLI request
        --password string                Password for basic authentication to the API server
        --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
    -s, --server string                  The address and port of the Kubernetes API server
        --token string                   Bearer token for authentication to the API server
        --user string                    The name of the kubeconfig user to use
        --username string                Username for basic authentication to the API server
  
  Use "kubectl-kruise [command] --help" for more information about a command.
```

### TODO
#### kubectl kruise migrate
   * [x] migrate [options]

> kubectl-kruise migrate demo
```bash
kubectl kruise migrate CloneSet --from Deployment --src-name deployment-demo --dst-name cloneset-demo --create --copy
```
   
#### kubectl kruise rollout for CloneSet workload
   * [x] undo
   * [x] history
   * [x] status
   * [x] pause
   * [x] resume
   * [x] restart
   
#### kubectl kruise rollout for Advanced StatefulSet
   * [x]  undo
   * [x] history
   * [x] status
   * [x] restart

#### kubectl kruise expose for CloneSet workload
   * [x] kubectl kruise expose cloneset demo-clone  --port=80 --target-port=8000
   
#### kubectl kruise set SUBCOMMAND [options] for CloneSet
   * [x] kubectl kruise set image cloneset/abc
   * [x] kubectl kruise set env cloneset/abc
   * [x] kubectl kruise set serviceaccount cloneset/abc
   * [x] kubectl kruise set resources cloneset/abc
   
#### kubectl kruise set SUBCOMMAND [options] for Advanced StatefulSet
   * [x] kubectl kruise set image asts/abc
   * [x] kubectl kruise set env asts/abc
   * [x] kubectl kruise set serviceaccount asts/abc
   * [x] kubectl kruise set resources asts/abc
   
#### kubectl kruise autoscale SUBCOMMAND [options]
   * [ ] kubectl kruise autoscale 
 

### Contributing
We encourage you to help out by reporting issues, improving documentation, fixing bugs, or adding new features. 
