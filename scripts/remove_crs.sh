#!/bin/bash
# Remove all Free5GCSlice CRs

CRS=$( kubectl get free5gcslice.bans.io -o jsonpath='{.items[*].metadata.name}' )
IFS=' ' read -r -a cr_list <<< "$CRS"

for cr in "${cr_list}"
do
    kubectl delete free5gcslice.bans.io $cr
done
