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

package podobj

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/monitor"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/pod-role"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/inter-cmpt/agt-mgr"
	"github.com/ZTE/Knitter/pkg/klog"
)

type PodObj struct {
	PortRole podrole.PortRole
	podrole.PodDataRole
	BuildPodRole podrole.PodBuilderRole

	MangerClient manager.ManagerClient
	PortObjs     []*portobj.PortObj
}

func CreatePodObj(cniParam *cni.CniParam, pod *monitor.Pod) (*PodObj, error) {
	podObj := &PodObj{}
	podObj.PodName = cniParam.PodName
	podObj.PodNs = cniParam.PodNs
	podObj.TenantID = cniParam.TenantID
	podObj.HostType = cniParam.HostType
	podObj.VMID = cniParam.VMID
	podObj.VnicType = cniParam.VnicType
	err := podObj.BuildPodRole.Transform(cniParam.PodNs, cniParam.PodName, pod)
	if err != nil {
		return nil, err
	}
	podObj.PortRole.Init(&podObj.PodDataRole)
	podObj.PodDataRole.PodID = podObj.BuildPodRole.PodID
	podObj.PodDataRole.PodType = podObj.BuildPodRole.PodType
	podObj.PortObjs = podObj.BuildPodRole.PortObjs
	return podObj, nil
}

func (self *PodObj) TransformAgtPodReq() (*agtmgr.AgentPodReq, error) {
	klog.Infof("PodObj.TransformAgtPodReq: TRANSFORM START, PortObj is :[%v] ", self)
	ports := make([]agtmgr.PortInfo, 0)
	for index, port := range self.PortObjs {
		portInfo := agtmgr.PortInfo{
			PortId:       port.LazyAttr.ID,
			NetworkName:  port.EagerAttr.NetworkName,
			NetworkPlane: port.EagerAttr.NetworkPlane,
		}
		if len(port.LazyAttr.FixedIps) < 1 {
			klog.Errorf("PodObj.TransformAgtPodReq: len(port.LazyAttr.FixedIps) "+
				"is %v, error is %v", len(port.LazyAttr.FixedIps), errobj.ErrFixIpsIsNil)
			return nil, errobj.ErrFixIpsIsNil
		}
		portInfo.FixIP = port.LazyAttr.FixedIps[0].IPAddress
		ports = append(ports, portInfo)
		klog.Infof("PodObj.TransformAgtPodReq:index is [%v],port is [%v] ", index, portInfo)
	}
	agtPodReq := agtmgr.AgentPodReq{
		PodName:  self.PodName,
		TenantId: self.TenantID,
		PodNs:    self.PodNs,
		Ports:    ports,
	}
	klog.Infof("PodObj.TransformAgtPodReq: TRANSFORM SUCC, agtPodReq is :[%v]", agtPodReq)
	return &agtPodReq, nil
}
