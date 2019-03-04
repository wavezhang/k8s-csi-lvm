#!/bin/bash

# run on master only
kubectl proxy &
nodes=`kubectl get nodes |grep -v NAME |awk '{print $1}'`
for n in $nodes; do
  ip=`kubectl describe node $n| grep InternalIP | awk '{print $2}'`
  size=`ssh $ip "vgdisplay k8s --unit=m | grep Free" | awk '{print \$7}'`
  curl -s --header "Content-Type: application/json-patch+json" \
  --request PATCH \
  --data "[{\"op\": \"add\", \"path\": \"/status/capacity/paas.com~1lvm\", \"value\": \"${size}Mi\"}]" \
  http://localhost:8001/api/v1/nodes/$n/status > /dev/null
done
ps -ef |grep "kubectl proxy"|grep -v grep | awk '{print $2}' |xargs -r kill -9

