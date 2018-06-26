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
	"golang.org/x/net/context"
)

var fakeDriverName = "LocalDriver"
var fakeBackendType = "fake"
var fakeNodeID = "LocalNodeID"
var fakeNodeName = "LocalNode"
var fakeEndpoint = "tcp://127.0.0.1:10000"
var fakeCtx = context.Background()
var fakeVolName = "LocalVolumeName"
var fakeVolID = "LocalVolumeID"
var fakeCapacity = 100*1024*1024*1024
var fakeDevicePath = "/dev/xxx"
var fakeStagingPath = "/mnt/global"
var fakeTargetPath = "/mnt/local"