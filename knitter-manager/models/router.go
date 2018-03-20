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
	//log
	"encoding/json"
	"github.com/ZTE/Knitter/knitter-manager/iaas"
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

type Rt struct {
	Router     *iaasaccessor.Router
	TenantUUID string
}

type EncapRouter struct {
	Router *iaasaccessor.Router `json:"router"`
}

type EncapRouters struct {
	Routers []*iaasaccessor.Router `json:"routers"`
}

func (self *Rt) ListAll() []*iaasaccessor.Router {
	klog.Info("Now in Router.ListAll Function")
	var rts []*iaasaccessor.Router
	rtURL := dbaccessor.GetKeyOfRouterGroup(self.TenantUUID)
	nodes, errE := common.GetDataBase().ReadDir(rtURL)
	if errE != nil {
		klog.Warning("Read Router info from ETCD Error:", errE)
		return nil
	}

	for _, node := range nodes {
		rtString, err1 := common.GetDataBase().ReadLeaf(node.Key + "/self")
		if err1 != nil {
			continue
		}
		router := iaasaccessor.Router{}
		json.Unmarshal([]byte(rtString), &router)
		rt := iaasaccessor.Router{Name: router.Name, Id: router.Id, ExtNetId: router.ExtNetId}
		rts = append(rts, &rt)
	}
	return rts
}

func (self *Rt) GetFromEtcd() (*iaasaccessor.Router, error) {
	klog.Info("Now in GetNetworkById Function")
	url := dbaccessor.GetKeyOfRouterSelf(self.TenantUUID, self.Router.Id)
	rtString, err1 := common.GetDataBase().ReadLeaf(url)
	if err1 != nil {
		return nil, err1
	}
	router := iaasaccessor.Router{}
	json.Unmarshal([]byte(rtString), &router)
	rt := iaasaccessor.Router{Name: router.Name, Id: router.Id, ExtNetId: router.ExtNetId}
	klog.Info("Now out GetPortById Function")
	return &rt, nil
}

func (self *Rt) DelByID() error {
	klog.Info("Now in Router.DelById Function")
	_, err := self.GetFromEtcd()
	if err != nil {
		klog.Error("Router not found,  ERROR:", err)
		return err
	}
	err = iaas.GetIaaS(self.TenantUUID).DeleteRouter(self.Router.Id)
	if err != nil {
		return err
	}
	klog.Info("Now out Router.DelById Function")

	err = self.deleteRouterFromEtcd()
	if err != nil {
		return err
	}
	return nil
}

func (self *Rt) Create(router iaasaccessor.Router) error {
	klog.Info("Now in Router.Create Function")
	routerID, err := iaas.GetIaaS(self.TenantUUID).CreateRouter(router.Name, router.ExtNetId)
	if err != nil {
		return err
	}

	rt, err := iaas.GetIaaS(self.TenantUUID).GetRouter(routerID)
	if err != nil {
		klog.Warning("Get router error:", err)
		return err
	}

	self.Router = rt
	err = self.saveRouterToEtcd()
	if err != nil {
		iaas.GetIaaS(self.TenantUUID).DeleteRouter(routerID)
		return err
	}

	return nil
}

func (self *Rt) Update(router iaasaccessor.Router) error {
	klog.Info("Now in Router.Update Function")
	err := iaas.GetIaaS(self.TenantUUID).UpdateRouter(self.Router.Id, router.Name, router.ExtNetId)
	if err != nil {
		return err
	}

	rt, err := iaas.GetIaaS(self.TenantUUID).GetRouter(self.Router.Id)
	if err != nil {
		klog.Warning("Get router error:", err)
		return err
	}
	self.Router = rt
	err = self.saveRouterToEtcd()
	if err != nil {
		klog.Warning("Save Router Info to ETCD Err:", err)
		return err
	}
	return nil
}

func (self *Rt) getSubNetID(netID string) (string, error) {
	url := dbaccessor.GetKeyOfNetworkSelf(self.TenantUUID, netID)
	nwStr, err := common.GetDataBase().ReadLeaf(url)
	if err != nil {
		klog.Error("getSubNetId call GetDataBase().ReadLeaf ERROR:", err)
		return "", err
	}
	net := Net{}
	json.Unmarshal([]byte(nwStr), &net)
	klog.Info("getSubNetId OK:", net.Subnet.Id)
	return net.Subnet.Id, nil
}

func (self *Rt) Attach(netID string) error {
	subNetID, err := self.getSubNetID(netID)
	if err != nil {
		klog.Error("(self *Rt)Attach call getSubNetId ERROR:", err)
		return err
	}
	protID, err := iaas.GetIaaS(self.TenantUUID).AttachNetToRouter(self.Router.Id, subNetID)
	if err != nil {
		klog.Error("(self *Rt)Attach call GetIaaS().AttachNetToRouter ERROR:", err)
		return err
	}
	intf := iaasaccessor.Interface{Id: protID}
	attachPort := Intf{Interface: intf, TenantUUID: self.TenantUUID}
	err1 := attachPort.savePortToEtcd(self.Router.Id, netID)
	if err1 != nil {
		klog.Error("Attach success, but Save to ETCD ERROR:", err1)
		return err1
	}
	err2 := self.savePortToNetwork(protID, netID)
	if err2 != nil {
		klog.Error("Attach success, but Save to ETCD ERROR:", err2)
		return err2
	}
	err3 := self.savePortToRouter(protID)
	if err3 != nil {
		klog.Error("Attach success, but Save to ETCD ERROR:", err3)
		return err3
	}
	klog.Info("(self *Rt)Attach OK.")
	return nil
}

func (self *Rt) Detach(netID string) error {
	subNetID, err := self.getSubNetID(netID)
	if err != nil {
		klog.Error("(self *Rt)Detach call getSubNetId ERROR:", err)
		return err
	}
	protID, err := iaas.GetIaaS(self.TenantUUID).DetachNetFromRouter(self.Router.Id, subNetID)
	if err != nil {
		klog.Error("(self *Rt)Detach call GetIaaS().DetachNetFromRouter ERROR:", err)
		return err
	}
	err1 := self.deletePortFromNetwork(protID, netID)
	if err1 != nil {
		klog.Error("Attach success, but Save to ETCD ERROR:", err1)
		return err1
	}
	err2 := self.deletePortFromRouter(protID)
	if err2 != nil {
		klog.Error("Attach success, but Save to ETCD ERROR:", err2)
		return err2
	}
	intf := iaasaccessor.Interface{Id: protID}
	detachPort := Intf{Interface: intf, TenantUUID: self.TenantUUID}
	err3 := detachPort.deletePortFromEtcd()
	if err3 != nil {
		klog.Error("Attach success, but Save to ETCD ERROR:", err3)
		return err3
	}
	klog.Error("(self *Rt)Detach OK.")
	return nil
}

func (self *Rt) saveRouterToEtcd() error {
	key := dbaccessor.GetKeyOfRouterSelf(self.TenantUUID, self.Router.Id)
	value, _ := json.Marshal(self.Router)
	errE := common.GetDataBase().SaveLeaf(key, string(value))
	if errE != nil {
		klog.Warning(errE)
		return errE
	}

	return nil
}

func (self *Rt) deleteRouterFromEtcd() error {
	key := dbaccessor.GetKeyOfRouterSelf(self.TenantUUID, self.Router.Id)
	errKey := common.GetDataBase().DeleteLeaf(key)
	if errKey != nil {
		klog.Warning(errKey)
		return errKey
	}
	errDir := common.GetDataBase().DeleteDir(key)
	if errDir != nil {
		klog.Warning(errDir)
		return errDir
	}

	return nil
}

func (self *Rt) savePortToNetwork(ptID, nwID string) error {
	key := dbaccessor.GetKeyOfInterfaceInNetwork(self.TenantUUID, nwID, ptID)
	value := dbaccessor.GetKeyOfInterfaceSelf(self.TenantUUID, ptID)
	errE := common.GetDataBase().SaveLeaf(key, value)
	if errE != nil {
		klog.Warning(errE)
		return errE
	}

	return nil
}

func (self *Rt) deletePortFromNetwork(ptID, nwID string) error {
	key := dbaccessor.GetKeyOfInterfaceInNetwork(self.TenantUUID, nwID, ptID)
	errKey := common.GetDataBase().DeleteLeaf(key)
	if errKey != nil {
		klog.Warning(errKey)
		return errKey
	}
	return nil
}

func (self *Rt) savePortToRouter(ptID string) error {
	key := dbaccessor.GetKeyOfInterfaceInRouter(self.TenantUUID, self.Router.Id, ptID)
	value := dbaccessor.GetKeyOfInterfaceSelf(self.TenantUUID, ptID)
	errE := common.GetDataBase().SaveLeaf(key, value)
	if errE != nil {
		klog.Warning(errE)
		return errE
	}

	return nil
}

func (self *Rt) deletePortFromRouter(ptID string) error {
	key := dbaccessor.GetKeyOfInterfaceInRouter(self.TenantUUID, self.Router.Id, ptID)
	errKey := common.GetDataBase().DeleteLeaf(key)
	if errKey != nil {
		klog.Warning(errKey)
		return errKey
	}

	return nil
}
