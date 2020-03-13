#!/bin/bash
# Wait for pods to be terminated under the given namespace

usage() {
    echo "Wait for pods to be terminated under the given namespace"
    echo ""
    echo "Usage: ./wait_pods_terminating.sh NAMESPACE"
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
    TERMINATING_POD_NUM=$( kubectl get pods $NAMESPACE | grep -c 'Terminating' )

    if [[ "$TERMINATING_POD_NUM" -eq "0" ]]; then
        break
    fi
    
    echo "Waiting for $TERMINATING_POD_NUM pod(s) to be terminated"
    sleep 3
done

echo "No terminating pod remains"
