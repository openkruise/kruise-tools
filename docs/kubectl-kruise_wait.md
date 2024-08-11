## kubectl-kruise wait

Experimental: Wait for a specific condition on one or many resources

### Synopsis

Experimental: Wait for a specific condition on one or many resources.

 The command takes multiple resources and waits until the specified condition is seen in the Status field of every given resource.

 Alternatively, the command can wait for the given set of resources to be deleted by providing the "delete" keyword as the value to the --for flag.

 A successful message will be printed to stdout indicating when the specified condition has been met. You can use -o option to change to output destination.

```
kubectl-kruise wait ([-f FILENAME] | resource.group/resource.name | resource.group [(-l label | --all)]) [--for=delete|--for condition=available|--for=jsonpath='{}'=value]
```

### Examples

```
  # Wait for the pod "busybox1" to contain the status condition of type "Ready"
  kubectl wait --for=condition=Ready pod/busybox1
  
  # The default value of status condition is true; you can wait for other targets after an equal delimiter (compared after Unicode simple case folding, which is a more general form of case-insensitivity):
  kubectl wait --for=condition=Ready=false pod/busybox1
  
  # Wait for the pod "busybox1" to contain the status phase to be "Running".
  kubectl wait --for=jsonpath='{.status.phase}'=Running pod/busybox1
  
  # Wait for the pod "busybox1" to be deleted, with a timeout of 60s, after having issued the "delete" command
  kubectl delete pod/busybox1
  kubectl wait --for=delete pod/busybox1 --timeout=60s
```

### Options

```
      --all                           Select all resources in the namespace of the specified resource types
  -A, --all-namespaces                If present, list the requested object(s) across all namespaces. Namespace in current context is ignored even if specified with --namespace.
      --allow-missing-template-keys   If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
      --field-selector string         Selector (field query) to filter on, supports '=', '==', and '!='.(e.g. --field-selector key1=value1,key2=value2). The server only supports a limited number of field queries per type.
  -f, --filename strings              identifying the resource.
      --for string                    The condition to wait on: [delete|condition=condition-name[=condition-value]|jsonpath='{JSONPath expression}'=JSONPath Condition]. The default condition-value is true.  Condition values are compared after Unicode simple case folding, which is a more general form of case-insensitivity.
  -h, --help                          help for wait
      --local                         If true, annotation will NOT contact api-server but run locally.
  -o, --output string                 Output format. One of: (json, yaml, name, go-template, go-template-file, template, templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
  -R, --recursive                     Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory. (default true)
  -l, --selector string               Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2)
      --show-managed-fields           If true, keep the managedFields when printing objects in JSON or YAML format.
      --template string               Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
      --timeout duration              The length of time to wait before giving up.  Zero means check once and don't wait, negative means wait for a week. (default 30s)
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

* [kubectl-kruise](kubectl-kruise.md)	 - kubectl-kruise controls the OpenKruise CRs

###### Auto generated by spf13/cobra on 11-Aug-2024
