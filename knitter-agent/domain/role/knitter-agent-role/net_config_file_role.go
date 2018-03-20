/*
Copyright 2018 ZTE Corporation. All rights reserved.
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

package knitteragentrole

import (
	"encoding/json"
	"github.com/ZTE/Knitter/knitter-agent/domain/bind"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/pod-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/klog"
	"os"
)

type NetConfigFileRole struct {
}

func (this NetConfigFileRole) Create(podObj *podobj.PodObj, nics []bind.Dpdknic) error {
	nicDevs := bind.Nicdevs{}
	nicDevs.Dpdknics = nics
	agtCtx := cni.GetGlobalContext()
	nicDevs.MTU = agtCtx.Mtu
	var busInfoPath string
	busInfoPath = "/var/lib/kubelet/pods/" + podObj.PodID + "/volumes/kubernetes.io~empty-dir/config"
	_, err := os.Stat(busInfoPath)
	if err != nil {
		busInfoPath = "/var/lib/origin/openshift.local.volumes/pods/" + podObj.PodID + "/volumes/kubernetes.io~empty-dir/config"
		_, err := os.Stat(busInfoPath)
		if err != nil {
			klog.Errorf("cmdAdd:path of businfo error! %v", err)
			return infra.ErrOsStatFailed
		}
	}
	info, err := json.Marshal(nicDevs)
	if err != nil {
		klog.Errorf("cmdAdd:nic-devs 2 json error! -%v", err)
		return infra.ErrJSONMarshalFailed
	}
	file := busInfoPath + "/net-config.json"
	f, err := os.Create(file)
	if err != nil {
		klog.Errorf("cmdAdd:create net-config.json error! %v", err)
		return infra.ErrJSONMarshalFailed
	}
	defer f.Close()
	f.WriteString(string(info))
	klog.Infof("cmdAdd success!")
	return nil
}
