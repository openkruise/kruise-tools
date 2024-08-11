## kubectl-kruise set serviceaccount

Update ServiceAccount of a resource

### Synopsis

Update ServiceAccount of pod template resources.

 Possible resources (case insensitive) can be:

 replicationcontroller (rc), deployment (deploy), daemonset (ds), job, replicaset (rs), statefulset, cloneset (cs)

```
kubectl-kruise set serviceaccount (-f FILENAME | TYPE NAME) SERVICE_ACCOUNT
```

### Examples

```
  # Set Deployment nginx-deployment's ServiceAccount to serviceaccount1
  kubectl-kruise set serviceaccount cloneset sample serviceaccount1
  
  # Print the result (in yaml format) of updated cloneset with serviceaccount from local file, without hitting apiserver
  kubectl-kruise set sa -f CloneSet.yaml serviceaccount1 --local --dry-run=client -o yaml
```

### Options

```
      --all                            Select all resources, including uninitialized ones, in the namespace of the specified resource types
      --allow-missing-template-keys    If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
      --dry-run string[="unchanged"]   Must be "none", "server", or "client". If client strategy, only print the object that would be sent, without sending it. If server strategy, submit server-side request without persisting the resource. (default "none")
  -f, --filename strings               Filename, directory, or URL to files identifying the resource to get from a server.
  -h, --help                           help for serviceaccount
  -k, --kustomize string               Process the kustomization directory. This flag can't be used together with -f or -R.
      --local                          If true, set serviceaccount will NOT contact api-server but run locally.
  -o, --output string                  Output format. One of: (json, yaml, name, go-template, go-template-file, template, templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
  -R, --recursive                      Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
      --show-managed-fields            If true, keep the managedFields when printing objects in JSON or YAML format.
      --template string                Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                  UID to impersonate for the operation.
      --cache-dir string               Default cache directory (default "${HOME}/.kube/cache")
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

###### Auto generated by spf13/cobra on 11-Aug-2024
