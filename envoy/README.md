# Non cascading deletion
Delete owner objects and orphan dependents
By default, when you tell Kubernetes to delete an object, the controller also deletes dependent objects. You can make Kubernetes orphan these dependents using kubectl or the Kubernetes API, depending on the Kubernetes version your cluster runs. To check the version, enter kubectl version.
Using kubectl

Run the following command:
```
kubectl delete deployment envoy-deployment --cascade=orphan
```
Using the Kubernetes API

Start a local proxy session:

kubectl proxy --port=8080
Use curl to trigger deletion:
```
curl -X DELETE localhost:8080/apis/apps/v1/namespaces/default/deployments/envoy-deployment \
    -d '{"kind":"DeleteOptions","apiVersion":"v1","propagationPolicy":"Orphan"}' \
    -H "Content-Type: application/json"
```
The output contains orphan in the finalizers field, similar to this:
```
"kind": "Deployment",
"apiVersion": "apps/v1",
"namespace": "default",
"uid": "6f577034-42a0-479d-be21-78018c466f1f",
"creationTimestamp": "2021-07-09T16:46:37Z",
"deletionTimestamp": "2021-07-09T16:47:08Z",
"deletionGracePeriodSeconds": 0,
"finalizers": [
  "orphan"
],
...
```
You can check that the Pods managed by the Deployment are still running:
```
kubectl get pods -l app=envoy
```
Reference: https://kubernetes.io/docs/tasks/administer-cluster/use-cascading-deletion/