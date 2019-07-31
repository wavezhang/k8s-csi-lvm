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
	"github.com/container-storage-interface/spec/lib/go/csi"
	"github.com/golang/glog"
	csicommon "github.com/kubernetes-csi/drivers/pkg/csi-common"
	"k8s.io/client-go/kubernetes"
)

type lvm struct {
	driver *csicommon.CSIDriver
	client kubernetes.Interface

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer

	cap   []*csi.VolumeCapability_AccessMode
	cscap []*csi.ControllerServiceCapability
}

var (
	lvmDriver     *lvm
	vendorVersion = "0.3.0"
)

func GetLVMDriver(client kubernetes.Interface) *lvm {
	return &lvm{client: client}
}

func NewIdentityServer(d *csicommon.CSIDriver) *identityServer {
	return &identityServer{
		DefaultIdentityServer: csicommon.NewDefaultIdentityServer(d),
	}
}

func NewControllerServer(d *csicommon.CSIDriver, c kubernetes.Interface, vgName string) *controllerServer {
	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
		client:                  c,
		vgName:                  vgName,
	}
}

func NewNodeServer(d *csicommon.CSIDriver, c kubernetes.Interface, nodeID string, vgName string) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
		client:            c,
		nodeID:            nodeID,
		vgName:            vgName,
	}
}

func (lvm *lvm) Run(driverName, nodeID, endpoint string, vgName string) {
	glog.Infof("Driver: %v ", driverName)

	// Initialize default library driver
	lvm.driver = csicommon.NewCSIDriver(driverName, vendorVersion, nodeID)

	if lvm.driver == nil {
		glog.Fatalln("Failed to initialize CSI Driver.")
	}
	lvm.driver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	lvm.driver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	// Create GRPC servers
	lvm.ids = NewIdentityServer(lvm.driver)
	lvm.ns = NewNodeServer(lvm.driver, lvm.client, nodeID, vgName)
	lvm.cs = NewControllerServer(lvm.driver, lvm.client, vgName)

	server := csicommon.NewNonBlockingGRPCServer()
	server.Start(endpoint, lvm.ids, lvm.cs, lvm.ns)
	server.Wait()
}
