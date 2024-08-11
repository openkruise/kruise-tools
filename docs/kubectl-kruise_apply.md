## kubectl-kruise apply

Apply a configuration to a resource by file name or stdin

### Synopsis

Apply a configuration to a resource by file name or stdin. The resource name must be specified. This resource will be created if it doesn't exist yet. To use 'apply', always create the resource initially with either 'apply' or 'create --save-config'.

 JSON and YAML formats are accepted.

 Alpha Disclaimer: the --prune functionality is not yet complete. Do not use unless you are aware of what the current state is. See https://issues.k8s.io/34274.

```
kubectl-kruise apply (-f FILENAME | -k DIRECTORY)
```

### Examples

```
  # Apply the configuration in pod.json to a pod
  kubectl apply -f ./pod.json
  
  # Apply resources from a directory containing kustomization.yaml - e.g. dir/kustomization.yaml
  kubectl apply -k dir/
  
  # Apply the JSON passed into stdin to a pod
  cat pod.json | kubectl apply -f -
  
  # Apply the configuration from all files that end with '.json' - i.e. expand wildcard characters in file names
  kubectl apply -f '*.json'
  
  # Note: --prune is still in Alpha
  # Apply the configuration in manifest.yaml that matches label app=nginx and delete all other resources that are not in the file and match label app=nginx
  kubectl apply --prune -f manifest.yaml -l app=nginx
  
  # Apply the configuration in manifest.yaml and delete all the other config maps that are not in the file
  kubectl apply --prune -f manifest.yaml --all --prune-whitelist=core/v1/ConfigMap
```

### Options

```
      --all                             Select all resources in the namespace of the specified resource types.
      --allow-missing-template-keys     If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
      --cascade string[="background"]   Must be "background", "orphan", or "foreground". Selects the deletion cascading strategy for the dependents (e.g. Pods created by a ReplicationController). Defaults to background. (default "background")
      --dry-run string[="unchanged"]    Must be "none", "server", or "client". If client strategy, only print the object that would be sent, without sending it. If server strategy, submit server-side request without persisting the resource. (default "none")
      --field-manager string            Name of the manager used to track field ownership. (default "kubectl-client-side-apply")
  -f, --filename strings                that contains the configuration to apply
      --force                           If true, immediately remove resources from API and bypass graceful deletion. Note that immediate deletion of some resources may result in inconsistency or data loss and requires confirmation.
      --force-conflicts                 If true, server-side apply will force the changes against conflicts.
      --grace-period int                Period of time in seconds given to the resource to terminate gracefully. Ignored if negative. Set to 1 for immediate shutdown. Can only be set to 0 when --force is true (force deletion). (default -1)
  -h, --help                            help for apply
  -k, --kustomize string                Process a kustomization directory. This flag can't be used together with -f or -R.
      --openapi-patch                   If true, use openapi to calculate diff when the openapi presents and the resource can be found in the openapi spec. Otherwise, fall back to use baked-in types. (default true)
  -o, --output string                   Output format. One of: (json, yaml, name, go-template, go-template-file, template, templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
      --overwrite                       Automatically resolve conflicts between the modified and live configuration by using values from the modified configuration (default true)
      --prune                           Automatically delete resource objects, that do not appear in the configs and are created by either apply or create --save-config. Should be used with either -l or --all.
      --prune-whitelist stringArray     Overwrite the default whitelist with <group/version/kind> for --prune
  -R, --recursive                       Process the directory used in -f, --filename recursively. Useful when you want to manage related manifests organized within the same directory.
  -l, --selector string                 Selector (label query) to filter on, supports '=', '==', and '!='.(e.g. -l key1=value1,key2=value2). Matching objects must satisfy all of the specified label constraints.
      --server-side                     If true, apply runs in the server instead of the client.
      --show-managed-fields             If true, keep the managedFields when printing objects in JSON or YAML format.
      --template string                 Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
      --timeout duration                The length of time to wait before giving up on a delete, zero means determine a timeout from the size of the object
      --validate string                 Must be one of: strict (or true), warn, ignore (or false).
                                        		"true" or "strict" will use a schema to validate the input and fail the request if invalid. It will perform server side validation if ServerSideFieldValidation is enabled on the api-server, but will fall back to less reliable client-side validation if not.
                                        		"warn" will warn about unknown or duplicate fields without blocking the request if server-side field validation is enabled on the API server, and behave as "ignore" otherwise.
                                        		"false" or "ignore" will not perform any schema validation, silently dropping any unknown or duplicate fields. (default "strict")
      --wait                            If true, wait for resources to be gone before returning. This waits for finalizers.
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
* [kubectl-kruise apply edit-last-applied](kubectl-kruise_apply_edit-last-applied.md)	 - Edit latest last-applied-configuration annotations of a resource/object
* [kubectl-kruise apply set-last-applied](kubectl-kruise_apply_set-last-applied.md)	 - Set the last-applied-configuration annotation on a live object to match the contents of a file
* [kubectl-kruise apply view-last-applied](kubectl-kruise_apply_view-last-applied.md)	 - View the latest last-applied-configuration annotations of a resource/object

###### Auto generated by spf13/cobra on 11-Aug-2024
