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
	"github.com/ZTE/Knitter/knitter-agent/domain/object/bridge-obj"
	"github.com/ZTE/Knitter/knitter-agent/infra/concurrency_ctrl"
	"github.com/ZTE/Knitter/pkg/klog"
	"github.com/ZTE/Knitter/pkg/trans-dsl"
)

type IsTenantNetworkNotExist struct {
}

func (this *IsTenantNetworkNotExist) Ok(transInfo *transdsl.TransInfo) bool {
	portObj := transInfo.AppInfo.(*KnitterInfo).podObj.PortObjs[transInfo.RepeatIdx]
	networkID := portObj.LazyAttr.NetAttr.ID

	concurrencyctrl.ChanMapLock.Lock()
	value, ok := concurrencyctrl.ChanMap[networkID]
	if ok {
		transInfo.AppInfo.(*KnitterInfo).Chan = value
	} else {
		transInfo.AppInfo.(*KnitterInfo).Chan = make(chan int, 1)
		transInfo.AppInfo.(*KnitterInfo).Chan <- 1
		concurrencyctrl.ChanMap[networkID] = transInfo.AppInfo.(*KnitterInfo).Chan
	}
	concurrencyctrl.ChanMapLock.Unlock()

	<-transInfo.AppInfo.(*KnitterInfo).Chan
	transInfo.AppInfo.(*KnitterInfo).ChanFlag = true
	bridgeObj := bridgeobj.GetBridgeObjSingleton()
	vlanID, err := bridgeObj.BrintRole.GetVlanID(networkID)
	if err != nil {
		klog.Infof("***IsTenantNetworkNotExist: true***")
		return true
	}
	portObj.LazyAttr.VlanID = vlanID
	klog.Infof("***IsTenantNetworkNotExist: false***")
	return false
}
