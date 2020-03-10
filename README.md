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

## Reset

```ShellSession
# Uninstall all that all performed in the $ make install
make uninstall

# Uninstall all BANS 5GC functions along with CR except Mongo DB
make reset-free5gc
```
