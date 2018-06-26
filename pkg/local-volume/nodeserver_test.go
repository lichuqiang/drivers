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
	"testing"

	"k8s.io/kubernetes/pkg/util/mount"
	"github.com/container-storage-interface/spec/lib/go/csi/v0"
	"reflect"
)

var fakeNs *nodeServer

// Init Node Server
func init() {
	if fakeNs == nil {
		fakeNs = &nodeServer{
			mounter: &mount.SafeFormatAndMount{Interface: &mount.FakeMounter{}, Exec: mount.NewFakeExec(nil)},
		}
	}
}

// TestNodePublishVolume tests behavior of NodePublishVolume
func TestNodePublishVolume(t *testing.T) {
	// Fake request
	fakeReq := &csi.NodePublishVolumeRequest{
		VolumeId:         fakeVolID,
		TargetPath:       fakeTargetPath,
		StagingTargetPath: fakeStagingPath,
		Readonly:         false,
		VolumeCapability: &csi.VolumeCapability{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{
					MountFlags: []string{},
				},
			},
		},
	}

	// Expected Result
	expectedRes := &csi.NodePublishVolumeResponse{}

	// Invoke NodePublishVolume
	actualRes, err := fakeNs.NodePublishVolume(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to NodePublishVolume: %v", err)
	}

	if !reflect.DeepEqual(expectedRes, actualRes) {
		t.Errorf("Expected response: %v, actual response: %v", expectedRes, actualRes)
	}
}

// TestNodeUnpublishVolume tests behavior of NodeUnpublishVolume
func TestNodeUnpublishVolume(t *testing.T) {
	// Fake request
	fakeReq := &csi.NodeUnpublishVolumeRequest{
		VolumeId:   fakeVolID,
		TargetPath: fakeTargetPath,
	}

	// Expected Result
	expectedRes := &csi.NodeUnpublishVolumeResponse{}

	// Invoke NodeUnpublishVolume
	actualRes, err := fakeNs.NodeUnpublishVolume(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to NodeUnpublishVolume: %v", err)
	}

	if !reflect.DeepEqual(expectedRes, actualRes) {
		t.Errorf("Expected response: %v, actual response: %v", expectedRes, actualRes)
	}
}

// TestNodeStageVolume tests behavior of NodeStageVolume
func TestNodeStageVolume(t *testing.T) {
	// Fake request
	fakeReq := &csi.NodeStageVolumeRequest{
		VolumeId:         fakeVolID,
		StagingTargetPath: fakeStagingPath,
		VolumeCapability: &csi.VolumeCapability{
			AccessType: &csi.VolumeCapability_Mount{
				Mount: &csi.VolumeCapability_MountVolume{
					FsType: "ext4",
				},
			},
		},
		VolumeAttributes: map[string]string{"device-path": fakeDevicePath},
	}

	// Expected Result
	expectedRes := &csi.NodeStageVolumeResponse{}

	// Invoke NodeStageVolume
	actualRes, err := fakeNs.NodeStageVolume(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to NodeStageVolume: %v", err)
	}

	if !reflect.DeepEqual(expectedRes, actualRes) {
		t.Errorf("Expected response: %v, actual response: %v", expectedRes, actualRes)
	}
}

// TestNodeUnstageVolume tests behavior of NodeUnstageVolume
func TestNodeUnstageVolume(t *testing.T) {
	// Fake request
	fakeReq := &csi.NodeUnstageVolumeRequest{
		VolumeId:   fakeVolID,
		StagingTargetPath: fakeTargetPath,
	}

	// Expected Result
	expectedRes := &csi.NodeUnstageVolumeResponse{}

	// Invoke NodeUnstageVolume
	actualRes, err := fakeNs.NodeUnstageVolume(fakeCtx, fakeReq)
	if err != nil {
		t.Errorf("failed to NodeUnstageVolume: %v", err)
	}

	if !reflect.DeepEqual(expectedRes, actualRes) {
		t.Errorf("Expected response: %v, actual response: %v", expectedRes, actualRes)
	}
}