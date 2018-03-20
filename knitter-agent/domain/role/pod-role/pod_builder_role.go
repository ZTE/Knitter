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

package podrole

import (
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/monitor"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/pkg/klog"
)

type PodBuilderRole struct {
	PodID    string
	PodType  string
	PortObjs []*portobj.PortObj
}

func (this *PodBuilderRole) Transform(podNs, podName string, pod *monitor.Pod) error {

	this.PodID = pod.PodID
	var err error

	this.PortObjs, err = this.AnalyzeV2PodNetTemplate(podNs, podName, pod.Ports)
	if err != nil {
		klog.Info("Create: AnalyzeV2PodNetTemplate error:", err)
		return err
	}
	this.PodType = this.GetPodType(this.PortObjs)
	klog.Infof("AnalyzeK8sRspd: podType is: %v", this.PodType)

	return nil
}

func (this *PodBuilderRole) AnalyzeV2PodNetTemplate(podNs, podName string, Ports []*monitor.Port) ([]*portobj.PortObj,
	error) {
	if len(Ports) == 0 {
		err := errors.New("get port config error")

		klog.Errorf("Get port config error! %v", err)
		return nil, err
	}
	var portObjs []*portobj.PortObj
	for _, portObj := range Ports {
		port, err := portobj.CreatePortObj(podNs, podName, portObj)
		if err != nil {
			return nil, err
		}
		klog.Infof("PodBuilderRole.AnalyzeV2PodNetTemplate: portobj is [%v]", port)
		portObjs = append(portObjs, port)
	}
	return portObjs, nil
}

func (this *PodBuilderRole) GetPodType(portList []*portobj.PortObj) string {
	var podType string
	var (
		ctlPlaneExist = false
		medPlaneExist = false
		ctlPlaneAcce  = false
		medPlaneAcce  = false
	)
	for _, port := range portList {
		if port.BuildPortRole.NetworkPlane == "control" {
			ctlPlaneExist = true
			if port.BuildPortRole.Accelerate == "true" {
				ctlPlaneAcce = true
			}
		}
		if port.BuildPortRole.NetworkPlane == "media" {
			medPlaneExist = true
			if port.BuildPortRole.Accelerate == "true" {
				medPlaneAcce = true
			}
		}
	}

	if ctlPlaneExist && medPlaneExist && ctlPlaneAcce && medPlaneAcce {
		podType = "ct"
	} else if ctlPlaneExist && medPlaneExist && !ctlPlaneAcce && !medPlaneAcce {
		podType = "ct_minus"
	} else {
		podType = "it"
	}

	return podType
}
