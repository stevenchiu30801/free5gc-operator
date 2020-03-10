# free5GC Operator

A Kubernetes operator for deploying and managing free5GC network slices in BANS 5GC

## Pre-install

```Shellsession
sudo apt instal make
```

## Run

```ShellSession
# Install all resources (CR/CRD's, RBAC and Operator)
make install
```

## Procedure Test

```ShellSession
export RANSIM_POD=$( kubectl get pod -l app.kubernetes.io/instance=free5gc -l app.kubernetes.io/name=ransim -o jsonpath='{.items[0].metadata.name}' )

# Test registration and data traffic with slice 1
kubectl exec $RANSIM_POD -- bash -c "cd src/test && go test -vet=off -run TestRegistration -ue-idx=1 -sst=1 -sd=010203"

# Create a new CR
kubectl apply -f deploy/crds/bans.io_v1alpha1_free5gcslice_cr2.yaml

# Check if the new slice is running before proceeding
kubectl get pods -l app.kubernetes.io/name=free5gc-smf,bans.io/slice=slice2 | grep Running

# Test registration and data traffic with slice 2
kubectl exec $RANSIM_POD -- bash -c "cd src/test && go test -vet=off -run TestRegistration -ue-idx=2 -sst=1 -sd=112233"
```

## Reset

```ShellSession
# Uninstall all that all performed in the $ make install
make uninstall

# Uninstall all BANS 5GC functions along with CR except Mongo DB
make reset-free5gc
```
