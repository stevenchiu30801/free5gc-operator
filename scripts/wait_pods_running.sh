#!/bin/bash
# Wait for pods to be running under the given namespace

usage() {
    echo "Wait for pods to be running under the given namespace"
    echo ""
    echo "Usage: ./wait_pods_running.sh NAMESPACE"
    echo "Arguments:"
    echo "    NAMESPACE         Namespace of pods to be waited"
}

NAMESPACE=""

if [[ $# -eq 1 ]]; then
    if [[ "$1" == "all" ]]; then
        NAMESPACE="--all-namespaces"
    else
        NAMESPACE="-n $1"
    fi
else
    usage
    exit 1
fi

while true; do
    PENDING_PODS=$( kubectl get pods $NAMESPACE -o jsonpath='{.items[?(@.status.phase=="Pending")].metadata.name}')

    if [[ -z $PENDING_PODS ]]; then
        break
    fi
    
    IFS=' ' read -r -a pending_pod_list <<< "$PENDING_PODS"
    PENDING_POD_NUM=${#pending_pod_list[@]}

    echo "Waiting for $PENDING_POD_NUM pod(s) to be running"
    sleep 3
done

echo "No pending pod remains"
