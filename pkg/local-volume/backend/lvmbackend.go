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

package backend

const (
	DefaultRootPath = "/dev"
)

type LvmBackend struct {
	volumeGroup string
	rootPath    string
}

func NewLvmBackend(volumeGroup, rootPath string) *LvmBackend {
	return &LvmBackend{
		volumeGroup: volumeGroup,
		rootPath:    rootPath,
	}
}

func (w *LvmBackend) CreateLocalVolume(volReq *LocalVolumeReq) (*LocalVolumeInfo, error) {
	_, err := createLv(volReq.VolumeID, volReq.SizeInGiB, w.volumeGroup)
	if err != nil {
		return nil, err
	}
	return &LocalVolumeInfo{
		VolumeID:   volReq.VolumeID,
		SizeInGiB:  volReq.SizeInGiB,
		VolumePath: getLvPath(w.rootPath, w.volumeGroup, volReq.VolumeID),
	}, nil
}

func (w *LvmBackend) DeleteLocalVolume(volID string) error {
	_, err := deleteLv(w.rootPath, w.volumeGroup, volID)
	return err
}

func (w *LvmBackend) GetAllLocalVolumes() ([]*LocalVolumeInfo, error) {
	lvInfoList := []*LocalVolumeInfo{}
	lvNames, err := getLvNames(w.volumeGroup)
	if err != nil || len(lvNames) == 0 {
		return lvInfoList, err
	}
	for _, lvName := range lvNames {
		lvInfo, err := w.GetLocalVolume(lvName)
		if err != nil {
			return lvInfoList, err
		}
		lvInfoList = append(lvInfoList, lvInfo)
	}

	return lvInfoList, nil
}

func (w *LvmBackend) GetLocalVolume(volID string) (*LocalVolumeInfo, error) {
	lvpath := getLvPath(w.rootPath, w.volumeGroup, volID)
	size, err := getLvSize(lvpath)
	if err != nil {
		return nil, err
	}
	return &LocalVolumeInfo{
		VolumeID:   volID,
		SizeInGiB:  size,
		VolumePath: lvpath,
	}, nil
}

func (w *LvmBackend) GetCapacity() (int64, error) {
	return getVgSize(w.volumeGroup)
}
