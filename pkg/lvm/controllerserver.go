/*
Copyright 2017 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package lvm

import (
	"fmt"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/client-go/kubernetes"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/wavezhang/k8s-csi-lvm/pkg/lvmd"
)

const (
	defaultFs      = "ext4"
	connectTimeout = 3 * time.Second
)

type controllerServer struct {
	*csicommon.DefaultControllerServer
	client kubernetes.Interface
	vgName string
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {
	if err := cs.Driver.ValidateControllerServiceRequest(csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME); err != nil {
		glog.V(3).Infof("invalid create volume req: %v", req)
		return nil, err
	}
	// Check sanity of request Name, Volume Capabilities
	if len(req.Name) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Volume Name cannot be empty")
	}
	if req.VolumeCapabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume Capabilities cannot be empty")
	}

	volumeId := req.GetName()

	response := &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			Id:            volumeId,
			CapacityBytes: req.GetCapacityRange().GetRequiredBytes(),
			Attributes:    req.GetParameters(),
		},
	}
	return response, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {
	vid := req.GetVolumeId()
	node, err := getVolumeNode(cs.client, vid)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Failed to getVolumeNode for %v: %v", vid, err))
	}
	if node != "" {
		addr, err := getLVMDAddr(cs.client, node)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, fmt.Sprintf("Failed to getLVMDAddr for %v: %v", node, err))
		}

		conn, err := lvmd.NewLVMConnection(addr, connectTimeout)
		defer conn.Close()
		if err != nil {
			return nil, status.Error(codes.Internal, fmt.Sprintf("Failed to connect to %v: %v", addr, err))
		}

		if _, err := conn.GetLV(ctx, cs.vgName, vid); err == nil {
			if err := conn.RemoveLV(ctx, cs.vgName, vid); err != nil {
				return nil, status.Errorf(
					codes.Internal,
					"Failed to remove volume: err=%v",
					err)
			}
		}
	}
	response := &csi.DeleteVolumeResponse{}
	return response, nil
}
