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

package brintsubrole

import (
	"github.com/ZTE/Knitter/knitter-agent/domain/bind"
	"github.com/ZTE/Knitter/knitter-agent/domain/const-value"
	"github.com/ZTE/Knitter/knitter-agent/domain/manager"
	"github.com/ZTE/Knitter/knitter-agent/domain/ovs"
	"github.com/ZTE/Knitter/knitter-agent/domain/role/port-role"
	"github.com/ZTE/Knitter/pkg/klog"
	"time"
)

const (
	DefaultVethOpsRetryTimes  = 5
	DefaultVethOpsRetryIntval = 30
)

type PortRole struct {
}

func (this PortRole) AttachPort(vethNameOfBridge, vlanID string) error {
	err := bind.AddVethToOvsWithVlanID(constvalue.OvsBrint, vethNameOfBridge, vlanID)
	if err != nil {
		klog.Errorf("PortRole:AttachPort AttachPort:bind.AddVethToOvsWithVlanId error: %v", err)
		return err
	}

	for idx := 0; idx < DefaultVethOpsRetryTimes; idx++ {
		err = bind.ActiveVethPort(vethNameOfBridge)
		if err == nil {
			return nil
		}
		klog.Infof("PortRole:AttachPort:bind.ActiveVethPort %d time error: %v, total %d times",
			idx, err, DefaultVethOpsRetryTimes)
		if idx+1 < DefaultVethOpsRetryTimes {
			time.Sleep(DefaultVethOpsRetryIntval * time.Millisecond)
		}
	}
	klog.Errorf("PortRole:AttachPort:bind.ActiveVethPort error: %v", err)
	this.DetachPort(vethNameOfBridge)
	return err
}

func (this PortRole) DetachPort(vethNameOfBridge string) error {
	err := bind.DelVethFromOvs(constvalue.OvsBrint, vethNameOfBridge)
	if err != nil {
		klog.Errorf("detachPortFromBrint vethBr: %s error: %v", vethNameOfBridge, err)
	}
	return err
}

func (this PortRole) DelPort(portID, tenantID string, vethPair *ovs.VethPair, mgr *manager.ManagerClient) {
	err := portrole.GetPortTableSingleton().Delete(portID)
	if err != nil {
		klog.Errorf("PortRole:portrole.GetPortTableSingleton().Delete err: %v", err)
	}
	err = ovs.DeleteVethPair(*vethPair)
	if err != nil {
		klog.Errorf("PortRole:ovs.DeleteVethPair vethPair[%v] error: %v", vethPair, err)
	}

	mgr.DeleteNeutronPort(portID, tenantID)
	if err != nil {
		klog.Errorf("PortRole:cniObj.Manager.DeleteNeutronPort port[id: %s] error! %v", portID, err)
	}

}
