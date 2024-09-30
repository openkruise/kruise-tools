## kubectl-kruise set env

Update environment variables on a pod template

### Synopsis

Update environment variables on a pod template.

 List environment variable definitions in one or more pods, pod templates. Add, update, or remove container environment variable definitions in one or more pod templates (within replication controllers or deployment configurations). View or modify the environment variable definitions on all containers in the specified pods or pod templates, or just those that match a wildcard.

 If "--env -" is passed, environment variables can be read from STDIN using the standard env syntax.

 Possible resources include (case insensitive):

  pod (po), deployment (deploy), statefulset (sts), daemonset (ds), replicaset (rs), cloneset (clone), statefulset.apps.kruise.io (asts), daemonset.apps.kruise.io (ads)

```
kubectl-kruise set env RESOURCE/NAME KEY_1=VAL_1 ... KEY_N=VAL_N
```

### Examples

```
  # Update cloneset 'sample' with a new environment variable
  kubectl-kruise set env cloneset/sample STORAGE_DIR=/local
  
  # List the environment variables defined on a cloneset 'sample'
  kubectl-kruise set env cloneset/sample --list
  
  # List the environment variables defined on all pods
  kubectl-kruise set env pods --all --list
  
  # Output modified cloneset in YAML, and does not alter the object on the server
  kubectl-kruise set env cloneset/sample STORAGE_DIR=/data -o yaml
  
  # Update all containers in all replication controllers in the project to have ENV=prod
  kubectl-kruise set env rc --all ENV=prod
  
  # Import environment from a secret
  kubectl-kruise set env --from=secret/mysecret cloneset/sample
  
  # Import environment from a config map with a prefix
  kubectl-kruise set env --from=configmap/myconfigmap --prefix=MYSQL_ cloneset/sample
  
  # Import specific keys from a config map
  kubectl-kruise set env --keys=my-example-key --from=configmap/myconfigmap cloneset/sample
  
  # Remove the environment variable ENV from container 'c1' in all deployment configs
  kubectl-kruise set env clonesets --all --containers="c1" ENV-
  
  # Remove the environment variable ENV from a deployment definition on disk and
  # update the deployment config on the server
  kubectl-kruise set env -f deploy.json ENV-
  
  # Set some of the local shell environment into a deployment config on the server
  env | grep RAILS_ | kubectl-kruise set env -e - cloneset/sample
```

### Options

```
      --all                            If true, select all resources in the namespace of the specified resource types
      --allow-missing-template-keys    If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
  -c, --containers string              The names of containers in the selected pod templates to change - may use wildcards (default "*")
      --dry-run string[="unchanged"]   Must be "none", "server", or "client". If client strategy, only print the object that would be sent, without sending it. If server strategy, submit server-side request without persisting the resource. (default "none")
  -e, --env stringArray                Specify a key-value pair for an environment variable to set into each container.
  -f, --filename strings               Filename, directory, or URL to files the resource to update the env
      --from string                    The name of a resource from which to inject environment variables
  -h, --help                           help for env
      --keys strings                   Comma-separated list of keys to import from specified resource
  -k, --kustomize string               Process the kustomization directory. This flag can't be used together with -f or -R.
      --list                           If true, display the environment and any changes in the standard format. this flag will removed when we have kubectl view env.
      --local                          If true, set env will NOT contact api-server but run locally.
  -o, --output string                  Output format. One of: (json, yaml, name, go-template, go-template-file, template, templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
      --overwrite                      If true, allow environment to be overwritten, otherwise reject updates that overwrite existing environment. (default true)
      --prefix string                  Prefix to append to variable names
  -R, --recursive                      Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
      --resolve                        If true, show secret or configmap references when listing variables
  -l, --selector string                Selector (label query) to filter on
      --show-managed-fields            If true, keep the managedFields when printing objects in JSON or YAML format.
      --template string                Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
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

* [kubectl-kruise set](kubectl-kruise_set.md)	 - Set specific features on objects

