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
	"os"

	"github.com/golang/glog"
	"golang.org/x/net/context"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"k8s.io/kubernetes/pkg/util/mount"
	volumeutil "k8s.io/kubernetes/pkg/volume/util"

	"github.com/kubernetes-csi/drivers/pkg/csi-common"
)

type nodeServer struct {
	mounter *mount.SafeFormatAndMount
	*csicommon.DefaultNodeServer
}

func (ns *nodeServer) NodePublishVolume(ctx context.Context, req *csi.NodePublishVolumeRequest) (*csi.NodePublishVolumeResponse, error) {

	// Check arguments
	if req.GetVolumeCapability() == nil || req.GetVolumeCapability().GetMount() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}

	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Staging Target path missing in request")
	}

	targetPath := req.GetTargetPath()
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(targetPath, 0750); err != nil {
				return nil, status.Error(codes.Internal, err.Error())
			}
			notMnt = true
		} else {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	if !notMnt {
		return &csi.NodePublishVolumeResponse{}, nil
	}

	readOnly := req.GetReadonly()
	volumeID := req.GetVolumeId()
	mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()


	options := []string{"bind"}
	options = append(options, mountFlags...)
	if readOnly {
		options = append(options, "ro")
	}
	if err := ns.mounter.Mount(req.GetStagingTargetPath(), targetPath, "", options); err != nil {
		return nil, err
	}
	glog.V(4).Infof("volume %s has been mounted to %s.", volumeID, targetPath)

	return &csi.NodePublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnpublishVolume(ctx context.Context, req *csi.NodeUnpublishVolumeRequest) (*csi.NodeUnpublishVolumeResponse, error) {

	// Check arguments
	if len(req.GetTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	// Unmounting the image
	err := ns.mounter.Unmount(req.GetTargetPath())
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	volumeID := req.GetVolumeId()
	glog.V(4).Infof("volume %s has been unmounted.", volumeID)

	return &csi.NodeUnpublishVolumeResponse{}, nil
}

func (ns *nodeServer) NodeStageVolume(ctx context.Context, req *csi.NodeStageVolumeRequest) (*csi.NodeStageVolumeResponse, error) {

	// Check arguments
	if req.GetVolumeCapability() == nil || req.GetVolumeCapability().GetMount() == nil {
		return nil, status.Error(codes.InvalidArgument, "Volume capability missing in request")
	}

	if len(req.GetVolumeAttributes()["device-path"]) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Device path missing in request")
	}

	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	targetPath := req.GetStagingTargetPath()
	fsType := req.GetVolumeCapability().GetMount().GetFsType()
	accessMode := req.GetVolumeCapability().GetAccessMode()
	devicePath := req.GetVolumeAttributes()["device-path"]

	// Verify whether mounted
	notMnt, err := ns.mounter.IsLikelyNotMountPoint(targetPath)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	// Volume Mount
	if notMnt {
		// Get Options
		var options []string
		if accessMode.GetMode() == csi.VolumeCapability_AccessMode_MULTI_NODE_READER_ONLY ||
			accessMode.GetMode() == csi.VolumeCapability_AccessMode_SINGLE_NODE_READER_ONLY {
			options = append(options, "ro")
		} else {
			options = append(options, "rw")
		}
		mountFlags := req.GetVolumeCapability().GetMount().GetMountFlags()
		options = append(options, mountFlags...)

		// Mount
		err = ns.mounter.FormatAndMount(devicePath, targetPath, fsType, options)
		if err != nil {
			return nil, status.Error(codes.Internal, err.Error())
		}
	}
	volumeID := req.GetVolumeId()
	glog.V(4).Infof("volume %s has been staged to %s.", volumeID, targetPath)

	return &csi.NodeStageVolumeResponse{}, nil
}

func (ns *nodeServer) NodeUnstageVolume(ctx context.Context, req *csi.NodeUnstageVolumeRequest) (*csi.NodeUnstageVolumeResponse, error) {

	// Check arguments
	if len(req.GetStagingTargetPath()) == 0 {
		return nil, status.Error(codes.InvalidArgument, "Target path missing in request")
	}

	if err := volumeutil.UnmountPath(req.GetStagingTargetPath(), ns.mounter); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}
	volumeID := req.GetVolumeId()
	glog.V(4).Infof("volume %s has been unstaged.", volumeID)

	return &csi.NodeUnstageVolumeResponse{}, nil
}
