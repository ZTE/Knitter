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
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

type PortObjRole struct {
}

func (this PortObjRole) Create(value string) (*portobj.PortObj, error) {
	agtCtx := cni.GetGlobalContext()
	portJSON, err := agtCtx.DB.ReadLeaf(value)
	if err != nil {
		klog.Errorf("DetachPortFromBrintAction:agtCtx.DB.ReadLeaf err: %v", err)
		return nil, errobj.ErrContinue
	}
	port := iaasaccessor.Interface{}
	err = json.Unmarshal([]byte(portJSON), &port)
	if err != nil {
		klog.Errorf("DetachPortFromBrintAction:json.Unmarshal err: %v", err)
		return nil, errobj.ErrContinue
	}

	portObj := portobj.PortObj{}
	build(&portObj, &port)
	return &portObj, nil
}

func build(portObj *portobj.PortObj, port *iaasaccessor.Interface) {
	portObj.EagerAttr.NetworkName = port.NetPlaneName
	portObj.EagerAttr.NetworkPlane = port.NetPlane
	portObj.EagerAttr.PortName = port.Name
	portObj.EagerAttr.VnicType = port.NicType
	portObj.EagerAttr.PodNs = port.PodNs
	portObj.EagerAttr.PodName = port.PodName
	portObj.EagerAttr.Accelerate = port.Accelerate
	portObj.LazyAttr.ID = port.Id
	portObj.LazyAttr.TenantID = port.PodNs
	portObj.LazyAttr.NetAttr.ID = port.NetworkId
	portObj.LazyAttr.BusInfos = port.BusInfos
	portObj.LazyAttr.MacAddress = port.MacAddress
	portObj.LazyAttr.OrgDriver = port.OrgDriver
}

func (this PortObjRole) Deserialize(portJSON string) (*portobj.PortObj, error) {
	port := iaasaccessor.Interface{}
	err := json.Unmarshal([]byte(portJSON), &port)
	if err != nil {
		klog.Errorf("Deserialize: json.Unmarshal err: %v", err)
		return nil, err
	}

	portObj := portobj.PortObj{}
	build(&portObj, &port)
	klog.Infof("Deserialize: build portObj:[%v] result port: %v", portObj, port)
	return &portObj, nil
}
