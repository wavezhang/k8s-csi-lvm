package lvm

import (
	"errors"
	"fmt"
	"math"
	"context"
	"os/exec"
	"strings"

	"k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	utilnode "k8s.io/kubernetes/pkg/util/node"
	"k8s.io/cloud-provider/volume/helpers"
)

const (
	lvmNodeAnnKey = "lvm/node"
	NodeLabelKey  = v1.LabelHostname
	lvmdPort      = "1736"
)

// RoundOffBytes converts roundoff the size
// 1.1Mib will be round off to 2Mib same for GiB
// size less than 1MiB will be round off to 1MiB.
func RoundOffBytes(bytes int64) int64 {
	var num int64
	floatBytes := float64(bytes)
	// round off the value if its in decimal
	if floatBytes < helpers.GiB {
		num = int64(math.Ceil(floatBytes / helpers.MiB))
		num *= helpers.MiB
	} else {
		num = int64(math.Ceil(floatBytes / helpers.GiB))
		num *= helpers.GiB
	}
	return num
}

func getLVMDAddr(client kubernetes.Interface, node string) (string, error) {
	n, err := getNode(client, node)
	if err != nil {
		return "", err
	}
	ip, err := utilnode.GetNodeHostIP(n)
	if err != nil {
		return "", err
	}
	return ip.String() + ":" + lvmdPort, nil
}

func updatePV(client kubernetes.Interface, pv *v1.PersistentVolume) (*v1.PersistentVolume, error) {
	return client.CoreV1().PersistentVolumes().Update(context.TODO(), pv, metav1.UpdateOptions{})
}

func getPV(client kubernetes.Interface, volumeId string) (*v1.PersistentVolume, error) {
	return client.CoreV1().PersistentVolumes().Get(context.TODO(), volumeId, metav1.GetOptions{})
}

func getNode(client kubernetes.Interface, nodeId string) (*v1.Node, error) {
	return client.CoreV1().Nodes().Get(context.TODO(), nodeId, metav1.GetOptions{})
}

func getVolumeNode(client kubernetes.Interface, volumeId string) (string, error) {
	pv, err := getPV(client, volumeId)
	if err != nil {
		return "", err
	}
	return pv.Annotations[lvmNodeAnnKey], nil
}

func generateNodeAffinity(node *v1.Node) (*v1.VolumeNodeAffinity, error) {
	if node.Labels == nil {
		return nil, fmt.Errorf("Node does not have labels")
	}
	nodeValue, found := node.Labels[NodeLabelKey]
	if !found {
		return nil, fmt.Errorf("Node does not have expected label %s", NodeLabelKey)
	}

	return &v1.VolumeNodeAffinity{
		Required: &v1.NodeSelector{
			NodeSelectorTerms: []v1.NodeSelectorTerm{
				{
					MatchExpressions: []v1.NodeSelectorRequirement{
						{
							Key:      NodeLabelKey,
							Operator: v1.NodeSelectorOpIn,
							Values:   []string{nodeValue},
						},
					},
				},
			},
		},
	}, nil
}

func formatDevice(devicePath, fstype string) error {
	output, err := exec.Command("mkfs", "-t", fstype, devicePath).CombinedOutput()
	if err != nil {
		return errors.New("csi-lvm: formatDevice: " + string(output))
	}
	return nil
}

func determineFilesystemType(devicePath string) (string, error) {
	// We use `file -bsL` to determine whether any filesystem type is detected.
	// If a filesystem is detected (ie., the output is not "data", we use
	// `blkid` to determine what the filesystem is. We use `blkid` as `file`
	// has inconvenient output.
	// We do *not* use `lsblk` as that requires udev to be up-to-date which
	// is often not the case when a device is erased using `dd`.
	output, err := exec.Command("file", "-bsL", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(output)) == "data" {
		// No filesystem detected.
		return "", nil
	}
	// Some filesystem was detected, we use blkid to figure out what it is.
	output, err = exec.Command("blkid", "-c", "/dev/null", "-o", "export", devicePath).CombinedOutput()
	if err != nil {
		return "", err
	}
	parseErr := errors.New("Cannot parse output of blkid.")
	lines := strings.Split(string(output), "\n")
	for _, line := range lines {
		fields := strings.Split(strings.TrimSpace(line), "=")
		if len(fields) != 2 {
			return "", parseErr
		}
		if fields[0] == "TYPE" {
			return fields[1], nil
		}
	}
	return "", parseErr
}
