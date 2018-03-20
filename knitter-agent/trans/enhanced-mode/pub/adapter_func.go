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

package modelsext

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
)

var BuildIaasPorts = func(portObjs []*portobj.PortObj) []*iaasaccessor.Interface {
	iaasPorts := make([]*iaasaccessor.Interface, 0)
	for _, portObj := range portObjs {
		iaasPorts = append(iaasPorts, BuildIaasPort(portObj))
	}
	return iaasPorts
}

var BuildIaasPort = func(portObj *portobj.PortObj) *iaasaccessor.Interface {
	iaasPort := iaasaccessor.Interface{}
	iaasPort.NetPlaneName = portObj.EagerAttr.NetworkName
	iaasPort.NetPlane = portObj.EagerAttr.NetworkPlane
	iaasPort.Name = portObj.EagerAttr.PortName
	iaasPort.NicType = portObj.EagerAttr.VnicType
	iaasPort.PodNs = portObj.EagerAttr.PodNs
	iaasPort.TenantID = portObj.EagerAttr.PodNs
	iaasPort.PodName = portObj.EagerAttr.PodName
	iaasPort.Accelerate = portObj.EagerAttr.Accelerate
	iaasPort.Id = portObj.LazyAttr.ID
	iaasPort.NetworkId = portObj.LazyAttr.NetAttr.ID
	iaasPort.BusInfos = portObj.LazyAttr.BusInfos
	iaasPort.OrgDriver = portObj.LazyAttr.OrgDriver
	return &iaasPort
}
