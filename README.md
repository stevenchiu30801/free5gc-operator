# free5GC Operator

A Kubernetes operator for deploying and managing free5GC network slices in BANS 5GC.

## Prerequisites

See [operator-framework/operator-sdk](https://github.com/operator-framework/operator-sdk#prerequisites).

```Shellsession
# Pre-install
sudo apt instal make
```

## Usage

### Run

```ShellSession
# Install all resources (CRD's, RBAC and Operator)
make install
```

### Procedure Test

```ShellSession
# Create a new CR
kubectl apply -f deploy/crds/bans.io_v1alpha1_free5gcslice_cr1.yaml

# Check if the new slice is running before proceeding
kubectl get pods -l app.kubernetes.io/name=free5gc-smf,bans.io/slice=slice1 | grep Running

# Set ransim pod variable
export RANSIM_POD=$( kubectl get pods -l app.kubernetes.io/instance=free5gc -l app.kubernetes.io/name=ransim -o jsonpath='{.items[0].metadata.name}' )

# Test registration and data traffic with slice 1
kubectl exec $RANSIM_POD -- bash -c "cd src/test && go test -vet=off -run TestRegistration -ue-idx=1 -sst=1 -sd=010203"

# Create a new CR
kubectl apply -f deploy/crds/bans.io_v1alpha1_free5gcslice_cr2.yaml

# Check if the new slice is running before proceeding
kubectl get pods -l app.kubernetes.io/name=free5gc-smf,bans.io/slice=slice2 | grep Running

# Test registration and data traffic with slice 2
kubectl exec $RANSIM_POD -- bash -c "cd src/test && go test -vet=off -run TestRegistration -ue-idx=2 -sst=1 -sd=112233"
```

### Reset

```ShellSession
# Uninstall all that all performed in the $ make install
make uninstall

# Execute the following scripts if uninstallation blocks at removing Free5GCSlice CRD
./scripts/remove_finalizers.sh

# Uninstall all BANS 5GC functions along with CR except Mongo DB
make reset-free5gc
```
