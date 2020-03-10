#!/bin/bash
# Remove all free5GC slices (SMFs and UPFs)

# Delete SMFs
SMF_DEPLOYS=$( kubectl get deployments.apps -l app.kubernetes.io/name=free5gc-smf -o jsonpath='{.items[*].metadata.name}' )
IFS=' ' read -r -a smf_deploy_list <<< "$SMF_DEPLOYS"

for smf_deploy in "${smf_deploy_list[@]}"
do
    helm uninstall $smf_deploy
done

# Delete UPFs
UPF_DEPLOYS=$( kubectl get deployments.apps -l app.kubernetes.io/name=free5gc-upf -o jsonpath='{.items[*].metadata.name}' )
IFS=' ' read -r -a upf_deploy_list <<< "$UPF_DEPLOYS"

for upf_deploy in "${upf_deploy_list[@]}"
do
    helm uninstall $upf_deploy
done
