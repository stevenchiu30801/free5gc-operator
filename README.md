# free5GC Operator

A Kubernetes operator for deploy and manage free5GC network slices in BANS 5GC

## Run

```ShellSession
# Register Free5GCSlice CRD with the Kubernetes apiserver
kubectl create -f deploy/crds/bans.io_free5gcslice_crd.yaml

# Setup RBAC
kubectl create -f deploy/service_account.yaml
kubectl create -f deploy/role.yaml
kubectl create -f deploy/role_binding.yaml
kubectl create -f deploy/cluster_role.yaml
kubectl create -f deploy/cluster_role_binding.yaml

# Deploy free5GC Operator
kubectl create -f deploy/operator.yaml

# Create example CR
kubectl create -f deploy/crds/bans.io_v1alpha1_free5gcslice_cr.yaml
```
