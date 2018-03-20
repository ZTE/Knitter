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
	"github.com/ZTE/Knitter/knitter-manager/public"
	"github.com/ZTE/Knitter/pkg/db-accessor"
	"github.com/ZTE/Knitter/pkg/iaas-accessor"
	"github.com/ZTE/Knitter/pkg/klog"
)

type Intf struct {
	Interface  iaasaccessor.Interface
	TenantUUID string
	DcID       string
}

func (self *Intf) savePortToEtcd(rtID, nwID string) error {
	klog.Info("Save Port[", self.Interface.Id, "] to network[",
		nwID, "] and router[", rtID, "] now.")
	key := dbaccessor.GetKeyOfInterfaceSelf(self.TenantUUID, self.Interface.Id)
	value, _ := json.Marshal(self.Interface)
	errPt := common.GetDataBase().SaveLeaf(key, string(value))
	if errPt != nil {
		klog.Error(errPt)
		return errPt
	}
	if rtID != "" {
		keyRouter := dbaccessor.GetKeyOfRouterByInterface(self.TenantUUID, self.Interface.Id)
		valueRouter := dbaccessor.GetKeyOfRouterSelf(self.TenantUUID, rtID)
		errRt := common.GetDataBase().SaveLeaf(keyRouter, valueRouter)
		if errRt != nil {
			klog.Warning(errRt)
			return errRt
		}
	}
	keyNetwork := dbaccessor.GetKeyOfNetworkByInterface(self.TenantUUID, self.Interface.Id)
	valueNetwork := dbaccessor.GetKeyOfNetworkSelf(self.TenantUUID, nwID)
	errNw := common.GetDataBase().SaveLeaf(keyNetwork, valueNetwork)
	if errNw != nil {
		klog.Error(errNw)
		return errNw
	}

	keyIfInDc := dbaccessor.GetKeyOfInterfaceInDc(self.DcID, self.Interface.Id)
	err := common.GetDataBase().SaveLeaf(keyIfInDc, nwID)
	if err != nil {
		klog.Errorf("savePortToEtcd: save key:%s, value:%s error: %v", keyIfInDc, nwID, err)
		return err
	}

	return nil
}

func (self *Intf) deletePortFromEtcd() error {
	key := dbaccessor.GetKeyOfInterface(self.TenantUUID, self.Interface.Id)
	errDir := common.GetDataBase().DeleteDir(key)
	if errDir != nil {
		klog.Warning(errDir)
	}
	return errDir
}
