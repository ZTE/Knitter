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

package models

import (
	"encoding/json"
	"fmt"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

type LogicPod struct {
	PodName  string     `json:"pod_name"`
	PodNs    string     `json:"pod_ns"`
	TenantID string     `json:"tenant_id"`
	Ports    []PortInfo `json:"ports"`
}

type PortInfo struct {
	PortID       string `json:"port_id"`
	NetworkName  string `json:"network_name"`
	NetworkPlane string `json:"network_plane"`
	FixIP        string `json:"fix_ip"`
}

//struct for restful
type PodForResponse struct {
	Name       string             `json:"name"`
	PodIps     []*PortForResponse `json:"ips"`
	tenantUUID string
}

type PortForResponse struct {
	NetworkPlane      string `json:"network_plane"`
	NewtworkPlaneName string `json:"network_plane_name"`
	IPAddress         string `json:"ip_address"`
}

type EncapPodForResponse struct {
	Pod *PodForResponse `json:"pod"`
}

type EncapPodsForResponse struct {
	Pods []*PodForResponse `json:"pods"`
}

func (self *LogicPod) Save() error {
	klog.Infof("LogicPod.Save: SAVE POD START")
	if self.TenantID == "" || self.PodName == "" {
		klog.Errorf("LogicPod.Get:error is [%v]", errobj.ErrTenantsIDOrPodNameIsNil)
		return errobj.ErrTenantsIDOrPodNameIsNil
	}
	fmt.Printf("***************************")
	klog.Infof("LogicPod.Save: pod name is %s", self.PodName)
	key := dbaccessor.GetKeyOfPodName(self.TenantID, self.PodName)
	value, err := json.Marshal(self)
	if err != nil {
		klog.Errorf("LogicPod.Save: json.Marshal(%v) error, error is %v", self, err)
		return err
	}
	klog.Infof("LogicPod.Save: Logic Pod Info is %s", string(value))

	err = common.GetDataBase().SaveLeaf(key, string(value))
	if err != nil {
		klog.Errorf("LogicPod.Save: SaveLeaf(%s,%s) error, error is %v", key, value, err)
		return err
	}
	klog.Infof("LogicPod.Save: SAVE POD SUCC,pod name is %s", self.PodName)
	return nil
}

func (self *LogicPod) Delete() error {
	klog.Infof("LogicPod.Delete: DELETE POD START")
	if self.TenantID == "" || self.PodName == "" {
		klog.Errorf("LogicPod.Get:error is [%v]", errobj.ErrTenantsIDOrPodNameIsNil)
		return errobj.ErrTenantsIDOrPodNameIsNil
	}
	klog.Infof("LogicPod.Delete:pod name is %s", self.PodName)
	key := dbaccessor.GetKeyOfPodName(self.TenantID, self.PodName)
	err := common.GetDataBase().DeleteLeaf(key)
	if err != nil && !IsKeyNotFoundError(err) {
		klog.Infof("LogicPod.Delete: DeleteLeaf(%s) error, error is [%v]", key, err)
		return err
	}
	klog.Infof("LogicPod.Delete: DELETE POD SUCC, pod name is %s", self.PodName)
	return nil
}

func (self *LogicPod) Get() error {
	klog.Infof("LogicPod.Get: Get POD START")
	if self.TenantID == "" || self.PodName == "" {
		klog.Errorf("LogicPod.Get:error is [%v]", errobj.ErrTenantsIDOrPodNameIsNil)
		return errobj.ErrTenantsIDOrPodNameIsNil
	}

	klog.Infof("LogicPod.Get: Get POD, pod name is %s", self.PodName)
	key := dbaccessor.GetKeyOfPodName(self.TenantID, self.PodName)
	klog.Infof("LogicPod.Get:common.GetDataBase().Readleaf(key),key is [%s]", key)

	value, err := common.GetDataBase().ReadLeaf(key)
	if err != nil {
		klog.Errorf("LogicPod.Get: common.GetDataBase().ReadLeaf(key) error,key is [%s], error is [%v]", key, err)
		return err
	}
	klog.Infof("LogicPod.Get: json.Unmarshal([]byte(value), self) value is [%s], self is %v", value, self)

	err = json.Unmarshal([]byte(value), self)
	if err != nil {
		klog.Errorf("LogicPod.Get: json.Unmarshal([]byte(value), self) value is [%s], self is %v", value, self)
		return err
	}

	klog.Infof("LogicPod.Get: Get POD SUCC, pod  is %v", self)
	return nil
}

func (self *LogicPod) GetEncapPodForResponse() (*EncapPodForResponse, error) {
	klog.Infof("LogicPod.getEncapPodForResponse START")
	if self.TenantID == "" || self.PodName == "" {
		klog.Errorf("LogicPod.getEncapPodForResponse:error is [%v]", errobj.ErrTenantsIDOrPodNameIsNil)
		return nil, errobj.ErrTenantsIDOrPodNameIsNil
	}
	err := self.Get()
	if err != nil {
		klog.Errorf("LogicPod.getEncapPodForResponse:error is [%v]", err)
		return nil, err
	}
	podForResponse := self.TransformToPodForResponse()
	klog.Infof("LogicPod.getEncapPodForResponse: podForResponse is [%v]", podForResponse)
	encapPodForResponse := &EncapPodForResponse{
		Pod: podForResponse,
	}
	klog.Infof("LogicPod.getEncapPodForResponse SUCC")
	return encapPodForResponse, nil
}

func (self *LogicPod) TransformToPodForResponse() *PodForResponse {
	klog.Infof("LogicPod.TransformToPodForResponse START LogicPod is [%v]", self)
	podForResponse := &PodForResponse{}
	portForResponses := make([]*PortForResponse, 0)
	for _, port := range self.Ports {
		portForResponse := &PortForResponse{}
		portForResponse.IPAddress = port.FixIP
		portForResponse.NetworkPlane = port.NetworkPlane
		portForResponse.NewtworkPlaneName = port.NetworkName
		portForResponses = append(portForResponses, portForResponse)
	}
	podForResponse.Name = self.PodName
	podForResponse.tenantUUID = self.TenantID
	podForResponse.PodIps = portForResponses
	klog.Infof("LogicPod.TransformToPodForResponse SUCC podForResponse is [%v]", podForResponse)
	return podForResponse
}

//LogicPodManager

type LogicPodManager struct {
}

func (self *LogicPodManager) GetEncapPodsForResponse(tenantID string) (*EncapPodsForResponse, error) {
	klog.Infof("LogicPod.getEncapPodsForResponse START")
	if tenantID == "" {
		klog.Errorf("LogicPodManager.getEncapPodsForResponse:error is [%v]", errobj.ErrTenantsIDIsNil)
		return nil, errobj.ErrTenantsIDIsNil
	}

	logicPods, err := self.GetAll(tenantID)
	if err != nil {
		klog.Errorf("LogicPodManager.getEncapPodsForResponse: self.GetAll() eroor,error is [%v]", err)
		return nil, err
	}
	podForResponses := make([]*PodForResponse, 0)
	for _, logicPod := range logicPods {
		podForResponse := logicPod.TransformToPodForResponse()
		podForResponses = append(podForResponses, podForResponse)
	}
	encapPodsForResponse := &EncapPodsForResponse{
		Pods: podForResponses,
	}
	klog.Infof("LogicPodManager.getEncapPodsForResponse SUCC")

	return encapPodsForResponse, nil
}

func (self *LogicPodManager) GetAll(tenantID string) ([]*LogicPod, error) {
	klog.Infof("LogicPodManager.GetAll: GetAll POD START")
	if tenantID == "" {
		klog.Errorf("LogicPodManager.GetAll:error is [%v]", errobj.ErrTenantsIDIsNil)
		return nil, errobj.ErrTenantsIDIsNil
	}

	klog.Infof("LogicPodManager.GetAll:tenantID  is %s", tenantID)
	key := dbaccessor.GetKeyOfPodsTenantId(tenantID)
	nodes, err := common.GetDataBase().ReadDir(key)
	klog.Infof("LogicPodManager.GetAll: common.GetDataBase().ReadDir(key), key is [%s]", key)
	if err != nil {
		klog.Errorf("LogicPodManager.GetAll: common.GetDataBase().ReadDir(key), key is [%s],error is [%v]", key, err)
		return nil, err
	}

	logicPods := make([]*LogicPod, 0)
	for _, node := range nodes {
		logicPod := &LogicPod{}
		err := json.Unmarshal([]byte(node.Value), logicPod)
		if err != nil {
			klog.Errorf("LogicPodManager.GetAll: "+
				"json.Unmarshal([]byte(node.Value),logicPod) ,node.value is [%s],error is [%v] ", node.Value, err)
			return nil, err
		}
		logicPods = append(logicPods, logicPod)
	}

	klog.Infof("LogicPodManager.GetAll: GetAll POD SUCC, POD NUM is %v", len(logicPods))
	return logicPods, nil
}
