## kubectl-kruise rollout status

Show the status of the rollout

### Synopsis

Show the status of the rollout.

 By default 'rollout status' will watch the status of the latest rollout until it's done. If you don't want to wait for the rollout to finish then you can use --watch=false. Note that if a new rollout starts in-between, then 'rollout status' will continue watching the latest revision. If you want to pin to a specific revision and abort if it is rolled over by another revision, use --revision=N where N is the revision you need to watch for.

```
kubectl-kruise rollout status (TYPE NAME | TYPE/NAME) [flags]
```

### Examples

```
  # Watch the rollout status of a deployment
  kubectl-kruise rollout status deployment/nginx
  
  # Watch the rollout status of a cloneset
  kubectl-kruise rollout status cloneset/nginx
  
  # Watch the rollout status of a advanced statefulset
  kubectl-kruise rollout status asts/nginx
```

### Options

```
  -d, --detail             Show the detail status of the rollout.
  -f, --filename strings   Filename, directory, or URL to files identifying the resource to get from a server.
  -h, --help               help for status
  -k, --kustomize string   Process the kustomization directory. This flag can't be used together with -f or -R.
  -R, --recursive          Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
      --revision int       Pin to a specific revision for showing its status. Defaults to 0 (last revision).
      --timeout duration   The length of time to wait before ending watch, zero means never. Any other values should contain a corresponding time unit (e.g. 1s, 2m, 3h).
  -w, --watch              Watch the status of the rollout until it's done. (default true)
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                  UID to impersonate for the operation.
      --cache-dir string               Default cache directory (default "$HOME/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --client-certificate string      Path to a client certificate file for TLS
      --client-key string              Path to a client key file for TLS
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --disable-compression            If true, opt-out of response compression for all requests to the server
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              Path to the kubeconfig file to use for CLI requests.
      --match-server-version           Require server version to match client version
  -n, --namespace string               If present, the namespace scope for this CLI request
      --password string                Password for basic authentication to the API server
      --profile string                 Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex) (default "none")
      --profile-output string          Name of the file to write the profile to (default "profile.pprof")
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --token string                   Bearer token for authentication to the API server
      --user string                    The name of the kubeconfig user to use
      --username string                Username for basic authentication to the API server
      --warnings-as-errors             Treat warnings received from the server as errors and exit with a non-zero exit code
```

### SEE ALSO

* [kubectl-kruise rollout](kubectl-kruise_rollout.md)	 - Manage the rollout of a resource

###### Auto generated by spf13/cobra on 12-Mar-2025
