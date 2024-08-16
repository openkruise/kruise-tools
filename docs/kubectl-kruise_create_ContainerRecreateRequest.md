## kubectl-kruise create ContainerRecreateRequest

Create a crr with the specified name.

### Synopsis

Create a crr with the specified name.

```
kubectl-kruise create ContainerRecreateRequest NAME --pod=podName [--containers=container]
```

### Examples

```
  NOTE: The default value of CRR is:
  strategy:
  failurePolicy: Fail
  orderedRecreate: false
  unreadyGracePeriodSeconds: 3
  activeDeadlineSeconds: 300
  ttlSecondsAfterFinished: 1800
  # Create a crr with default value to restart all containers in pod-1
  kubectl kruise create ContainerRecreateRequest my-crr --namespace=ns --pod=pod-1
  
  # Create a crr with default value to restart  container-1 in pod-1
  kubectl kruise create ContainerRecreateRequest my-crr --namespace=ns --pod=pod-1 --containers=container-1
  
  # Create a crr with unreadyGracePeriodSeconds 5 and terminationGracePeriodSeconds 30 to restart  container-1 in pod-1
  kubectl kruise create ContainerRecreateRequest my-crr --namespace=ns --pod=pod-1 --containers=container-1 --unreadyGracePeriodSeconds=5 --terminationGracePeriodSeconds=30
```

### Options

```
      --allow-missing-template-keys     If true, ignore any errors in templates when a field or map key is missing in the template. Only applies to golang and jsonpath output formats. (default true)
      --containers strings              The containers those need to restarted.
      --dry-run string[="unchanged"]    Must be "none", "server", or "client". If client strategy, only print the object that would be sent, without sending it. If server strategy, submit server-side request without persisting the resource. (default "none")
      --field-manager string            Name of the manager used to track field ownership. (default "kubectl kruise-create")
  -h, --help                            help for ContainerRecreateRequest
  -m, --minStartedSeconds int32         Minimum number of seconds for which a newly created container should be started and ready without any of its container crashing, for it to be considered Succeeded.Defaults to 0 (container will be considered Succeeded as soon as it is started and ready)
  -o, --output string                   Output format. One of: (json, yaml, name, go-template, go-template-file, template, templatefile, jsonpath, jsonpath-as-json, jsonpath-file).
  -p, --pod string                      The name of the pod).
      --save-config                     If true, the configuration of current object will be saved in its annotation. Otherwise, the annotation will be unchanged. This flag is useful when you want to perform kubectl apply on this object in the future.
      --show-managed-fields             If true, keep the managedFields when printing objects in JSON or YAML format.
      --template string                 Template string or path to template file to use when -o=go-template, -o=go-template-file. The template format is golang templates [http://golang.org/pkg/text/template/#pkg-overview].
  -u, --unreadyGracePeriodSeconds int   UnreadyGracePeriodSeconds is the optional duration in seconds to mark Pod as not ready over this duration before executing preStop hook and stopping the container
      --validate string                 Must be one of: strict (or true), warn, ignore (or false).
                                        		"true" or "strict" will use a schema to validate the input and fail the request if invalid. It will perform server side validation if ServerSideFieldValidation is enabled on the api-server, but will fall back to less reliable client-side validation if not.
                                        		"warn" will warn about unknown or duplicate fields without blocking the request if server-side field validation is enabled on the API server, and behave as "ignore" otherwise.
                                        		"false" or "ignore" will not perform any schema validation, silently dropping any unknown or duplicate fields. (default "strict")
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

* [kubectl-kruise create](kubectl-kruise_create.md)	 - Create a resource from a file or from stdin.

