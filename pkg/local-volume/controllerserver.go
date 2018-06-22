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

package localvolume

import (
	"fmt"
	"strings"

	"github.com/golang/glog"
	"github.com/pborman/uuid"
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"github.com/kubernetes-csi/drivers/pkg/csi-common"
	"github.com/kubernetes-csi/drivers/pkg/local-volume/backend"
	"k8s.io/kubernetes/pkg/volume/util"
	"os"
	"google.golang.org/grpc/status"
	"google.golang.org/grpc/codes"
)

const(
	backendKey = "storage-backend"
	defaultBackend = "default"
	nodeTopologyKey = "kubernetes.io/hostname"
)

type controllerServer struct {
	backends map[string]backend.DynamicProvisioningBackend
	*csicommon.DefaultControllerServer
}

func (cs *controllerServer) CreateVolume(ctx context.Context, req *csi.CreateVolumeRequest) (*csi.CreateVolumeResponse, error) {

	// Volume ID
	volID := uuid.NewUUID().String()

	// Volume Size - Default is 1 GiB
	volSizeBytes := int64(1 * 1024 * 1024 * 1024)
	if req.GetCapacityRange() != nil {
		volSizeBytes = int64(req.GetCapacityRange().GetRequiredBytes())
	}
	volSizeGiB := util.RoundUpSize(volSizeBytes, 1024*1024*1024)

	// Volume Backend
	// Default to default backend if not specified.
	backendName := defaultBackend
	volBackend := cs.backends[defaultBackend]
	requestedName, ok := req.GetParameters()[backendKey]
	if ok {
		requestedBackend, exist := cs.backends[requestedName]
		if !exist {
			return nil, status.Error(codes.InvalidArgument, "Requested backend not exist")
		}
		backendName = requestedName
		volBackend = requestedBackend
	}

	// We need node name to set volume topology.
	nodeName := os.Getenv("NODE_NAME")
	if len(nodeName) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Node name unset in env")
	}

	volReq := &backend.LocalVolumeReq{
		VolumeID:   volID,
		SizeInGiB:  volSizeGiB,
	}

	// Volume Create
	volInfo, err := volBackend.CreateLocalVolume(volReq)
	if err != nil {
		glog.V(3).Infof("Failed to CreateVolume: %v", err)
		return nil, err
	}

	glog.V(4).Infof("Create volume % of %s type", volID, backendName)

	return &csi.CreateVolumeResponse{
		Volume: &csi.Volume{
			// Record backend as part of volume handle,
			// so that we can get it during volume deletion.
			Id: backendName + "/" + volID,
			CapacityBytes: volSizeGiB*(1024*1024*1024),
			Attributes: map[string]string{
				"device-path": volInfo.VolumePath,
			},
			AccessibleTopology: []*csi.Topology{
				{
					Segments: map[string]string{
						nodeTopologyKey: nodeName,
					},
				},
			},
		},
	}, nil
}

func (cs *controllerServer) DeleteVolume(ctx context.Context, req *csi.DeleteVolumeRequest) (*csi.DeleteVolumeResponse, error) {

	// Volume Delete
	volID := req.GetVolumeId()
	backendName, internalID, err := transformVolumeID(volID)
	if err != nil {
		glog.V(3).Infof("Failed to DeleteVolume: %v", err)
		return nil, err
	}

	backend, ok := cs.backends[backendName]
	if !ok {
		err = fmt.Errorf("requsted backend %s not exist", backendName)
		glog.V(3).Infof("Failed to DeleteVolume: %v", err)
		return nil, err
	}

	err = backend.DeleteLocalVolume(internalID)
	if err != nil {
		glog.V(3).Infof("Failed to DeleteVolume: %v", err)
		return nil, err
	}

	glog.V(4).Infof("Delete volume %s of %s type", internalID, backendName)

	return &csi.DeleteVolumeResponse{}, nil
}

func (cs *controllerServer) GetCapacity(ctx context.Context, req *csi.GetCapacityRequest) (*csi.GetCapacityResponse, error) {
	backendName, ok := req.GetParameters()[backendKey]
	if !ok {
		backendName = defaultBackend
	}
	volBackend, exist := cs.backends[backendName]
	if !exist {
		return nil, status.Error(codes.InvalidArgument, "Requested backend not exist")
	}

	capacityInBytes, err := volBackend.GetCapacity()
	if err != nil {
		return nil, err
	}

	return &csi.GetCapacityResponse{
		AvailableCapacity: capacityInBytes,
	}, nil
}

// transformVolumeID convert the recorted volume ID into storage backend name and internal ID.
func transformVolumeID(inputID string) (backend, internalID string, err error) {
	volumeInfo := strings.Split(inputID, "/")
	if len(volumeInfo) != 2 {
		return "", "", fmt.Errorf("input volume ID is not in format of 'backend/UUID' but %s", inputID)
	}

	return volumeInfo[0], volumeInfo[1], nil
}