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

package context

import (
	"encoding/json"
	"errors"
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/port-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
	"github.com/rackspace/gophercloud/openstack/networking/v2/ports"
	"strings"
)

type ReusePortInfoAction struct {
}

func (c *ReusePortInfoAction) Exec(transInfo *transdsl.TransInfo) error {
	klog.Infof("***ReusePortInfoAction:Exec begin***")
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	knitterObj := knitterInfo.KnitterObj
	agtCtx := cni.GetGlobalContext()
	portsInEtcd, err := getPortListOfPod(agtCtx, knitterObj.CniParam.TenantID, knitterObj.CniParam.PodNs, knitterObj.CniParam.PodName)
	if err != nil || len(portsInEtcd) == 0 {
		return err
	}
	for i, transInfoPort := range knitterInfo.podObj.PortObjs {
		if transInfoPort.EagerAttr.Accelerate == "true" &&
			infra.IsCTNetPlane(transInfoPort.EagerAttr.NetworkPlane) &&
			agtCtx.RunMode != "overlay" {
			continue
		}
		portInfoInEtcd, err := findPortInfoInEtcd(portsInEtcd, transInfoPort.EagerAttr.PortName)
		if err != nil {
			klog.Errorf("Do not find port[%v]", transInfoPort.EagerAttr.PortName)
			return err
		}
		buildPortObj(knitterInfo.podObj.PortObjs[i], portInfoInEtcd)
	}
	cleanPortsRecordInETCD(agtCtx, knitterObj.CniParam)
	cleanPortsRecordInLocalDB(agtCtx, knitterObj.CniParam)
	klog.Infof("***ReusePortInfoAction:Exec end***")
	return nil
}

func (c *ReusePortInfoAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***ReusePortInfoAction:RollBack begin***")
	klog.Infof("***ReusePortInfoAction:RollBack end***")
}

var getPortListOfPod = func(agtCtx *cni.AgentContext, tenantID string, podNs string, podName string) ([]*portobj.LogicPortObj, error) {
	keyPortsInPod := dbaccessor.GetKeyOfLogicPortsInPod(tenantID, podNs, podName)
	portsLintInEtcd, errLint := agtCtx.RemoteDB.ReadDir(keyPortsInPod)
	if errLint != nil || len(portsLintInEtcd) == 0 {
		klog.Errorf("getPortListOfPod[podns:%v;podname:%v] Err: portsLintInEtcd is none",
			podNs, podName)
		return nil, errors.New("port in etcd is none")
	}
	portObjsInEtcd := make([]*portobj.LogicPortObj, 0)
	for _, node := range portsLintInEtcd {
		var portObj portobj.LogicPortObj
		errUnmarshal := json.Unmarshal([]byte(node.Value), &portObj)
		if errUnmarshal != nil {
			klog.Errorf("Unmarshal for portInEtcd[%v] error: %v", node.Value, errUnmarshal)
			return nil, errUnmarshal
		}
		portObjsInEtcd = append(portObjsInEtcd, &portObj)
	}
	return portObjsInEtcd, nil
}

var findPortInfoInEtcd = func(portsInEtcd []*portobj.LogicPortObj, portName string) (*portobj.LogicPortObj, error) {
	for _, portObjInEtcd := range portsInEtcd {
		if portName == portObjInEtcd.Name {
			return portObjInEtcd, nil
		}
	}
	return nil, errors.New("Do not find port")
}

var buildPortObj = func(portObj *portobj.PortObj, logicPort *portobj.LogicPortObj) error {
	portObj.LazyAttr.ID = logicPort.ID
	portObj.LazyAttr.Name = portObj.EagerAttr.PortName
	portObj.LazyAttr.MacAddress = logicPort.MacAddress
	portObj.LazyAttr.FixedIps = []ports.IP{
		{SubnetID: logicPort.SubNetID, IPAddress: logicPort.IP}}
	portObj.LazyAttr.TenantID = logicPort.TenantID
	portObj.LazyAttr.Cidr = logicPort.Cidr
	portObj.LazyAttr.GatewayIP = logicPort.GatewayIP
	return nil
}

var cleanPortsRecordInETCD = func(agtCtx *cni.AgentContext, cniobj *cni.CniParam) error {
	keyNodes := dbaccessor.GetKeyOfInterfaceGroupInPod(cniobj.TenantID, cniobj.PodNs, cniobj.PodName)
	nodes, err := agtCtx.RemoteDB.ReadDir(keyNodes)
	if err != nil {
		klog.Errorf("ReadDir(%v) FAILED, errror: %v", keyNodes, err)
		return err
	}
	for _, node := range nodes {
		interLongID := strings.TrimPrefix(node.Key, keyNodes+"/")
		portJSON, err := agtCtx.RemoteDB.ReadLeaf(node.Value)
		if err != nil {
			klog.Warningf("Read port leaf err: %v", err)
			continue
		}
		port := iaasaccessor.Interface{}
		err = json.Unmarshal([]byte(portJSON), &port)
		if err != nil {
			klog.Errorf("json.Unmarshal err: %v", err)
			continue
		}
		keyNet := dbaccessor.GetKeyOfInterfaceInNetwork(port.TenantID, port.NetworkId, interLongID)
		agtCtx.RemoteDB.DeleteLeaf(keyNet)
		keyInter := strings.TrimSuffix(node.Value, "/self")
		agtCtx.RemoteDB.DeleteDir(keyInter)
		keyPod := dbaccessor.GetKeyOfPod(port.TenantID, port.PodNs, port.PodName)
		agtCtx.RemoteDB.DeleteDir(keyPod)
	}
	return nil
}

var cleanPortsRecordInLocalDB = func(agtCtx *cni.AgentContext, cniobj *cni.CniParam) error {
	keyNodes := dbaccessor.GetKeyOfInterfaceGroupInPod(cniobj.TenantID, cniobj.PodNs, cniobj.PodName)
	nodes, err := agtCtx.DB.ReadDir(keyNodes)
	if err != nil {
		klog.Errorf("ReadDir(%v) FAILED, errror: %v", keyNodes, err)
		return err
	}
	for _, node := range nodes {
		interLongID := strings.TrimPrefix(node.Key, keyNodes+"/")
		portJSON, err := agtCtx.DB.ReadLeaf(node.Value)
		if err != nil {
			klog.Warningf("Read port leaf err: %v", err)
			continue
		}
		port := iaasaccessor.Interface{}
		err = json.Unmarshal([]byte(portJSON), &port)
		if err != nil {
			klog.Errorf("json.Unmarshal err: %v", err)
			continue
		}
		keyNet := dbaccessor.GetKeyOfInterfaceInNetwork(port.TenantID, port.NetworkId, interLongID)
		agtCtx.DB.DeleteLeaf(keyNet)
		keyInter := strings.TrimSuffix(node.Value, "/self")
		agtCtx.DB.DeleteDir(keyInter)
		keyPod := dbaccessor.GetKeyOfPod(port.TenantID, port.PodNs, port.PodName)
		agtCtx.DB.DeleteDir(keyPod)
	}
	return nil
}
