#!/bin/bash
# Drop collections created by free5GC Stage 2 in Mongo DB

declare -a collectionlist=("NfProfile" "UriList" "subscriptionData.authenticationData.authenticationStatus" "subscriptionData.authenticationData.authenticationSubscription" "subscriptionData.contextData.amf3gppAccess" "subscriptionData.contextData.sdmSubscriptions" "subscriptionData.provisionedData.amData" "subscriptionData.provisionedData.smfSelectionSubscriptionData")

MONGO_POD=$( kubectl get pod -l app.kubernetes.io/instance=mongo -l app.kubernetes.io/name=mongo -o jsonpath='{.items[0].metadata.name}' )

for collection in "${collectionlist[@]}"
do
    kubectl exec $MONGO_POD -- mongo --eval "db.$collection.drop()" mongodb://localhost:27017/free5gc
done
