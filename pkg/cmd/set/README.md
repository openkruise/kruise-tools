# Note: The set package is still in development progress
Update cloneset 'sample' with a new environment variable
* [x] kubectl-kruise set env cloneset/sample STORAGE_DIR=/local

List the environment variables defined on a deployments 'sample-build'
* [x] kubectl-kruise set env cloneset/sample --list

List the environment variables defined on all pods
* [x] kubectl-kruise set env pods --all --list

Output modified deployment in YAML, and does not alter the object on the server
* [x] kubectl-kruise set env cloneset/sample STORAGE_DIR=/data -o yaml

Update all containers in all replication controllers in the project to have ENV=prod
* [x] kubectl-kruise set env rc --all ENV=prod

Import environment from a secret
* [x] kubectl-kruise set env --from=secret/mysecret deployment/myapp

Import environment from a config map with a prefix
* [x] kubectl-kruise set env --from=configmap/myconfigmap --prefix=MYSQL_ deployment/myapp

Import specific keys from a config map
* [x] kubectl-kruise set env --keys=my-example-key --from=configmap/myconfigmap deployment/myapp

Remove the environment variable ENV from container 'c1' in all deployment configs
* [x] kubectl-kruise set env deployments --all --containers="c1" ENV-

Remove the environment variable ENV from a deployment definition on disk and update the deployment config on the server
* [x] kubectl-kruise set env -f deploy.json ENV-

Set some of the local shell environment into a deployment config on the server
* [x] env | grep XDG_VTNR | kubectl-kruise set env -e - deployment/nginx-deployment

Set a deployment's nginx container image to 'nginx:1.9.1', and its busybox container image to 'busybox'.
* [x] kubectl-kruise set image cloneset/sample busybox=busybox nginx=nginx:1.9.1

Update all deployments' and rc's nginx container's image to 'nginx:1.9.1'
* [x] kubectl-kruise set image cloneset,rc nginx=nginx:1.9.1 --all

Update image of all containers of daemonset abc to 'nginx:1.9.1'
* [x] kubectl-kruise set image cloneset sample *=nginx:1.9.1

Print result (in yaml format) of updating nginx container image from local file, without hitting the server
* [x] kubectl-kruise set image -f path/to/file.yaml nginx=nginx:1.9.1 --local -o yaml

Set a deployments nginx container cpu limits to "200m" and memory to "512Mi"
* [x] kubectl-kruise set resources cloneset sample -c=nginx --limits=cpu=200m,memory=512Mi

Set the resource request and limits for all containers in nginx
* [x] kubectl-kruise set resources cloneset sample --limits=cpu=200m,memory=512Mi --requests=cpu=100m,memory=256Mi

Remove the resource requests for resources on containers in nginx
* [x] kubectl-kruise set resources cloneset sample --limits=cpu=0,memory=0 --requests=cpu=0,memory=0

Print the result (in yaml format) of updating nginx container limits from a local, without hitting the server
* [x] kubectl-kruise set resources -f path/to/file.yaml --limits=cpu=200m,memory=512Mi --local -o yaml

Set Deployment nginx-deployment's ServiceAccount to serviceaccount1
* [x] kubectl-kruise set serviceaccount cloneset sample serviceaccount1

Print the result (in yaml format) of updated nginx deployment with serviceaccount from local file, without hitting apiserver
* [x] kubectl-kruise set sa -f nginx-deployment.yaml serviceaccount1 --local --dry-run=client -o yaml

Update a ClusterRoleBinding for serviceaccount1
* [x] kubectl set subject clusterrolebinding admin --serviceaccount=namespace:serviceaccount1

Update a RoleBinding for user1, user2, and group1
* [x] kubectl set subject rolebinding admin --user=user1 --user=user2 --group=group1

Print the result (in yaml format) of updating rolebinding subjects from a local, without hitting the server
* [x] kubectl rolebinding admin --role=admin --user=admin -o yaml --dry-run=client | kubectl set subject --local -f - --user=foo -o yaml

