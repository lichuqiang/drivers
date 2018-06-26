/*
Copyright 2018 The Kubernetes Authors.

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

package localvolume

import (
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/golang/glog"

	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/kubernetes-csi/drivers/pkg/local-volume/backend"
	"k8s.io/kubernetes/pkg/util/mount"
)

var (
	version  = "0.2.0"
)

type driver struct {
	csiDriver *csicommon.CSIDriver
	endpoint  string

	ids *identityServer
	ns  *nodeServer
	cs  *controllerServer
}

func NewControllerServer(d *csicommon.CSIDriver, backendType string, volumeGroups []string) *controllerServer {
	backends := map[string]backend.DynamicProvisioningBackend{}

	switch backend.BackendType(backendType) {
	case backend.Lvm:
		for i, group := range volumeGroups {
			volBackend := backend.NewLvmBackend(group, backend.DefaultRootPath)
			backends[group] = volBackend
			if i == 0 {
				// Add the first backend as default one.
				backends["default"] = volBackend
			}
		}
	case backend.Fake:
		// Used for test
		backends["default"] = backend.NewFakeBackend()
	default:
		glog.Fatalf("Unrecognized backend type %s", backendType)
	}

	return &controllerServer{
		DefaultControllerServer: csicommon.NewDefaultControllerServer(d),
		backends: backends,
	}
}

func NewNodeServer(d *csicommon.CSIDriver) *nodeServer {
	return &nodeServer{
		DefaultNodeServer: csicommon.NewDefaultNodeServer(d),
		mounter: &mount.SafeFormatAndMount{Interface: mount.New(""), Exec: mount.NewOsExec()},
	}
}

func NewDriver(driverName, backendType, nodeID, endpoint string, volumeGroups []string) *driver {
	glog.Infof("Driver: %v ", driverName)

	d := &driver{}
	d.endpoint = endpoint

	// Initialize default library driver
	csiDriver := csicommon.NewCSIDriver(driverName, version, nodeID)
	if csiDriver == nil {
		glog.Fatalln("Failed to initialize CSI Driver.")
	}
	csiDriver.AddControllerServiceCapabilities([]csi.ControllerServiceCapability_RPC_Type{csi.ControllerServiceCapability_RPC_CREATE_DELETE_VOLUME})
	csiDriver.AddVolumeCapabilityAccessModes([]csi.VolumeCapability_AccessMode_Mode{csi.VolumeCapability_AccessMode_SINGLE_NODE_WRITER})

	// Create GRPC servers
	d.ns = NewNodeServer(csiDriver)
	d.cs = NewControllerServer(csiDriver, backendType, volumeGroups)

	d.csiDriver = csiDriver

	return d
}

func (d *driver) Run() {
	csicommon.RunControllerandNodePublishServer(d.endpoint, d.csiDriver, d.cs, d.ns)
}
