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
	"reflect"
	"testing"

	"github.com/container-storage-interface/spec/lib/go/csi/v0"
)

var fakeCs *controllerServer

// Init Controller Server
func init() {
	if fakeCs == nil {
		d := NewDriver(fakeDriverName, fakeBackendType, fakeNodeID, fakeEndpoint, []string{})
		fakeCs = d.cs
	}
}

// TestCreateVolume tests behavior of CreateVolume
func TestCreateVolume(t *testing.T) {

	scenarios := map[string]struct {
		// Inputs
		volSize      int64
		backendName  string
		setNodeName  bool
		shouldFail   bool
		expectedRes *csi.CreateVolumeResponse
	}{
		"valid-creation": {
			volSize:    10*1024*1024*1024,
			setNodeName:     true,
			expectedRes: &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					Id: fakeVolID,
					CapacityBytes: 10*1024*1024*1024,
					Attributes: map[string]string{
						"device-path": fakeDevicePath,
					},
					AccessibleTopology: []*csi.Topology{
						{
							Segments: map[string]string{
								nodeTopologyKey: fakeNodeName,
							},
						},
					},
				},
			},
		},
		"default-size": {
			setNodeName:     true,
			expectedRes: &csi.CreateVolumeResponse{
				Volume: &csi.Volume{
					Id: fakeVolID,
					// Should default to 1 GiB if unset
					CapacityBytes: 1*1024*1024*1024,
					Attributes: map[string]string{
						"device-path": fakeDevicePath,
					},
					AccessibleTopology: []*csi.Topology{
						{
							Segments: map[string]string{
								nodeTopologyKey: fakeNodeName,
							},
						},
					},
				},
			},
		},
		"no-nodename": {
			volSize:    10*1024*1024*1024,
			shouldFail: true,
		},
		"invalid-backendname": {
			volSize:    10*1024*1024*1024,
			backendName:"invalid",
			setNodeName:     true,
			shouldFail: true,
		},
	}
	for name, scenario := range scenarios {
		req := &csi.CreateVolumeRequest{
			Name: fakeVolName,
			Parameters: map[string]string{},
		}
		if scenario.volSize != 0 {
			req.CapacityRange = &csi.CapacityRange{
				RequiredBytes: scenario.volSize,
			}
		}
		if scenario.backendName != "" {
			req.Parameters[backendKey] = scenario.backendName
		}
		if scenario.setNodeName {
			os.Setenv("NODE_NAME", fakeNodeName)
		} else {
			os.Unsetenv("NODE_NAME")
		}

		// Invoke CreateVolume
		actualRes, err := fakeCs.CreateVolume(fakeCtx, req)
		// Validate
		if scenario.shouldFail && err == nil {
			t.Errorf("Test %q failed: returned success but expected error", name)
		}
		if !scenario.shouldFail {
			if err != nil {
				t.Errorf("Test %q failed: returned error: %v", name, err)
			}
			actualRes.Volume.Id = fakeVolID
			if !reflect.DeepEqual(scenario.expectedRes, actualRes) {
				t.Errorf("Test %q failed: expected response: %v, actual response: %v", name, scenario.expectedRes, actualRes)
			}

		}
	}

}
