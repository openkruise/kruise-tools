# Kruise-tools
kubectl plugin for OpenKruise

[Kruise-tools](https://github.com/openkruise/kruise-tools) provides commandline tools for kruise features, such as `kubectl-kruise`, which is a standard plugin of `kubectl`.

## Install
### Install via Krew
1. [Krew](https://krew.sigs.k8s.io/) itself is a kubectl plugin that is installed and updated via Krew (yes, Krew self-hosts).
   First, [install krew](https://krew.sigs.k8s.io/docs/user-guide/setup/install/).

2. Run `kubectl krew install kruise` to install kruise plugin via Krew.

3. Then you can use it with `kubectl-kruise` or `kubectl kruise`.

```bash
$ kubectl-kruise --help

# or
$ kubectl kruise --help
```
### Install manually
1. You can simply download the binary from the [releases](https://github.com/openkruise/kruise-tools/releases) page. Currently `linux`, `darwin`(OS X), `windows` with `x86_64` and `arm64` are provided. If you are using some other systems or architectures, you have to download the source code and execute `make build` to build the binary.

2. Extract and move it to system PATH.

```bash
$ tar xvf kubectl-kruise-darwin-amd64.tar.gz
$ mv darwin-amd64/kubectl-kruise /usr/local/bin/
```

3. Then you can use it with `kubectl-kruise` or `kubectl kruise`.

```bash
$ kubectl-kruise --help

# or
$ kubectl kruise --help
```

## Upgrade
### Upgrade via krew
Run `kubectl krew upgrade kruise` to upgrade kruise plugin via Krew.

### Upgrade manually
Same to `install manually`.

## Usage

### completion
```bash
To load auto completions:

Bash:
  $ source <(kubectl-kruise completion bash)

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it.  You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ kubectl-kruise completion zsh > "${fpath[1]}/_kubectl-kruise"

Fish:
  $ kubectl-kruise completion fish | source

PowerShell:
  PS> kubectl-kruise completion powershell | Out-String | Invoke-Expression

### expose

Take a workload(e.g. deployment, cloneset), service or pod and expose it as a new Kubernetes Service.

```bash
$ kubectl kruise expose cloneset nginx --port=80 --target-port=8000
```

### scale

Set a new size for a Deployment, ReplicaSet, CloneSet, or Advanced StatefulSet.

```bash
$ kubectl kruise scale --replicas=3 cloneset nginx
```

It equals to `kubectl scale --replicas=3 cloneset nginx`.

### rollout

Available commands: `history`, `pause`, `restart`, `resume`, `status`, `undo`, `approve`.

```bash
$ kubectl kruise rollout undo cloneset/nginx

# built-in statefulsets
$ kubectl kruise rollout status statefulsets/sts1

# kruise statefulsets
$ kubectl kruise rollout status statefulsets.apps.kruise.io/sts2

# approve a kruise rollout resource named "rollout-demo" in "ns-demo" namespace
$ kubectl kruise rollout approve rollout/rollout-demo -n ns-demo`

# undo a kruise rollout resource
$ kubectl kruise rollout undo rollout/rollout-demo
```

### set

Available commands: `env`, `image`, `resources`, `selector`, `serviceaccount`, `subject`.

```bash
$ kubectl kruise set env cloneset/nginx STORAGE_DIR=/local

$ kubectl kruise set image cloneset/nginx busybox=busybox nginx=nginx:1.9.1
```

### migrate

Currently it supports migrate from Deployment to CloneSet.

```bash
# Create an empty CloneSet from an existing Deployment.
$ kubectl kruise migrate CloneSet --from Deployment -n default --dst-name deployment-name --create

# Create a same replicas CloneSet from an existing Deployment.
$ kubectl kruise migrate CloneSet --from Deployment -n default --dst-name deployment-name --create --copy

# Migrate replicas from an existing Deployment to an existing CloneSet.
$ kubectl-kruise migrate CloneSet --from Deployment -n default --src-name cloneset-name --dst-name deployment-name --replicas 10 --max-surge=2
```

### scaledown

Scaledown a cloneset with selective Pods.

```bash
# Scale down 2 with  selective pods
$ kubectl kruise scaledown cloneset/nginx --pods pod-a,pod-b
```

It will decrease **replicas=replicas-2** of this cloneset and delete the specified pods.

### exec

Exec working sidecar container of pod when sidecarset is hot-upgrade.

```bash
# Get output from running 'date' command in working sidecar container from pod mypod
kubectl kruise exec mypod -S sidecar-container -- date

# Switch to raw terminal mode, sends stdin to 'bash' in working sidecar container from cloneset myclone 
# and sends stdout/stderr from 'bash' back to the client
kubectl kruise exec clone/myclone -S sidecar-container -it -- bash
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
