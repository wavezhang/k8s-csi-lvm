## Overview [![Build Status](https://travis-ci.org/wavezhang/k8s-csi-lvm.svg?branch=master)](https://travis-ci.org/wavezhang/k8s-csi-lvm)

Kubernetes LVM CSI plugin is developed to manage local storage dynamically on kubernetes.



## Deploy

1. kube-apiserver must be launched with ```--feature-gates=CSIPersistentVolume=true,MountPropagation=true``` and ```--runtime-config=storage.k8s.io/v1alpha1=true```
2. Exec ```deploy/node.sh``` on all nodes of kubernetes.
3. On master node, exec
```bash
kubectl create -f deploy/kubernetes
```
4. If you need aware node lvm capacity when schedule, on master node, exec ```deploy/capacity.sh``` and when using lvm in pod add  requests like following:
```yaml
    resources:
      limits:
        paas.com/lvm: 1Gi
```      


## Usage

See ```deploy/example```

## Troubleshooting

Please submit an issue at: [Issues](https://github.com/wavezhang/k8s-csi-lvm/issues)
