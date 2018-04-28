## Overview [![Build Status](https://travis-ci.org/wavezhang/k8s-csi-lvm.svg?branch=master)](https://travis-ci.org/wavezhang/k8s-csi-lvm)

Kubernetes LVM CSI plugin is developed to manage local storage dynamically on kubernetes.



## Deploy

1. Exec deploy/node.sh on all nodes of kubernetes.
2. On master node, exec
```bash
kubectl create -f deploy/kubernetes
```

## Usage

See ```deploy/example```

## Troubleshooting

Please submit an issue at: [Issues](https://github.com/wavezhang/k8s-csi-lvm/issues)
