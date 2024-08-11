## kubectl-kruise config set-credentials

Set a user entry in kubeconfig

### Synopsis

Set a user entry in kubeconfig.

 Specifying a name that already exists will merge new fields on top of existing values.

  Client-certificate flags:
  --client-certificate=certfile --client-key=keyfile
  
  Bearer token flags:
    --token=bearer_token
  
  Basic auth flags:
    --username=basic_user --password=basic_password
  
 Bearer token and basic auth are mutually exclusive.

```
kubectl-kruise config set-credentials NAME [--client-certificate=path/to/certfile] [--client-key=path/to/keyfile] [--token=bearer_token] [--username=basic_user] [--password=basic_password] [--auth-provider=provider_name] [--auth-provider-arg=key=value] [--exec-command=exec_command] [--exec-api-version=exec_api_version] [--exec-arg=arg] [--exec-env=key=value]
```

### Examples

```
  # Set only the "client-key" field on the "cluster-admin"
  # entry, without touching other values
  kubectl config set-credentials cluster-admin --client-key=~/.kube/admin.key
  
  # Set basic auth for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --username=admin --password=uXFGweU9l35qcif
  
  # Embed client certificate data in the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --client-certificate=~/.kube/admin.crt --embed-certs=true
  
  # Enable the Google Compute Platform auth provider for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --auth-provider=gcp
  
  # Enable the OpenID Connect auth provider for the "cluster-admin" entry with additional args
  kubectl config set-credentials cluster-admin --auth-provider=oidc --auth-provider-arg=client-id=foo --auth-provider-arg=client-secret=bar
  
  # Remove the "client-secret" config value for the OpenID Connect auth provider for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --auth-provider=oidc --auth-provider-arg=client-secret-
  
  # Enable new exec auth plugin for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --exec-command=/path/to/the/executable --exec-api-version=client.authentication.k8s.io/v1beta1
  
  # Define new exec auth plugin args for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --exec-arg=arg1 --exec-arg=arg2
  
  # Create or update exec auth plugin environment variables for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --exec-env=key1=val1 --exec-env=key2=val2
  
  # Remove exec auth plugin environment variables for the "cluster-admin" entry
  kubectl config set-credentials cluster-admin --exec-env=var-to-remove-
```

### Options

```
      --auth-provider string          Auth provider for the user entry in kubeconfig
      --auth-provider-arg strings     'key=value' arguments for the auth provider
      --client-certificate string     Path to client-certificate file for the user entry in kubeconfig
      --client-key string             Path to client-key file for the user entry in kubeconfig
      --embed-certs tristate[=true]   Embed client cert/key for the user entry in kubeconfig
      --exec-api-version string       API version of the exec credential plugin for the user entry in kubeconfig
      --exec-arg strings              New arguments for the exec credential plugin command for the user entry in kubeconfig
      --exec-command string           Command for the exec credential plugin for the user entry in kubeconfig
      --exec-env stringArray          'key=value' environment values for the exec credential plugin
  -h, --help                          help for set-credentials
      --password string               password for the user entry in kubeconfig
      --token string                  token for the user entry in kubeconfig
      --username string               username for the user entry in kubeconfig
```

### Options inherited from parent commands

```
      --as string                      Username to impersonate for the operation. User could be a regular user or a service account in a namespace.
      --as-group stringArray           Group to impersonate for the operation, this flag can be repeated to specify multiple groups.
      --as-uid string                  UID to impersonate for the operation.
      --cache-dir string               Default cache directory (default "${HOME}/.kube/cache")
      --certificate-authority string   Path to a cert file for the certificate authority
      --cluster string                 The name of the kubeconfig cluster to use
      --context string                 The name of the kubeconfig context to use
      --insecure-skip-tls-verify       If true, the server's certificate will not be checked for validity. This will make your HTTPS connections insecure
      --kubeconfig string              use a particular kubeconfig file
      --match-server-version           Require server version to match client version
  -n, --namespace string               If present, the namespace scope for this CLI request
      --profile string                 Name of profile to capture. One of (none|cpu|heap|goroutine|threadcreate|block|mutex) (default "none")
      --profile-output string          Name of the file to write the profile to (default "profile.pprof")
      --request-timeout string         The length of time to wait before giving up on a single server request. Non-zero values should contain a corresponding time unit (e.g. 1s, 2m, 3h). A value of zero means don't timeout requests. (default "0")
  -s, --server string                  The address and port of the Kubernetes API server
      --tls-server-name string         Server name to use for server certificate validation. If it is not provided, the hostname used to contact the server is used
      --user string                    The name of the kubeconfig user to use
      --warnings-as-errors             Treat warnings received from the server as errors and exit with a non-zero exit code
```

### SEE ALSO

* [kubectl-kruise config](kubectl-kruise_config.md)	 - Modify kubeconfig files

###### Auto generated by spf13/cobra on 11-Aug-2024
