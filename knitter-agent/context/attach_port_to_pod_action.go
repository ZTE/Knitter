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
	"github.com/ZTE/Knitter/knitter-agent/domain/cni"
	"github.com/ZTE/Knitter/knitter-agent/domain/object/bridge-obj"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/pod-role"
	"github.com/ZTE/Knitter/knitter-agent/err-obj"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type AttachPortToPodAction struct {
}

func (this *AttachPortToPodAction) Exec(transInfo *transdsl.TransInfo) (err error) {
	klog.Infof("***AttachPortToPodAction:Exec begin***")
	defer func() {
		if p := recover(); p != nil {
			RecoverErr(p, &err, "AttachPortToPodAction")
		}
		AppendActionName(&err, "AttachPortToPodAction")
	}()
	bridegeObj := bridgeobj.GetBridgeObjSingleton()
	knitterInfo := transInfo.AppInfo.(*KnitterInfo)
	portObj := knitterInfo.podObj.PortObjs[transInfo.RepeatIdx]
	if portObj.LazyAttr.NetAttr.ID == bridegeObj.BrintRole.GetDefaultGwNetworkID() {
		knitterInfo.mgrPort.IsDefaultGateway = true
		knitterInfo.mgrPort.GatewayIP = bridegeObj.BrintRole.GetDefaultGwIP()
	}

	nic, err := knitterInfo.podObj.PortRole.Attach(knitterInfo.KnitterObj.CniParam,
		knitterInfo.mgrPort, knitterInfo.vethPair.VethNameOfPod)
	if err != nil {
		klog.Errorf("AttachPortToPodAction:Exec:transInfo.podObj.PortRole.Attach err: %v", err)
		return err
	}
	err = bridegeObj.BrintRole.IncRefCount(knitterInfo.mgrPort.NetworkID,
		knitterInfo.podObj.PodNs, knitterInfo.podObj.PodName)
	if err != nil {
		klog.Errorf("AttachPortToPodAction:Exec:bridegeObj.BrintRole.IncRefCount error: %v", err)
		return errobj.ErrIncRefcountFailed
	}

	knitterInfo.Chan <- 1
	knitterInfo.ChanFlag = false

	err = knitterInfo.podObj.PortRole.StoreToDB(cni.GetGlobalContext().DB, knitterInfo.mgrPort,
		portObj, nic.BusInfo)
	if err != nil {
		klog.Errorf("AttachPortToPodAction:Exec:transInfo.podObj.PortRole.StoreToDB err: %v", err)
		return errobj.ErrStorePod2DBFailed
	}

	err = knitterInfo.podObj.PortRole.StoreToDB(cni.GetGlobalContext().RemoteDB, knitterInfo.mgrPort,
		portObj, nic.BusInfo)
	if err != nil {
		klog.Errorf("AttachPortToPodAction:Exec:transInfo.podObj.PortRole.StoreToDB(remote) err: %v", err)
		return errobj.ErrStorePod2etcdFailed
	}

	knitterInfo.Nics = append(knitterInfo.Nics, *nic)
	klog.Infof("***AttachPortToPodAction:Exec end***")
	return nil
}

func (this *AttachPortToPodAction) RollBack(transInfo *transdsl.TransInfo) {
	klog.Infof("***AttachPortToPodAction:RollBack begin***")
	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	portObj := transInfo.AppInfo.(*KnitterInfo).podObj.PortObjs[transInfo.RepeatIdx]

	portObj.LazyAttr.ID = transInfo.AppInfo.(*KnitterInfo).mgrPort.ID
	knitterObj := transInfo.AppInfo.(*KnitterInfo).KnitterObj
	portObj.LazyAttr.TenantID = knitterObj.CniParam.TenantID
	portObj.EagerAttr.PodNs = knitterObj.CniParam.PodNs
	portObj.EagerAttr.PodName = knitterObj.CniParam.PodName

	err := podrole.DeleteFromDB(cni.GetGlobalContext().DB, portObj)
	if err != nil {
		klog.Errorf("AttachPortToPodAction:podrole.DeleteFromDB return err:%v", err)
	}
	err = podrole.DeleteFromDB(cni.GetGlobalContext().RemoteDB, portObj)
	if err != nil {
		klog.Errorf("AttachPortToPodAction:podrole.DeleteFromDB(remote) return err:%v", err)
	}

	err = bridgeObj.BrintRole.DecRefCount(portObj.LazyAttr.NetAttr.ID, portObj.EagerAttr.PodNs, portObj.EagerAttr.PodName)
	if err != nil {
		klog.Errorf("AttachPortToPodAction:bridgeObj.BrintRole.DecRefCount networkId: %v, err: %v", portObj.LazyAttr.NetAttr.ID, err)
	}
	klog.Infof("***AttachPortToPodAction:RollBack end***")
}
