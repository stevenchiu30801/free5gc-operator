#!/bin/bash
# Remove finalizers in Free5GCSlice CRs

CRS=$( kubectl get free5gcslice.bans.io -o jsonpath='{.items[?(@.metadata.finalizers)].metadata.name}' )
IFS=' ' read -r -a cr_list <<< "$CRS"

for cr in "${cr_list[@]}"
do
    kubectl get free5gcslice.bans.io $cr -o=json | jq '.metadata.finalizers = null' | kubectl apply -f -
done
